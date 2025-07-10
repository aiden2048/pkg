// Package distSyncTask 分布式消息任务同步调度器
package distSyncTask

import (
	"fmt"
	"sync"
	"time"

	"github.com/aiden2048/pkg/frame"
	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/frame/runtime"
	"github.com/aiden2048/pkg/public/bucketQueue"
	"github.com/aiden2048/pkg/public/elect"
	"github.com/aiden2048/pkg/public/redisDeal"
	"github.com/aiden2048/pkg/public/redisKeys"
	uuid "github.com/satori/go.uuid"
)

type SyncTaskSchedConf struct {
	Name                                     string
	SyncQueueSize                            int
	ConcurSyncWorkerNum, BucketSyncConcurMax uint32
	ListenSyncTaskRegFromSlaveEvtName        string
	OnListenSyncTaskRegFromSlaveHdl          func([]byte)
	SendSyncTaskRegEvtForMasterHdl           func(task SyncTask)
	TaskAppIdSetRedisKey                     *redisKeys.RedisKeys                   // 把任务按分类应用id存储，方便按应用id查询挤压和管理任务，appId并非特指包网id,可以是任何业务归属应用id
	GenTaskSetRedisKeyHdl                    func(appId int32) *redisKeys.RedisKeys // 把任务按分类应用id存储，方便按应用id查询挤压和管理任务，appId并非特指包网id,可以是任何业务归属应用id
	DecodeSyncTaskFromString                 func(string) (SyncTask, error)
	OnExecSync                               func(task SyncTask) error
	ReloadRedoLogIntervalSec                 uint32
}

func NewSyncTaskSched(cfg *SyncTaskSchedConf) (*SyncTaskSched, error) {
	e, err := elect.NewRedisElect(cfg.Name, uuid.NewV4().String())
	if err != nil {
		logs.Errorf("err:%v", err)
		return nil, err
	}

	e.Run()

	if cfg.BucketSyncConcurMax == 0 {
		cfg.BucketSyncConcurMax = 1
	}
	// 总共并行线程concurSyncWorkerNum个，但是根据根据bucket进行资源限制调度，一个bucket同时只存在bucketSyncConcurMax个同步任务运行,
	// 队列会根据bucket交替执行任务，保证调度公平，不会有bucket进入饥饿等待状态
	queue := bucketQueue.NewBucketQueueWithBucketConcur(cfg.Name, cfg.ConcurSyncWorkerNum, cfg.BucketSyncConcurMax)
	if cfg.SyncQueueSize > 0 {
		queue.SetSize(cfg.SyncQueueSize)
	}
	sched := &SyncTaskSched{
		queue:              queue,
		elect:              e,
		cfg:                cfg,
		mapAppIdToTaskStrs: make(map[int32]map[string]interface{}),
	}

	// 任务进程同步
	runtime.Go(func() {
		queue.Run()
	})

	// 监听从节点广播的同步任务注册，只有master节点负责同步
	sched.initSyncTaskFromSlaveListener()
	// 初始化重试任务加载
	sched.initRedoSyncTaskLoader()
	// 初始化无效任务清理
	sched.initInvalidTxnLogClearer()

	return sched, nil
}

type SyncTaskSched struct {
	queue               *bucketQueue.BucketQueue
	elect               *elect.RedisElect
	cfg                 *SyncTaskSchedConf
	lastLoadRedoLogTime time.Time
	mapAppIdToTaskStrs  map[int32]map[string]interface{}
	mu                  sync.RWMutex
}

func (s *SyncTaskSched) IsMaster() bool {
	return s.elect.IsMaster()
}

func (s *SyncTaskSched) initSyncTaskFromSlaveListener() {
	if s.cfg.OnListenSyncTaskRegFromSlaveHdl != nil && s.cfg.ListenSyncTaskRegFromSlaveEvtName != "" {
		frame.ListenConfig(s.cfg.ListenSyncTaskRegFromSlaveEvtName, s.cfg.OnListenSyncTaskRegFromSlaveHdl)
	}
}

// 初始化无效日志清理器
func (s *SyncTaskSched) initInvalidTxnLogClearer() {
	runtime.Go(func() {
		for {
			if !s.elect.IsMaster() {
				time.Sleep(time.Second)
				continue
			}

			ok, err := s.clearInvalidTxnLog()
			if err != nil {
				logs.Errorf("err:%v", err)
			}

			logs.Debugf("I proc redo log clearer.Is it true: %v", ok)
			time.Sleep(time.Minute)
		}
	})
}

// 重载上次进程退出前未完成的同步任务
func (s *SyncTaskSched) initRedoSyncTaskLoader() {
	reloadIntervalSec := time.Duration(1)
	if s.cfg.ReloadRedoLogIntervalSec > 0 {
		reloadIntervalSec = time.Duration(s.cfg.ReloadRedoLogIntervalSec)
	}
	go func() {
		for {
			if !s.elect.IsMaster() {
				time.Sleep(time.Second)
				continue
			}

			// 本地没有任务了，去窃取一些回来同步，也可以避免其他进程异常退出了，异常进程的任务没有执行。
			if err := s.stealSyncTasks(); err != nil {
				logs.Errorf("err:%v", err)
			}

			time.Sleep(time.Second * reloadIntervalSec)
		}
	}()
}

// 本进程没有同步任务了去窃取一些任务，同时可以避免其他进程有异常退出的情况导致任务没有进行
func (s *SyncTaskSched) stealSyncTasks() error {
	if s.queue.Size() > 0 {
		return nil
	}
	err := s.loadAllRedoLog()
	if err != nil {
		logs.Errorf("err:%v", err)
		return err
	}
	return nil
}

// 加载redo log
func (s *SyncTaskSched) loadAllRedoLog() error {

	// 加载需要重试的同步任务
	appIds := redisDeal.RedisDoSMEMBERS(s.cfg.TaskAppIdSetRedisKey)
	for _, appId := range appIds {
		var cursor int64
		redoLogKey := s.getTxnRedoRedisKey(int32(appId))
		if time.Since(s.lastLoadRedoLogTime) > time.Minute {
			s.lastLoadRedoLogTime = time.Now()
			for {

				cursor, mapTaskStrToRedoScore, err := redisDeal.RedisDoZScan(redoLogKey, cursor, 1000)
				if err != nil {
					logs.Importantf("redisDeal.RedisDoZscan failed, key:%s, cursor:%s, err:%v", redoLogKey.Key, cursor, err)
					break
				}

				for _, task := range mapTaskStrToRedoScore {

					taskStr := fmt.Sprintf("%s", task)
					task, err := s.cfg.DecodeSyncTaskFromString(taskStr)
					if err != nil {
						logs.Errorf("err:%v", err)
						return err
					}

					logs.Debugf("steal task:%s redo log", taskStr)

					err = s.writeOpLog(int32(appId), task)
					if err != nil {
						logs.Errorf("err:%v", err)
						return err
					}
				}

				if cursor == 0 {
					break
				}

				time.Sleep(time.Millisecond * 100)
			}
		}

		opLogKey := s.cfg.GenTaskSetRedisKeyHdl(int32(appId))
		for {

			cursor, mapTaskStrToRedoScore, err := redisDeal.RedisDoZScan(opLogKey, cursor, 1000)
			if err != nil {
				logs.Importantf("redisDeal.RedisDoZscan failed, key:%s, cursor:%s, err:%v", opLogKey.Key, cursor, err)
				break
			}

			for _, task := range mapTaskStrToRedoScore {

				taskStr := fmt.Sprintf("%s", task)
				task, err := s.cfg.DecodeSyncTaskFromString(taskStr)
				if err != nil {
					logs.Errorf("err:%v, taskStr:%s", err, taskStr)
					return err
				}

				logs.Debugf("steal task:%s op log", taskStr)

				err = s.enqueueSyncTask(int32(appId), task)
				if err != nil {
					logs.Errorf("err:%v", err)
					return err
				}
			}

			if cursor == 0 {
				break
			}

			time.Sleep(time.Millisecond * 100)
		}
	}

	return nil
}

// 清理已经完成的同步任务，把redo日志移除
func (s *SyncTaskSched) clearInvalidTxnLog() (bool, error) {
	appIds := redisDeal.RedisDoSMEMBERS(s.cfg.TaskAppIdSetRedisKey)
	for _, appId := range appIds {
		err := redisDeal.RedisSendZremrangebyscore(s.cfg.GenTaskSetRedisKeyHdl(int32(appId)), -9999, 0)
		if err != nil {
			logs.Errorf("err:%v", err)
			return false, err
		}

		err = redisDeal.RedisSendZremrangebyscore(s.getTxnRedoRedisKey(int32(appId)), -9999, 0)
		if err != nil {
			logs.Errorf("err:%v", err)
			return false, err
		}
	}

	return true, nil
}

func (s *SyncTaskSched) getTxnRedoRedisKey(appId int32) *redisKeys.RedisKeys {
	redoLogKey := s.cfg.GenTaskSetRedisKeyHdl(appId)
	redoLogKey.Key += ".redo"
	return redoLogKey
}

func (s *SyncTaskSched) cancelRedoLog(appId int32, task SyncTask) error {
	err := redisDeal.RedisDoZincrby(s.getTxnRedoRedisKey(appId), -1, task.String(), redisKeys.InfoTtlDay)
	if err != nil {
		logs.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (s *SyncTaskSched) writeRedoLog(appId int32, task SyncTask) error {
	errGrp := runtime.NewErrGrp()

	errGrp.Go(func() error {
		err := redisDeal.RedisSendSadd(s.cfg.TaskAppIdSetRedisKey, fmt.Sprintf("%d", appId), redisKeys.InfoTtlDay)
		if err != nil {
			logs.Errorf("err:%v", err)
			return err
		}
		return nil
	})

	errGrp.Go(func() error {
		err := redisDeal.RedisDoZincrby(s.getTxnRedoRedisKey(appId), 1, task.String(), redisKeys.InfoTtlDay)
		if err != nil {
			logs.Errorf("err:%v", err)
			return err
		}
		return nil
	})

	if err := errGrp.Wait(); err != nil {
		logs.Errorf("err:%v", err)
		return err
	}

	return nil
}

func (s *SyncTaskSched) writeOpLog(appId int32, task SyncTask) error {
	errGrp := runtime.NewErrGrp()

	errGrp.Go(func() error {
		err := s.cancelRedoLog(appId, task)
		if err != nil {
			logs.Errorf("err:%v", err)
			return err
		}
		return nil
	})

	errGrp.Go(func() error {
		err := redisDeal.RedisSendSadd(s.cfg.TaskAppIdSetRedisKey, fmt.Sprintf("%d", appId), redisKeys.InfoTtlDay)
		if err != nil {
			logs.Errorf("err:%v", err)
			return err
		}
		return nil
	})

	errGrp.Go(func() error {
		err := redisDeal.RedisDoZincrby(s.cfg.GenTaskSetRedisKeyHdl(appId), 1, task.String(), redisKeys.InfoTtlDay)
		if err != nil {
			logs.Errorf("err:%v", err)
			return err
		}
		return nil
	})

	if err := errGrp.Wait(); err != nil {
		logs.Errorf("err:%v", err)
		return err
	}

	return nil
}

// PreWriteRedoLog 预写redo日志。用于避免进程准备任务任务数据时还没来得及注册任务异常退出了,下次新的master加载重做事务日志保证任务完成
func (s *SyncTaskSched) PreWriteRedoLog(appId int32, task SyncTask) error {
	logs.Debugf("pre write sync task redo log, task:%+v:", task)

	if err := s.writeRedoLog(appId, task); err != nil {
		logs.Errorf("err:%v", err)
		return err
	}

	return nil
}

type SyncTask interface {
	GetBucketId() int64
	String() string
}

func (s *SyncTaskSched) checkRepeatOrMemSyncTask(appId int32, syncTask SyncTask) bool {
	s.mu.RLock()
	taskStrs, ok := s.mapAppIdToTaskStrs[appId]
	if ok {
		if _, ok = taskStrs[syncTask.String()]; ok {
			s.mu.RUnlock()
			return true
		}
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	taskStrs, ok = s.mapAppIdToTaskStrs[appId]
	if ok {
		if _, ok = taskStrs[syncTask.String()]; ok {
			return true
		}
	} else {
		taskStrs = make(map[string]interface{})
		s.mapAppIdToTaskStrs[appId] = taskStrs
	}
	taskStrs[syncTask.String()] = struct{}{}

	return false
}

// RegSyncTask 注册同步任务
func (s *SyncTaskSched) RegSyncTask(appId int32, syncTask SyncTask) error {
	logs.Debugf("reg sync task, task:%+v:", syncTask)

	if !s.IsMaster() {
		if s.cfg.SendSyncTaskRegEvtForMasterHdl != nil {
			s.cfg.SendSyncTaskRegEvtForMasterHdl(syncTask)
			return nil
		}
	}

	if err := s.writeOpLog(appId, syncTask); err != nil {
		logs.Errorf("err:%v", err)
		return err
	}

	if !s.IsMaster() {
		return nil
	}

	if err := s.enqueueSyncTask(appId, syncTask); err != nil {
		logs.Importantf("reg task(%+v) failed, err:%v", syncTask, err)
		return err
	}

	return nil
}

func (s *SyncTaskSched) enqueueSyncTask(appId int32, syncTask SyncTask) error {
	if s.checkRepeatOrMemSyncTask(appId, syncTask) {
		return nil
	}

	task := bucketQueue.NewTask(syncTask.GetBucketId(), fmt.Sprintf("syncTask.%s", syncTask.String()), func(task *bucketQueue.Task) error {
		logs.Debugf("exec sync task, task:%+v:", syncTask)

		s.mu.Lock()
		taskStrs, ok := s.mapAppIdToTaskStrs[appId]
		if ok {
			delete(taskStrs, syncTask.String())
		}
		s.mu.Unlock()

		taskZetRedisKey := s.cfg.GenTaskSetRedisKeyHdl(appId)
		updateCnt, _ := redisDeal.RedisDoZScore(taskZetRedisKey, syncTask.String())
		logs.Infof("updateCnt:%d", updateCnt)
		if updateCnt <= 0 {
			return nil
		}

		if err := s.cfg.OnExecSync(syncTask); err != nil {
			logs.Errorf("err:%v", err)
			s.checkRepeatOrMemSyncTask(appId, syncTask)
			return err
		}

		err := redisDeal.RedisDoZincrby(taskZetRedisKey, -updateCnt, syncTask.String(), redisKeys.InfoTtlDay)
		if err != nil {
			logs.Errorf("err:%v", err)
		}

		return nil
	}, 0, 10)

	if err := s.queue.Reg(task); err != nil {
		logs.Importantf("reg task(%+v) failed, err:%v", task, err)
		return err
	}

	return nil
}

// TxnRegSyncTask 便捷地开启一个注册同步任务的事务（包含预写重做日志，注册任务步骤）
func (s *SyncTaskSched) TxnRegSyncTask(appId int32, syncTask SyncTask, prepareTxnFn func() error) error {
	if err := s.PreWriteRedoLog(appId, syncTask); err != nil {
		logs.Errorf("err:%v", err)
		return err
	}

	if err := prepareTxnFn(); err != nil {
		logs.Errorf("err:%v", err)
		return err
	}

	if err := s.RegSyncTask(appId, syncTask); err != nil {
		logs.Errorf("err:%v", err)
		return err
	}

	return nil
}
