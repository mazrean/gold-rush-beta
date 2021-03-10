package scheduler

import (
	"container/heap"
	"sync"
)

var (
	heapLocker = sync.Mutex{}
	ph         = &PointHeap{}
)

func Setup() {
	heap.Init(ph)
}

func Push(point *Point) {
	heapLocker.Lock()
	heap.Push(ph, &point)
	heapLocker.Unlock()
}

func Pop() *Point {
	heapLocker.Lock()
	iPoint := heap.Pop(ph)
	heapLocker.Unlock()

	return iPoint.(*Point)
}
