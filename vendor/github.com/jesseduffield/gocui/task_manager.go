package gocui

import "sync"

// Tracks whether the program is busy (i.e. either something is happening on
// the main goroutine or a worker goroutine). Used by integration tests
// to wait until the program is idle before progressing.
type TaskManager struct {
	// each of these listeners will be notified when the program goes from busy to idle
	idleListeners []chan struct{}
	tasks         map[int]Task
	// auto-incrementing id for new tasks
	nextId       int
	pendingTasks []*PendingTask

	mutex sync.Mutex
}

func newTaskManager() *TaskManager {
	return &TaskManager{
		tasks:         make(map[int]Task),
		idleListeners: []chan struct{}{},
	}
}

func (self *TaskManager) PendingTaskNames() []string {
	names := make([]string, 0, len(self.pendingTasks))
	for _, task := range self.pendingTasks {
		names = append(names, task.DisplayText)
	}
	return names
}

func (self *TaskManager) NewTask() *TaskImpl {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	self.nextId++
	taskId := self.nextId

	onDone := func() { self.delete(taskId) }
	task := &TaskImpl{id: taskId, busy: true, onDone: onDone, withMutex: self.withMutex}
	self.tasks[taskId] = task

	return task
}

func (self *TaskManager) NewPendingTask(name string, cancel <-chan struct{}, begin <-chan struct{}) *PendingTask {
	underyling := self.NewTask()
	// TODO: Keep a record of which tasks are there

	self.mutex.Lock()
	defer self.mutex.Unlock()

	self.nextId++
	taskId := self.nextId
	onDone := func() { self.deletePendingTask(taskId) }

	pendingTask := PendingTask{
		id:          taskId,
		DisplayText: name,
		Cancel:      cancel,
		Begin:       begin,
		Underlying:  underyling,
		onDone:      onDone,
	}
	self.pendingTasks = append(self.pendingTasks, &pendingTask)
	return &pendingTask
}

func (self *TaskManager) addIdleListener(c chan struct{}) {
	self.idleListeners = append(self.idleListeners, c)
}

func (self *TaskManager) withMutex(f func()) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	f()

	// Check if all tasks are done
	for _, task := range self.tasks {
		if task.isBusy() {
			return
		}
	}

	// If we get here, all tasks are done, so
	// notify listeners that the program is idle
	for _, listener := range self.idleListeners {
		listener <- struct{}{}
	}
}

func (self *TaskManager) delete(taskId int) {
	self.withMutex(func() {
		delete(self.tasks, taskId)
	})
}

func (self *TaskManager) deletePendingTask(pendingTaskId int) {
	self.withMutex(func() {
		pendingTasks := make([]*PendingTask, 0, len(self.pendingTasks)-1)
		for _, task := range self.pendingTasks {
			if task.id != pendingTaskId {
				pendingTasks = append(pendingTasks, task)
			}
		}
		self.pendingTasks = pendingTasks
	})
}
