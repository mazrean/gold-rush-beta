package main

var (
	manageChan = make(chan struct{}, 100)
)

func push() {
	manageChan <- struct{}{}
}

func pop() {
	<-manageChan
}
