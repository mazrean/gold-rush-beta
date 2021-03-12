package main

var (
	manageChan   = make(chan struct{}, 100)
	coinAllowMap = [11]int{3, 5, 5, 5, 5, 5, 10, 10, 10, 10, 10}
)

func push(coinNum int) {
	for i := 0; i < coinAllowMap[coinNum]; i++ {
		manageChan <- struct{}{}
	}
}

func pop() {
	<-manageChan
}
