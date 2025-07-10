package distSyncTask

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/public/redisKeys"
)

var (
	ErrRoutedSyncTaskDecodeFailed = errors.New("decode routed sync task failed")
	ErrRouteCmdNotFound           = errors.New("route cmd not found")
)

func NewSyncTaskSchedMux(cfg *SyncTaskSchedMuxConf) (*SyncTaskSchedMux, error) {
	sched, err := NewSyncTaskSched(&SyncTaskSchedConf{
		Name:                              cfg.Name,
		SyncQueueSize:                     cfg.SyncQueueSize,
		ConcurSyncWorkerNum:               cfg.ConcurSyncWorkerNum,
		BucketSyncConcurMax:               cfg.BucketSyncConcurMax,
		ListenSyncTaskRegFromSlaveEvtName: cfg.ListenSyncTaskRegFromSlaveEvtName,
		OnListenSyncTaskRegFromSlaveHdl:   cfg.OnListenSyncTaskRegFromSlaveHdl,
		TaskAppIdSetRedisKey:              cfg.TaskAppIdSetRedisKey,
		GenTaskSetRedisKeyHdl:             cfg.GenTaskSetRedisKeyHdl,
		ReloadRedoLogIntervalSec:          cfg.ReloadRedoLogIntervalSec,
		DecodeSyncTaskFromString:          DecodeRoutedSyncTaskWrapperFunc(cfg),
		OnExecSync:                        OnExecRoutedSyncTaskFunc(cfg),
	})
	if err != nil {
		logs.Errorf("err:%v", err)
		return nil, err
	}
	return &SyncTaskSchedMux{
		sched: sched,
	}, nil
}

// SyncTaskSchedMux 分布式消息任务同步调度路由器
type SyncTaskSchedMux struct {
	sched *SyncTaskSched
}

// PreWriteRedoLog 预写redo日志
func (s *SyncTaskSchedMux) PreWriteRedoLog(appId int32, syncTask RoutedSyncTask) error {
	err := s.sched.PreWriteRedoLog(appId, &RoutedSyncTaskWrapper{
		RouteCmd:       syncTask.GetRouteCmd(),
		RoutedSyncTask: syncTask,
	})
	if err != nil {
		logs.Errorf("err:%v", err)
		return err
	}
	return nil
}

// RegSyncTask 注册同步任务
func (s *SyncTaskSchedMux) RegSyncTask(appId int32, syncTask RoutedSyncTask) error {
	err := s.sched.RegSyncTask(appId, &RoutedSyncTaskWrapper{
		RouteCmd:       syncTask.GetRouteCmd(),
		RoutedSyncTask: syncTask,
	})
	if err != nil {
		logs.Errorf("err:%v", err)
		return err
	}
	return nil
}

// TxnRegSyncTask 便捷地开启一个注册同步任务的事务（包含预写重做日志，注册任务步骤）
func (s *SyncTaskSchedMux) TxnRegSyncTask(appId int32, syncTask RoutedSyncTask, prepareTxnFn func() error) error {
	err := s.sched.TxnRegSyncTask(appId, &RoutedSyncTaskWrapper{
		RouteCmd:       syncTask.GetRouteCmd(),
		RoutedSyncTask: syncTask,
	}, prepareTxnFn)
	if err != nil {
		logs.Errorf("err:%v", err)
		return err
	}
	return nil
}

func OnExecRoutedSyncTaskFunc(cfg *SyncTaskSchedMuxConf) func(task SyncTask) error {
	return func(task SyncTask) error {
		routedTaskWrapper := task.(*RoutedSyncTaskWrapper)

		route, ok := cfg.GetRoute(routedTaskWrapper.RouteCmd)
		if !ok {
			return ErrRouteCmdNotFound
		}

		if err := route.OnExecRoutedSyncTaskFunc(routedTaskWrapper.RoutedSyncTask); err != nil {
			logs.Errorf("err:%v", err)
			return err
		}

		return nil
	}
}

type SyncTaskSchedMuxConf struct {
	Name                                     string
	SyncQueueSize                            int
	ConcurSyncWorkerNum, BucketSyncConcurMax uint32
	ListenSyncTaskRegFromSlaveEvtName        string
	OnListenSyncTaskRegFromSlaveHdl          func([]byte)
	TaskAppIdSetRedisKey                     *redisKeys.RedisKeys                   // 把任务按分类应用id存储，方便按应用id查询挤压和管理任务，appId并非特指包网id,可以是任何业务归属应用id
	GenTaskSetRedisKeyHdl                    func(appId int32) *redisKeys.RedisKeys // 把任务按分类应用id存储，方便按应用id查询挤压和管理任务，appId并非特指包网id,可以是任何业务归属应用id
	ReloadRedoLogIntervalSec                 uint32
	Routes                                   map[RouteCmd]*Route
	mu                                       sync.RWMutex
}

func (c *SyncTaskSchedMuxConf) AddRoutes(routes map[RouteCmd]*Route) error {
	for cmd, route := range routes {
		if err := c.AddRoute(cmd, route); err != nil {
			return err
		}
	}
	return nil
}

func (c *SyncTaskSchedMuxConf) AddRoute(cmd RouteCmd, route *Route) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Routes == nil {
		c.Routes = map[RouteCmd]*Route{}
	}
	if _, ok := c.Routes[cmd]; ok {
		return fmt.Errorf("route cmd:%d duplicated", cmd)
	}
	c.Routes[cmd] = route
	return nil
}

func (c *SyncTaskSchedMuxConf) GetRoute(cmd RouteCmd) (*Route, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.Routes == nil {
		return nil, false
	}
	route, ok := c.Routes[cmd]
	return route, ok
}

func DecodeRoutedSyncTaskWrapperFunc(cfg *SyncTaskSchedMuxConf) func(s string) (SyncTask, error) {
	return func(s string) (SyncTask, error) {
		routedTaskComponent := strings.SplitN(s, "@", 2)
		if len(routedTaskComponent) != 2 {
			logs.Errorf("err:%v", ErrRoutedSyncTaskDecodeFailed)
			return nil, ErrRoutedSyncTaskDecodeFailed
		}

		routeCmdStr := routedTaskComponent[0]
		routeCmdInt64, err := strconv.ParseInt(routeCmdStr, 10, 32)
		if err != nil {
			logs.Errorf("err:%v", err)
			return nil, err
		}

		routeCmd := RouteCmd(routeCmdInt64)

		route, ok := cfg.GetRoute(routeCmd)
		if !ok {
			logs.Errorf("err:%v", ErrRoutedSyncTaskDecodeFailed)
			return nil, ErrRoutedSyncTaskDecodeFailed
		}

		routedSyncTask, err := route.DecodeRoutedSyncTaskFunc(routedTaskComponent[1])
		if err != nil {
			logs.Errorf("err:%v", err)
			return nil, err
		}

		return &RoutedSyncTaskWrapper{
			RouteCmd:       routeCmd,
			RoutedSyncTask: routedSyncTask,
		}, nil
	}
}

var _ SyncTask = (*RoutedSyncTaskWrapper)(nil)

type Route struct {
	DecodeRoutedSyncTaskFunc func(s string) (RoutedSyncTask, error)
	OnExecRoutedSyncTaskFunc func(task RoutedSyncTask) error
}

type RouteCmd int32

type RoutedSyncTaskWrapper struct {
	RouteCmd       RouteCmd
	RoutedSyncTask RoutedSyncTask
}

func (w *RoutedSyncTaskWrapper) GetBucketId() int64 {
	return w.RoutedSyncTask.GetBucketId()
}

func (w *RoutedSyncTaskWrapper) String() string {
	return fmt.Sprintf("%d@%s", w.RouteCmd, w.RoutedSyncTask.String())
}

type RoutedSyncTask interface {
	SyncTask
	GetRouteCmd() RouteCmd
}
