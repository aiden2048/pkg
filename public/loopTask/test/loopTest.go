package main

import (
	"fmt"
	"time"

	"github.com/aiden2048/pkg/frame"
	"github.com/aiden2048/pkg/public/loopTask"
	"github.com/aiden2048/pkg/public/mongodb"
	"github.com/aiden2048/pkg/public/redisDeal"
	uuid "github.com/satori/go.uuid"
)

type TestSubTask struct {
	Id            string
	Retry         bool
	PrintBoundary bool
	Attempted     int
}

func (t *TestSubTask) Exec() (nextInterval time.Duration, retry bool) {
	if t.PrintBoundary {
		fmt.Printf("=================\n")
	}
	fmt.Printf("sub task (Id:%s) exec. attempted:%d\n", t.Id, t.Attempted)
	if t.Retry {
		//fmt.Printf("i am ready to retry, %s\n", t.Id)
	}

	t.Attempted++
	return time.Second, true
	//return time.Second * 1, t.Retry
}

var _ loopTask.SubTask = (*TestSubTask)(nil)

type TestTask struct {
	Id string
}

func (t *TestTask) Ready() bool {
	fmt.Printf("task(id:%s) init.\n", t.Id)
	return true
}

func (t TestTask) GetSubTaskConcur() uint32 {
	//return 3
	return 1
}

func (t TestTask) NextBatchSubTasks() []loopTask.SubTask {
	return []loopTask.SubTask{
		//&TestSubTask{
		//	Id:    t.Id + "@subtask:1",
		//	Retry: true,
		//},
		&TestSubTask{
			Id:            t.Id + "@subtask:2",
			PrintBoundary: true,
		},
		&TestSubTask{
			Id:    t.Id + "@subtask:3",
			Retry: true,
		},
		&TestSubTask{
			Id:            t.Id + "@subtask:4",
			PrintBoundary: true,
		},
	}
}

func (t TestTask) OnProceed() {
	fmt.Printf("task(id:%s) on proceed.\n", t.Id)
}

var _ loopTask.Task = (*TestTask)(nil)

func main() {
	if err := frame.InitConfig("loopTaskTest", &frame.FrameOption{}); err != nil {
		fmt.Printf("InitConfig error:%s\n", err.Error())
		return
	}

	// 初始化mongo
	if err := mongodb.StartMgoDb(mongodb.WLevel1); err != nil {
		fmt.Printf("InitMongodb %+v failed: %s", frame.GetMgoCoinfig(), err.Error())
		return
	}

	if err := redisDeal.StartRedis(); err != nil {
		fmt.Printf("InitRedis failed: %s", err.Error())
		return
	}

	var refreshTimes int
	sched, err := loopTask.NewSched(loopTask.SchedCfg{
		Name:             "loopTask1",
		ConcurTaskMax:    3,
		GrpConcurTaskMax: 1,
		//GrpConcurTaskMax: 2,
		NewTaskHdl: func(taskId, taskGrp string) (loopTask.Task, error) {
			return &TestTask{
				Id: taskId,
			}, nil
		},
		RefreshTasksHdl: func() (mapGrpToTaskIds map[string][]string) {
			if refreshTimes > 0 {
				return map[string][]string{}
			}
			refreshTimes++
			taskIdMap := map[string][]string{
				//"1": {"@1", "@2", "@3", "@4"},
				//"2": {"@1", "@2", "@3"},
				"3": {"@1", "@2"},
				"4": {"@1"},
			}
			newTaskIdMap := map[string][]string{}
			for grpId, taskIds := range taskIdMap {
				var newTaskIds []string
				for _, taskId := range taskIds {
					newTaskIds = append(newTaskIds, "grp:"+grpId+"@task:"+taskId+"@"+uuid.NewV4().String()+"@refresh:"+fmt.Sprintf("%d", refreshTimes))
				}
				newTaskIdMap[grpId] = newTaskIds
			}
			return newTaskIdMap
		},
		CheckTaskStoppedHdl: func(taskIds []string) (stoppedTaskIds []string) {
			return nil
		},
		RefreshTaskIntervalMs:    1000,
		CheckTasksStopIntervalMs: 1000,
		SubTaskMaxAttempt:        300,
	})
	if err != nil {
		fmt.Printf("err:%v", err)
		return
	}
	sched.SetGrpConcur("3", 1)
	sched.Run()
	select {}
}
