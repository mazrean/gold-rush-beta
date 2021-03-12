package main

var (
	manageChan = make(chan struct{}, 10)
)

func setup() {
	for i := 0; i < 10; i++ {
		manageChan <- struct{}{}
	}
}

func push() {
	manageChan <- struct{}{}
}

func pop() {
	<-manageChan
}
