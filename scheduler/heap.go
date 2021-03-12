package scheduler

import "github.com/mazrean/gold-rush-beta/openapi"

type Point struct {
	*openapi.Dig
	Amount int32
}

var (
	depthTimeMap = [10]float64{8, 9, 10, 11, 12, 12.5, 13, 13.5, 14, 14.5}
	depthCoinMap = [10]float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
)

func (p *Point) priority() float64 {
	return float64(p.Amount) * depthCoinMap[p.Depth-1] / ((11 - float64(p.Depth)) * depthTimeMap[p.Depth-1])
}

type PointHeap []*Point

func (ph PointHeap) Len() int { return len(ph) }

func (ph PointHeap) Less(i, j int) bool {
	return ph[i].priority() > ph[j].priority()
}

func (ph PointHeap) Swap(i, j int) { ph[i], ph[j] = ph[j], ph[i] }

func (ph *PointHeap) Push(x interface{}) {
	*ph = append(*ph, x.(*Point))
}

func (ph *PointHeap) Pop() interface{} {
	old := *ph
	n := len(old)
	x := old[n-1]
	*ph = old[0 : n-1]
	return x
}
