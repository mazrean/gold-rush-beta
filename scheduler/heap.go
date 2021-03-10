package scheduler

type Point struct {
	X      int32
	Y      int32
	Depth  int
	Amount int
}

func (p *Point) priority() float64 {
	return float64(p.Amount)
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
