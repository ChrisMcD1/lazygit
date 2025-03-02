package gocui

type PendingTask struct {
	Cancel     <-chan struct{}
	Begin      <-chan struct{}
	IsWaiting  bool
	Underlying Task
}
