// Package redis2dbSync redis同步数据库的组件
// Deprecated 该包将放弃维护，请使用distSyncTask包
package redis2dbSync

import (
	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/public/distSyncTask"
	"github.com/aiden2048/pkg/public/redisKeys"
)

type SyncCacheToDBConf struct {
	Name                                     string
	SyncQueueSize                            int
	ConcurSyncWorkerNum, BucketSyncConcurMax uint32
	ListenSyncTaskRegFromSlaveEvtName        string
	OnListenSyncTaskRegFromSlaveHdl          func([]byte)
	TaskAppIdSetRedisKey                     *redisKeys.RedisKeys                   // 把任务按分类应用id存储，方便按应用id查询挤压和管理任务，appId并非特指包网id,可以是任何业务归属应用id
	GenTaskSetRedisKeyHdl                    func(appId int32) *redisKeys.RedisKeys // 把任务按分类应用id存储，方便按应用id查询挤压和管理任务，appId并非特指包网id,可以是任何业务归属应用id
	DecodeSyncTaskFromString                 func(string) (SyncTask, error)
	OnExecSync                               func(task SyncTask) error
	ReloadRedoLogIntervalSec                 uint32
}

// Deprecated: 不再维护, 请使用distSyncTask.NewSyncTaskSched
func NewSyncCacheToDBSched(cfg *SyncCacheToDBConf) (*SyncCacheToDBSched, error) {
	distSyncTaskSched, err := distSyncTask.NewSyncTaskSched(&distSyncTask.SyncTaskSchedConf{
		Name:                              cfg.Name,
		SyncQueueSize:                     cfg.SyncQueueSize,
		ConcurSyncWorkerNum:               cfg.ConcurSyncWorkerNum,
		BucketSyncConcurMax:               cfg.BucketSyncConcurMax,
		ListenSyncTaskRegFromSlaveEvtName: cfg.ListenSyncTaskRegFromSlaveEvtName,
		OnListenSyncTaskRegFromSlaveHdl:   cfg.OnListenSyncTaskRegFromSlaveHdl,
		TaskAppIdSetRedisKey:              cfg.TaskAppIdSetRedisKey,
		GenTaskSetRedisKeyHdl:             cfg.GenTaskSetRedisKeyHdl,
		DecodeSyncTaskFromString: func(s string) (distSyncTask.SyncTask, error) {
			return cfg.DecodeSyncTaskFromString(s)
		},
		OnExecSync: func(task distSyncTask.SyncTask) error {
			return cfg.OnExecSync(task)
		},
		ReloadRedoLogIntervalSec: cfg.ReloadRedoLogIntervalSec,
	})
	if err != nil {
		logs.Errorf("err:%v", err)
		return nil, err
	}

	sched := &SyncCacheToDBSched{
		SyncTaskSched: distSyncTaskSched,
	}

	return sched, nil
}

type SyncCacheToDBSched struct {
	*distSyncTask.SyncTaskSched
}

type SyncTask interface {
	distSyncTask.SyncTask
}
