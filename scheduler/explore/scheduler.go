package explore

import (
	"container/heap"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	startTime  time.Time
	heapLocker = sync.Mutex{}
	ah         = &AreaHeap{}

	pushCalledNum int64 = 0
	popCalledNum  int64 = 0

	schedulerMetricsLocker = sync.Mutex{}
	pushPriorities         = []float64{}
	pushSize               = []int32{}
	pushAmount             = []float64{}
	popPriorities          = []float64{}
	popSize                = []int32{}
	popAmount              = []float64{}
)

func Setup() {
	startTime = time.Now()
	heap.Init(ah)
}

func Len() int {
	return len(*ah)
}

func Push(area *Area) {
	atomic.AddInt64(&pushCalledNum, 1)

	heapLocker.Lock()
	heap.Push(ah, area)
	heapLocker.Unlock()

	schedulerMetricsLocker.Lock()
	pushSize = append(pushSize, *area.SizeX**area.SizeY)
	pushAmount = append(pushAmount, area.Amount)
	schedulerMetricsLocker.Unlock()
}

func Pop() *Area {
	atomic.AddInt64(&popCalledNum, 1)

	heapLocker.Lock()
	iArea := heap.Pop(ah)
	heapLocker.Unlock()

	area := iArea.(*Area)

	schedulerMetricsLocker.Lock()
	popSize = append(popSize, *area.SizeX**area.SizeY)
	popAmount = append(popAmount, area.Amount)
	schedulerMetricsLocker.Unlock()

	return area
}

func Statistic(sb *strings.Builder) {
	var avePushSize float64 = 0
	for _, depth := range pushSize {
		avePushSize += float64(depth)
	}
	avePushSize /= float64(len(pushSize))

	var avePushAmount float64 = 0
	for _, amount := range pushAmount {
		avePushAmount += float64(amount)
	}
	avePushAmount /= float64(len(pushAmount))

	var avePopSize float64 = 0
	for _, depth := range popSize {
		avePopSize += float64(depth)
	}
	avePopSize /= float64(len(popSize))

	var avePopAmount float64 = 0
	for _, amount := range popAmount {
		avePopAmount += amount
	}
	avePopAmount /= float64(len(popAmount))

	sb.WriteString(fmt.Sprintf(`explore scheduler:
	push:
		called num:%d
		size:%g
		amount:%g
	pop:
		called num:%d
		size:%g
		amount:%g
`,
		pushCalledNum,
		avePushSize,
		avePushAmount,
		popCalledNum,
		avePopSize,
		avePopAmount,
	))
}
