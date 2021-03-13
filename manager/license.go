package manager

var (
	ManageChan = make(chan struct{}, 10)
)

func Push() {
	ManageChan <- struct{}{}
}

func Pop() {
	<-ManageChan
}
