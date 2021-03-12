package main

var (
	manageChan = make(chan struct{}, 100)
)

func push() {
	for i := 0; i < 10; i++ {
		manageChan <- struct{}{}
	}
}

func pop() {
	<-manageChan
}
