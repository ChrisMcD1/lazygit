package gocui

type PendingTask struct {
	id          int
	DisplayText string
	Cancel      <-chan struct{}
	Begin       <-chan struct{}
	Underlying  Task
	onDone      func()
}

func (self *PendingTask) Done() {
	self.onDone()
	self.Underlying.Done()
}

func (self *PendingTask) DoCancel() {
	self.onDone()
}
