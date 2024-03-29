package explore

import (
	"math"
	"time"

	"github.com/mazrean/gold-rush-beta/openapi"
)

type Area struct {
	*openapi.Area
	Amount float64
}

var (
	depthTimeMap = [10]float64{8, 9, 10, 11, 12, 12.5, 13, 13.5, 14, 14.5}
	depthCoinMap = [10]float64{0.5, 1, 2, 3, 4, 5, 7.5, 10, 15, 35}
)

func (a *Area) priority(t *time.Time) float64 {
	if a.Amount == 0 {
		return 0.15
	}
	return a.Amount / math.Pow(float64(*a.SizeX)*float64(*a.SizeY), float64(t.Sub(startTime).Minutes()+1))
}

type AreaHeap []*Area

func (ah AreaHeap) Len() int { return len(ah) }

func (ah AreaHeap) Less(i, j int) bool {
	t := time.Now()
	return ah[i].priority(&t) > ah[j].priority(&t)
}

func (ah AreaHeap) Swap(i, j int) { ah[i], ah[j] = ah[j], ah[i] }

func (ah *AreaHeap) Push(x interface{}) {
	*ah = append(*ah, x.(*Area))
}

func (ah *AreaHeap) Pop() interface{} {
	old := *ah
	n := len(old)
	x := old[n-1]
	*ah = old[0 : n-1]
	return x
}
