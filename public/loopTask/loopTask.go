// Package loopTask 循环任务系统，循环加载任务，把任务按进度拆分成子多个自任务并行处理
// Sched为任务调度器，Task用于控制任务进度，SubTask为任务总进度中具体执行的各个子任务
package loopTask

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/frame/runtime"
	"github.com/aiden2048/pkg/public/elect"
	uuid "github.com/satori/go.uuid"
)

const (
	defaultRefreshTaskIntervalMs    = 5000
	defaultCheckTasksStopIntervalMs = 5000
)

const (
	taskEvtAdd = iota
	taskEvtDel
	taskEvtRefresh
)

type taskEvt struct {
	taskId  string
	taskGrp string
	evt     int
}

type SchedCfg struct {
	Name                     string                                           // 名称
	ConcurTaskMax            uint32                                           // 最大并发处理的任务数量
	GrpConcurTaskMax         uint32                                           // 相同任务组最大并发处理的任务数量
	NewTaskHdl               func(taskId, taskGrp string) (Task, error)       // 生成任务对象的处理器
	RefreshTasksHdl          func() (mapGrpToTaskIds map[string][]string)     // 刷新任务的间隔，返回任务组映射任务列表的map
	CheckTaskStoppedHdl      func(taskIds []string) (stoppedTaskIds []string) // 检查任务停止处理器
	RefreshTaskIntervalMs    int                                              // 刷新任务的间隔
	CheckTasksStopIntervalMs int                                              // 检查任务停止的间隔
	SubTaskMaxAttempt        int                                              // 子任务最大运行次数，限制重试次数，比如1，则代表只能运行一次，不重试
}

func NewSched(cfg SchedCfg) (*Sched, error) {
	if cfg.RefreshTasksHdl == nil {
		return nil, errors.New("refresh task list func is nil")
	}
	if cfg.CheckTaskStoppedHdl == nil {
		return nil, errors.New("check task stop func is nil")
	}
	if cfg.NewTaskHdl == nil {
		return nil, errors.New("new task func is nil")
	}
	if cfg.Name == "" {
		return nil, errors.New("missed name")
	}
	if cfg.RefreshTaskIntervalMs == 0 {
		cfg.RefreshTaskIntervalMs = defaultRefreshTaskIntervalMs
	}
	if cfg.CheckTasksStopIntervalMs == 0 {
		cfg.CheckTasksStopIntervalMs = defaultCheckTasksStopIntervalMs
	}
	// 选举器，同时间只有一个进程参与任务调度
	e, err := elect.NewRedisElect(cfg.Name, uuid.NewV4().String())
	if err != nil {
		logs.LogError("err:%v", err)
		return nil, err
	}
	sched := &Sched{
		elect:            e,
		SchedCfg:         cfg,
		taskGrpConcurRec: make(map[string]uint32),
		inQueueTasks:     make(map[string]*taskRunner),
		runningTasks:     make(map[string]*taskRunner),
		evtCh:            make(chan *taskEvt),
	}
	sched.init()
	return sched, nil
}

type Sched struct {
	SchedCfg
	elect              *elect.RedisElect
	taskGrpConcurRecMu sync.Mutex
	taskGrpConcurRec   map[string]uint32
	taskGrpConcurMaxes sync.Map // map[string]int32, -1代表不限制并发
	inQueueTasks       map[string]*taskRunner
	runningTasks       map[string]*taskRunner
	evtCh              chan *taskEvt
}

func (s *Sched) SetGrpConcur(grp string, concur int32) {
	if concur == 0 {
		s.taskGrpConcurMaxes.Delete(grp)
		return
	}
	if concur < 0 {
		concur = -1
	}
	s.taskGrpConcurMaxes.Store(grp, concur)
}

func (s *Sched) GetGrpConcur(grp string) (uint32, bool) {
	concurAny, ok := s.taskGrpConcurMaxes.Load(grp)
	if !ok {
		if s.GrpConcurTaskMax > 0 {
			return s.GrpConcurTaskMax, true
		}
		return 0, false
	}
	concur := concurAny.(int32)
	if concur < 0 {
		return 0, false
	}
	return uint32(concur), true
}

func (s *Sched) init() {
	s.elect.Run()
}

func (s *Sched) IsMaster() bool {
	return s.elect.IsMaster()
}

func (s *Sched) Run() {
	runtime.Go(func() {
		cleanStoppedTasksTk := time.NewTicker(time.Second)
		defer cleanStoppedTasksTk.Stop()
		refreshTasksTk := time.NewTicker(time.Duration(s.RefreshTaskIntervalMs) * time.Millisecond)
		defer refreshTasksTk.Stop()
		checkTasksStoppedTk := time.NewTicker(time.Duration(s.CheckTasksStopIntervalMs) * time.Millisecond)
		defer checkTasksStoppedTk.Stop()
		reSchedTk := time.NewTicker(time.Millisecond * 100)
		defer reSchedTk.Stop()
		for {
			_, _ = runtime.CreateTrace(nil)

			if !s.IsMaster() {
				logs.LogDebug("not master now")
				time.Sleep(time.Second)
				continue
			}

			select {
			case evt := <-s.evtCh:
				s.hdlEvt(evt)
			case <-cleanStoppedTasksTk.C:
				s.cleanStoppedTasks()
			case <-refreshTasksTk.C:
				s.refreshTasks()
			case <-checkTasksStoppedTk.C:
				s.checkTaskStopped()
			case <-reSchedTk.C:
				s.reSchedule()
			}
		}
	})
	select {
	case s.evtCh <- &taskEvt{evt: taskEvtRefresh}:
	default:
	}
}

func (s *Sched) Add(taskId, taskGrp string) {
	if !s.IsMaster() {
		logs.Trace("im not master")
		return
	}
	s.evtCh <- &taskEvt{
		taskGrp: taskGrp,
		taskId:  taskId,
		evt:     taskEvtAdd,
	}
}

func (s *Sched) Del(taskId string) {
	if !s.IsMaster() {
		logs.Trace("im not master")
		return
	}
	s.evtCh <- &taskEvt{
		taskId: taskId,
		evt:    taskEvtDel,
	}
}

func (s *Sched) hdlEvt(evt *taskEvt) {
	if !s.IsMaster() {
		logs.LogDebug("not master now")
		return
	}
	switch evt.evt {
	case taskEvtAdd:
		s.hdlAddEvt(evt.taskId, evt.taskGrp)
	case taskEvtDel:
		s.hdlDelEvt(evt.taskId)
	case taskEvtRefresh:
		s.refreshTasks()
	}
}

func (s *Sched) checkTaskStopped() {
	var taskIds []string
	for taskId := range s.inQueueTasks {
		taskIds = append(taskIds, taskId)
	}
	stoppedTaskIds := s.CheckTaskStoppedHdl(taskIds)
	for _, taskId := range stoppedTaskIds {
		s.hdlEvt(&taskEvt{
			taskId: taskId,
			evt:    taskEvtDel,
		})
	}
}

func (s *Sched) cleanStoppedTasks() {
	logs.LogDebug("task: %v clean stopped task", s.Name)
	var finishTaskIds []string
	for _, runner := range s.runningTasks {
		if runner.running.Load() {
			continue
		}
		if runner.stop.Load() {
			finishTaskIds = append(finishTaskIds, runner.taskId)
		}
	}
	for _, taskId := range finishTaskIds {
		delete(s.runningTasks, taskId)
		delete(s.inQueueTasks, taskId)
	}
	s.reSchedule()
}

func (s *Sched) refreshTasks() {
	mapGrpToTaskIds := s.RefreshTasksHdl()
	if mapGrpToTaskIds == nil {
		logs.LogDebug("mapGrpToTaskIds is nil")
		return
	}
	for grp, taskIds := range mapGrpToTaskIds {
		for _, taskId := range taskIds {
			logs.LogDebug("add task evt, task grp:%s, task id:%s", grp, taskId)
			s.hdlEvt(&taskEvt{
				taskId:  taskId,
				taskGrp: grp,
				evt:     taskEvtAdd,
			})
		}
	}
}

func (s *Sched) hdlDelEvt(taskId string) {
	delete(s.inQueueTasks, taskId)
	runner, ok := s.runningTasks[taskId]
	if !ok {
		return
	}
	delete(s.runningTasks, taskId)
	s.stopTask(runner)
	s.reSchedule()
}

func (s *Sched) stopTask(runner *taskRunner) {
	select {
	case runner.stopCh <- struct{}{}:
	default:
		logs.Trace("stop task chan full (%v_%v) ", s.Name, runner.taskId)
	}
}

func (s *Sched) reSchedule() {
	n := s.ConcurTaskMax - uint32(len(s.runningTasks))
	if n <= 0 {
		return
	}

	var (
		readyN           uint32
		readyTaskRunners []*taskRunner
	)
	for taskId, runner := range s.inQueueTasks {
		_, ok := s.runningTasks[taskId]
		if ok {
			continue
		}
		readyN++
		readyTaskRunners = append(readyTaskRunners, runner)
		if readyN >= n {
			break
		}
	}

	if len(readyTaskRunners) == 0 {
		return
	}

	for _, runner := range readyTaskRunners {
		s.schedNewTask(runner)
	}
}

func (s *Sched) hdlAddEvt(taskId, taskGrp string) {
	_, exists := s.inQueueTasks[taskId]
	if exists {
		logs.Debugf("task %s already in queue waiting", taskId)
		return
	}

	s.schedNewTask(newTaskRunner(taskId, taskGrp, s))
}

func (s *Sched) schedNewTask(runner *taskRunner) {
	s.inQueueTasks[runner.taskId] = runner
	concurTaskNum := uint32(len(s.runningTasks))
	if concurTaskNum >= s.ConcurTaskMax {
		logs.Debugf("running queue size over max, max:%d, current:%d", s.ConcurTaskMax, concurTaskNum)
		return
	}
	grpConcurMax, limited := s.GetGrpConcur(runner.taskGrp)
	if limited {
		s.taskGrpConcurRecMu.Lock()
		grpConcur := s.taskGrpConcurRec[runner.taskGrp]
		// 任务组并行任务数已经打到最大限制，排队进入排队等待，不进行调度
		if grpConcur >= grpConcurMax {
			s.taskGrpConcurRecMu.Unlock()
			logs.Debugf("grp %s concur over max, max:%d, current:%d", runner.taskGrp, s.GrpConcurTaskMax, grpConcur)
			return
		}
		// 任务组最大并发任务数+1
		s.taskGrpConcurRec[runner.taskGrp] = grpConcur + 1
		s.taskGrpConcurRecMu.Unlock()
	}
	s.runningTasks[runner.taskId] = runner
	runner.run()
}

type Task interface {
	Ready() bool
	GetSubTaskConcur() uint32
	NextBatchSubTasks() []SubTask
	OnProceed()
}

func newTaskRunner(taskId, taskGrp string, sched *Sched) *taskRunner {
	return &taskRunner{
		taskId:  taskId,
		taskGrp: taskGrp,
		sched:   sched,
		stopCh:  make(chan struct{}, 1),
	}
}

type taskRunner struct {
	taskId         string
	taskGrp        string
	sched          *Sched
	stopCh         chan struct{}
	subTaskRunners []*subTaskRunner
	running        atomic.Bool
	stop           atomic.Bool
}

func (r *taskRunner) run() {
	r.running.Store(true)
	r.stop.Store(false)
	runtime.Go(func() {
		defer func() {
			r.running.Store(false)
			r.stop.Store(true)

			r.sched.taskGrpConcurRecMu.Lock()
			defer r.sched.taskGrpConcurRecMu.Unlock()
			if concur := r.sched.taskGrpConcurRec[r.taskGrp]; concur > 0 {
				concur = concur - 1
				if concur > 0 {
					r.sched.taskGrpConcurRec[r.taskGrp] = concur
				} else {
					delete(r.sched.taskGrpConcurRec, r.taskGrp)
				}
			}
		}()

		task, err := r.sched.NewTaskHdl(r.taskId, r.taskGrp)
		if err != nil {
			logs.LogError("err:%v", err)
			return
		}

		if !task.Ready() {
			logs.LogDebug("task not ready.")
			return
		}

		concur := task.GetSubTaskConcur()

		var wg sync.WaitGroup
		for i := uint32(0); i < concur; i++ {
			subTasks := task.NextBatchSubTasks()
			// 没有子任务了
			if len(subTasks) == 0 {
				logs.LogDebug("sub tasks is empty now.")
				continue
			}
			wg.Add(1)
			subRunner := newSubTaskRunner(r.taskId, r.taskGrp, i, subTasks, &wg, r.sched)
			r.subTaskRunners = append(r.subTaskRunners, subRunner)
			subRunner.run()
		}

		var isAllSubTasksFinish bool
		runtime.Go(func() {
			wg.Wait()
			isAllSubTasksFinish = true
			select {
			case r.stopCh <- struct{}{}:
			default:
			}
		})

		<-r.stopCh

		// 任务正常运行完成，没有中断信号打断
		if isAllSubTasksFinish {
			logs.LogInfo("task(id:%s, taskGrp:%s) exec finish", r.taskId, r.taskGrp)
			task.OnProceed()
			return
		}

		// 中断信号打断导致任务执行退出,通知所有自任务退出
		for _, subRunner := range r.subTaskRunners {
			select {
			case subRunner.stopCh <- struct{}{}:
			default:
			}
		}
	})
}

type SubTask interface {
	Exec() (nextInterval time.Duration, retry bool)
}

func newSubTaskRunner(taskId, taskGrp string, id uint32, subTasks []SubTask, wg *sync.WaitGroup, sched *Sched) *subTaskRunner {
	if len(subTasks) == 0 {
		panic("sub list empty")
	}
	r := &subTaskRunner{
		nextExecAt: 0,
		stopCh:     make(chan struct{}, 1),
		wg:         wg,
		sched:      sched,
		taskId:     taskId,
		taskGrp:    taskGrp,
		id:         id,
	}
	for _, v := range subTasks {
		r.subTaskDescs = append(r.subTaskDescs, &SubTaskDesc{
			subTask: v,
		})
	}
	return r
}

type SubTaskDesc struct {
	subTask   SubTask
	attempted int
}

type subTaskRunner struct {
	subTaskDescs []*SubTaskDesc
	nextExecAt   int64
	stopCh       chan struct{}
	sched        *Sched
	wg           *sync.WaitGroup
	taskId       string
	taskGrp      string
	id           uint32
}

func (r *subTaskRunner) run() {
	runtime.Go(func() {
		defer func() {
			notDoneLen := len(r.subTaskDescs)
			logs.LogInfo("task: %v(%v_%v) run sub task is stop run, next run at %v, waiting for exec num %v",
				r.sched.Name, r.taskId, r.id, r.nextExecAt, notDoneLen)
			r.wg.Done()
		}()
		for len(r.subTaskDescs) > 0 {
			subTaskDesc := r.subTaskDescs[0]
			r.subTaskDescs = r.subTaskDescs[1:]
			next, retry := subTaskDesc.subTask.Exec()
			subTaskDesc.attempted++
			if retry {
				if subTaskDesc.attempted < r.sched.SubTaskMaxAttempt {
					r.subTaskDescs = append(r.subTaskDescs, subTaskDesc)
				}
			}
			r.nextExecAt = time.Now().Unix() + int64(next.Seconds())
			if next == 0 {
				select {
				case <-r.stopCh:
					goto out
				default:
				}
			} else {
				tk := time.NewTicker(next)
				exit := false
				select {
				case <-tk.C:
				case <-r.stopCh:
					exit = true
				}
				tk.Stop()
				if exit {
					goto out
				}
			}
		}
	out:
		return
	})
}
