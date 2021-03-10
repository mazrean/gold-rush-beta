package scheduler

import (
	"container/heap"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	heapLocker = sync.Mutex{}
	ph         = &PointHeap{}

	pushCalledNum int64 = 0
	popCalledNum  int64 = 0

	schedulerMetricsLocker = sync.Mutex{}
	pushPriorities         = []float64{}
	pushDepth              = []int32{}
	pushAmount             = []int32{}
	popPriorities          = []float64{}
	popDepth               = []int32{}
	popAmount              = []int32{}
)

func Setup() {
	heap.Init(ph)
}

func Push(point *Point) {
	atomic.AddInt64(&pushCalledNum, 1)

	heapLocker.Lock()
	heap.Push(ph, point)
	heapLocker.Unlock()

	schedulerMetricsLocker.Lock()
	pushDepth = append(pushDepth, point.Depth)
	pushAmount = append(pushAmount, point.Amount)
	schedulerMetricsLocker.Unlock()
}

func Pop() *Point {
	atomic.AddInt64(&popCalledNum, 1)

	heapLocker.Lock()
	iPoint := heap.Pop(ph)
	heapLocker.Unlock()

	point := iPoint.(*Point)

	schedulerMetricsLocker.Lock()
	popDepth = append(popDepth, point.Depth)
	popAmount = append(popAmount, point.Amount)
	schedulerMetricsLocker.Unlock()

	return point
}

func Statistic(sb strings.Builder) {
	var avePushDepth float64 = 0
	for _, depth := range pushDepth {
		avePushDepth += float64(depth)
	}
	avePushDepth /= float64(len(pushDepth))

	var avePushAmount float64 = 0
	for _, amount := range pushAmount {
		avePushAmount += float64(amount)
	}
	avePushAmount /= float64(len(pushAmount))

	var avePopDepth float64 = 0
	for _, depth := range popDepth {
		avePopDepth += float64(depth)
	}
	avePopDepth /= float64(len(popDepth))

	var avePopAmount float64 = 0
	for _, amount := range popAmount {
		avePopAmount += float64(amount)
	}
	avePopAmount /= float64(len(popAmount))

	sb.WriteString(fmt.Sprintf(`scheduler:
	push:
		called num:%d
		depth:%g
		amount:%g
	pop:
		called num:%d
		depth:%g
		amount:%g`,
		pushCalledNum,
		avePushDepth,
		avePushAmount,
		popCalledNum,
		avePopDepth,
		avePopAmount,
	))
}
