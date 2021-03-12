package manager

var (
	manageChan = make(chan struct{}, 10)
)

func Push() {
	manageChan <- struct{}{}
}

func Pop() {
	<-manageChan
}
