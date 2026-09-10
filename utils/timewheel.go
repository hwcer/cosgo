package utils

import (
	"container/list"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// time wheel struct
type TimeWheel struct {
	interval       time.Duration
	ticker         *time.Ticker
	slots          []*list.List
	currentPos     int
	slotNum        int
	addTaskChannel chan *task
	stopChannel    chan struct{}
	stopOnce       sync.Once
	taskRecord     *sync.Map
}

// Job callback function
type Job func(TaskData)

// TaskData callback params
type TaskData map[any]any

// task struct
// times/taskData/interval 会被调用方goroutine(RemoveTask/UpdateTask)与
// 时间轮goroutine(scanAddRunTask/addTask)并发读写,因此使用原子类型
type task struct {
	interval atomic.Int64
	times    atomic.Int32 //-1:no limit >=1:run times 0:removed
	circle   int          //仅时间轮goroutine读写
	key      any
	job      Job
	taskData atomic.Pointer[TaskData]
}

// New create a empty time wheel
func New(interval time.Duration, slotNum int) *TimeWheel {
	if interval <= 0 || slotNum <= 0 {
		return nil
	}
	tw := &TimeWheel{
		interval:       interval,
		slots:          make([]*list.List, slotNum),
		currentPos:     0,
		slotNum:        slotNum,
		addTaskChannel: make(chan *task),
		stopChannel:    make(chan struct{}),
		taskRecord:     &sync.Map{},
	}

	tw.init()

	return tw
}

// Start start the time wheel
func (tw *TimeWheel) Start() {
	tw.ticker = time.NewTicker(tw.interval)
	go tw.start()
}

// Stop 停止时间轮,幂等:重复调用安全。
// 注意: 已启动的 task.job goroutine(见 scanAddRunTask)是 fire-and-forget,
// Stop 不会等待它们,也无法中断其运行。调用方需要同步等待的话应自行协调 context。
func (tw *TimeWheel) Stop() {
	tw.stopOnce.Do(func() {
		close(tw.stopChannel)
	})
}

func (tw *TimeWheel) start() {
	for {
		select {
		case <-tw.ticker.C:
			tw.tickHandler()
		case task := <-tw.addTaskChannel:
			tw.addTask(task)
		case <-tw.stopChannel:
			tw.ticker.Stop()
			return
		}
	}
}

// AddTask add new task to the time wheel
func (tw *TimeWheel) AddTask(interval time.Duration, times int, key any, data TaskData, job Job) error {
	if interval <= 0 || key == nil || job == nil || times < -1 || times == 0 {
		return errors.New("illegal task params")
	}

	_, ok := tw.taskRecord.Load(key)
	if ok {
		return errors.New("duplicate task key")
	}

	t := &task{key: key, job: job}
	t.interval.Store(int64(interval))
	t.times.Store(int32(times))
	t.taskData.Store(&data)

	//时间轮停止后start goroutine已退出,无接收者;select避免调用方永久阻塞
	select {
	case tw.addTaskChannel <- t:
		return nil
	case <-tw.stopChannel:
		return errors.New("time wheel stopped")
	}
}

// RemoveTask remove the task from time wheel
func (tw *TimeWheel) RemoveTask(key any) error {
	if key == nil {
		return nil
	}

	value, ok := tw.taskRecord.Load(key)

	if !ok {
		return errors.New("task not exists, please check you task key")
	} else {
		// lazy remove task
		task := value.(*task)
		task.times.Store(0)
		tw.taskRecord.Delete(task.key)
	}
	return nil
}

// UpdateTask update task times and data
func (tw *TimeWheel) UpdateTask(key any, interval time.Duration, taskData TaskData) error {
	if key == nil {
		return errors.New("illegal key, please try again")
	}

	value, ok := tw.taskRecord.Load(key)

	if !ok {
		return errors.New("task not exists, please check you task key")
	}
	task := value.(*task)
	task.taskData.Store(&taskData)
	task.interval.Store(int64(interval))
	return nil
}

// time wheel initialize
func (tw *TimeWheel) init() {
	for i := 0; i < tw.slotNum; i++ {
		tw.slots[i] = list.New()
	}
}

func (tw *TimeWheel) tickHandler() {
	l := tw.slots[tw.currentPos]
	tw.scanAddRunTask(l)
	if tw.currentPos == tw.slotNum-1 {
		tw.currentPos = 0
	} else {
		tw.currentPos++
	}
}

// add task
func (tw *TimeWheel) addTask(task *task) {
	if task.times.Load() == 0 {
		return
	}

	pos, circle := tw.getPositionAndCircle(time.Duration(task.interval.Load()))
	task.circle = circle

	tw.slots[pos].PushBack(task)

	//record the task
	tw.taskRecord.Store(task.key, task)
}

// scan task list and run the task
func (tw *TimeWheel) scanAddRunTask(l *list.List) {

	if l == nil {
		return
	}

	for item := l.Front(); item != nil; {
		task := item.Value.(*task)
		times := task.times.Load()

		if times == 0 {
			next := item.Next()
			l.Remove(item)
			tw.taskRecord.Delete(task.key)
			item = next
			continue
		}

		if task.circle > 0 {
			task.circle--
			item = item.Next()
			continue
		}

		if data := task.taskData.Load(); data != nil {
			go task.job(*data)
		}
		next := item.Next()
		l.Remove(item)
		item = next

		if times == 1 {
			task.times.Store(0)
			tw.taskRecord.Delete(task.key)
		} else {
			if times > 0 {
				task.times.Store(times - 1)
			}
			tw.addTask(task)
		}
	}
}

// get the task position
// 使用整数除法计算tick数,避免interval小于1秒时int(Seconds())截断为0导致除零panic
func (tw *TimeWheel) getPositionAndCircle(d time.Duration) (pos int, circle int) {
	delay := int(d / tw.interval)
	circle = delay / tw.slotNum
	pos = (tw.currentPos + delay) % tw.slotNum
	return
}
