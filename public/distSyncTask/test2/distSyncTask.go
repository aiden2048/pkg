package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/public/distSyncTask"
	"github.com/aiden2048/pkg/public/redisKeys"

	"github.com/aiden2048/pkg/frame"
	"github.com/aiden2048/pkg/public/mongodb"
	"github.com/aiden2048/pkg/public/redisDeal"
)

func main() {
	if err := frame.InitConfig("activity", &frame.FrameOption{}); err != nil {
		log.Fatalf("InitConfig error:%s", err.Error())
		return
	}

	// 初始化mongo
	if err := mongodb.StartMgoDb(mongodb.WLevel2); err != nil {
		log.Fatalf("InitMongodb %+v failed: %s", frame.GetMgoCoinfig(), err.Error())
		return
	}
	// 初始化REDIS
	if err := redisDeal.StartRedis(); err != nil {
		log.Fatalf("InitRedis failed: %s", err.Error())
		return
	}
	// 初始化activity Redis
	if err := redisDeal.StartActivityRedis(); err != nil {
		log.Fatalf("InitActivityRedis failed: %s", err.Error())
		return
	}

	InitDistSyncTask()
	select {}
}

func InitDistSyncTask() {
	sched, err := distSyncTask.NewSyncTaskSched(&distSyncTask.SyncTaskSchedConf{
		Name:                "distSyncTaskSchedTest",
		BucketSyncConcurMax: 1,
		ConcurSyncWorkerNum: 32,
		TaskAppIdSetRedisKey: &redisKeys.RedisKeys{
			Name: redisKeys.REDIS_INDEX_COMMON,
			Key:  "testDistSyncTaskAppIdSet",
		},
		GenTaskSetRedisKeyHdl: func(appId int32) *redisKeys.RedisKeys {
			return &redisKeys.RedisKeys{
				Name: redisKeys.REDIS_INDEX_COMMON,
				Key:  fmt.Sprintf("testDistSyncTaskAppId:%d", appId),
			}
		},
		DecodeSyncTaskFromString: func(s string) (distSyncTask.SyncTask, error) {
			uid, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				logs.Errorf("err:%v", err)
				return nil, err
			}
			return &TestDistSyncTask{
				Uid: uid,
				Ts:  time.Now().Unix(),
			}, nil
		},
		OnExecSync: func(task distSyncTask.SyncTask) error {
			fmt.Printf("started 1 %+v %+v\n", task, task.(*TestDistSyncTask).Ts)
			time.Sleep(time.Second)
			return nil
		},
	})
	if err != nil {
		logs.Errorf("err:%v", err)
		panic(err)
	}

	for i := uint64(1); i < 1500; i++ {
		err = sched.PreWriteRedoLog(3001, &TestDistSyncTask{
			Uid: i,
		})
		if err != nil {
			logs.Errorf("err:%v", err)
			panic(err)
		}

		err = sched.RegSyncTask(3001, &TestDistSyncTask{
			Uid: i,
		})
		if err != nil {
			logs.Errorf("err:%v", err)
			panic(err)
		}
	}

	select {}
}

var _ distSyncTask.SyncTask = (*TestDistSyncTask)(nil)

type TestDistSyncTask struct {
	Uid uint64
	Ts  int64
}

func (t *TestDistSyncTask) GetBucketId() int64 {
	return int64(t.Uid) % 100
}

func (t *TestDistSyncTask) String() string {
	return fmt.Sprintf("%d", t.Uid)
}
