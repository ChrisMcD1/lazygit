package gocui

type PendingTask struct {
	id              int
	CancelListeners []chan<- struct{}
	Underlying      Task
	onDone          func()
}

func (self *PendingTask) Done() {
	self.onDone()
	self.Underlying.Done()
}

func (self *PendingTask) DoCancel() {
	self.onDone()
	for _, listener := range self.CancelListeners {
		listener <- struct{}{}
	}
}
