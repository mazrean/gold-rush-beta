package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mazrean/gold-rush-beta/api"
	"github.com/mazrean/gold-rush-beta/manager"
	"github.com/mazrean/gold-rush-beta/openapi"
	digScheduler "github.com/mazrean/gold-rush-beta/scheduler/dig"
	exploreScheduler "github.com/mazrean/gold-rush-beta/scheduler/explore"
)

var startTime time.Time

func main() {
	startTime = time.Now()

	api.Setup()
	digScheduler.Setup()
	exploreScheduler.Setup()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg := sync.WaitGroup{}

	timer := time.Tick(9*time.Minute + 40*time.Second)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println(<-timer)

		finish()
	}()

	schedule(ctx)

	wg.Wait()
}

func finish() {
	//log.Printf("finish called\n")
	sb := &strings.Builder{}

	api.Statistic(sb)
	digScheduler.Statistic(sb)
	exploreScheduler.Statistic(sb)

	log.Println(sb.String())
	log.Printf("cashChan:%d,digChan:%d,exploreChan:%d,digLicenseChan:%d, api.LicenseChan:%d,normalChan:%d,manageChan:%d\n",
		len(cashChan), len(digChan), len(exploreChan), len(digLicenseChan), len(api.LicenseChan), len(normalChan), len(manager.ManageChan))
}

const (
	exploreWorkerNum    = 5 //4はrate limitが厳しい
	licenseWorkerNum    = 3
	digWorkerNum        = 7
	cashWorkerNum       = 7
	middleWorkerNum     = 7
	normalWorkerNum     = 5
	channelBuf          = 100000
	exploreSubWorkerNum = 3
	reserveNum          = 10
)

type digArg struct {
	point  *digScheduler.Point
	isLast bool
}

var (
	cashChan    chan string
	digChan     chan *digArg
	exploreChan chan struct{}

	digLicenseChan chan struct{}

	normalChan  chan func()
	normalChan2 chan func()

	reservedLicenseNum int32 = 0

	size int32 = 16

	coinUses = [11]int{6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6}
)

func schedule(ctx context.Context) {
	cashChan = make(chan string, channelBuf)
	digChan = make(chan *digArg, channelBuf)
	exploreChan = make(chan struct{}, 10000000)

	digLicenseChan = make(chan struct{}, channelBuf)

	normalChan = make(chan func(), channelBuf)
	normalChan2 = make(chan func(), channelBuf)

	//sem := semaphore.NewWeighted(int64(totalWorkerNum))

	for i := 0; i < exploreWorkerNum; i++ {
		go func() {
			for range exploreChan {
				//sem.Acquire(ctx, 1)
				arg := exploreScheduler.Pop()
				explore(ctx, arg.Area)
				//sem.Release(1)
			}
		}()
	}

	for i := 0; i < licenseWorkerNum; i++ {
		go func() {
			for {
				coins := api.PreserveCoin(coinUses[int(time.Since(startTime).Minutes())])
				license(ctx, coins)
			}
		}()
	}

	for i := 0; i < cashWorkerNum; i++ {
		go func() {
		CASH_WORKER:
			for {
				select {
				case <-ctx.Done():
					break CASH_WORKER
				case arg := <-cashChan:
					//sem.Acquire(ctx, 1)
					cash(ctx, arg)
					//sem.Release(1)
				}
			}
		}()
	}

	for i := 0; i < digWorkerNum; i++ {
		go func() {
		REQUEST_WORKER:
			for {
				select {
				case <-ctx.Done():
					break REQUEST_WORKER
				case arg := <-digChan:
					//sem.Acquire(ctx, 1)
					dig(ctx, arg.point, arg.isLast)
					//sem.Release(1)
				}
			}
		}()
	}

	for i := 0; i < middleWorkerNum; i++ {
		go func() {
		DIG_SCHEDULER:
			for {
				select {
				case <-ctx.Done():
					break DIG_SCHEDULER
				case <-digLicenseChan:
					select {
					case <-ctx.Done():
						break DIG_SCHEDULER
					case license := <-api.LicenseChan:
						licenseID := license.ID
						point := digScheduler.Pop()
						point.Dig.LicenseID = licenseID
						digChan <- &digArg{
							point:  point,
							isLast: license.IsLast,
						}
					}
				case license := <-api.LicenseChan:
					select {
					case <-ctx.Done():
						break DIG_SCHEDULER
					case <-digLicenseChan:
						point := digScheduler.Pop()
						point.Dig.LicenseID = license.ID
						digChan <- &digArg{
							point:  point,
							isLast: license.IsLast,
						}
					}
				}
			}
		}()
	}

	for i := 0; i < normalWorkerNum; i++ {
		go func() {
			for fun := range normalChan2 {
				fun()
			}
		}()
	}

	for i := 0; i < normalWorkerNum; i++ {
		go func() {
			for fun := range normalChan {
				fun()
			}
		}()
	}

	var k int32
	for k = 0; k < exploreSubWorkerNum; k++ {
		go func(k int32) {
			maxX := 3500 * (k + 1) / exploreSubWorkerNum
			for i := 3500 * k / exploreSubWorkerNum; i < maxX; i++ {
				var j int32
				for j = 0; j < 3500; j += size {
					var sizeX int32 = 1
					sizeY := size
					if i+sizeX > maxX {
						sizeX = maxX - i
					}
					if j+sizeY > 3500 {
						sizeY = 3500 - j
					}
					exploreScheduler.Push(&exploreScheduler.Area{
						Area: &openapi.Area{
							PosX:  i,
							PosY:  j,
							SizeX: &sizeX,
							SizeY: &sizeY,
						},
						Amount: 0.2 * float64(sizeX) * float64(sizeY),
					})
					exploreChan <- struct{}{}
				}
			}
		}(k)
	}
}

func cash(ctx context.Context, arg string) {
	api.Cash(ctx, arg)
}

func insertDig(arg *digScheduler.Point) {
	if arg.Amount <= 0 {
		return
	}
	//log.Printf("x:%d,y:%d, depth:%d", arg.PosX, arg.PosY, arg.Depth)
	digScheduler.Push(arg)
	digLicenseChan <- struct{}{}
}

func dig(ctx context.Context, arg *digScheduler.Point, isLast bool) {
	treasures, err := api.Dig(ctx, arg.Dig)
	if err != nil {
		//log.Printf("failed to dig: %+v", err)
		return
	}
	if isLast {
		manager.Pop()
	}

	atomic.AddInt32(&arg.Depth, 1)
	arg.Amount -= int32(len(treasures))
	//log.Printf("dig: x:%d,y:%d, depth:%d", arg.PosX, arg.PosY, arg.Depth)
	insertDig(arg)

	if len(treasures) > 0 {
		normalChan2 <- func(treasures []string) func() {
			return func() {
				//log.Printf("insert to cash chan start\n")
				//defer log.Printf("insert to cash chan end\n")
				for _, treasure := range treasures {
					//log.Printf("cash channel send start\n")
					cashChan <- treasure
					//log.Printf("cash channel send end\n")
				}
			}
		}(treasures)
	}
}

func license(ctx context.Context, arg []int32) {
	//log.Printf("license start\n")
	//defer log.Printf("license end\n")
	manager.Push()
	_ = api.IssueLicense(ctx, arg)
	atomic.AddInt32(&reservedLicenseNum, -reserveNum)
	/*for i := 0; i < 10-int(license.DigAllowed); i++ {
		pop()
	}*/
	//log.Printf("license:%+v\n", license)
}

func insertExplore(arg *exploreScheduler.Area) {
	if arg.Amount <= 0 {
		return
	}
	exploreScheduler.Push(arg)
	exploreChan <- struct{}{}
}

func explore(ctx context.Context, arg *openapi.Area) {
	//log.Printf("explore start\n")
	//defer log.Printf("explore end\n")
	report := api.Explore(ctx, arg)
	//log.Printf("report:%+v\n", report)

	//log.Printf("license to channel start\n")
	normalChan <- func(report *openapi.Report) func() {
		return func() {
			//log.Printf("explore insertDig start\n")
			//defer log.Printf("explore insertDig end\n")
			if *report.Area.SizeX == 1 && *report.Area.SizeY == 1 {
				//log.Printf("explore: x:%d,y:%d", arg.PosX, arg.PosY)
				insertDig(&digScheduler.Point{
					Dig: &openapi.Dig{
						PosX:  report.Area.PosX,
						PosY:  report.Area.PosY,
						Depth: 1,
					},
					Amount: report.Amount,
				})
			} else if report.Amount > 0 {
				if *report.Area.SizeX != 1 {
					sizeX1 := *report.Area.SizeX / 2
					newArea1 := exploreScheduler.Area{
						Area: &openapi.Area{
							PosX:  report.Area.PosX,
							PosY:  report.Area.PosY,
							SizeX: &sizeX1,
							SizeY: report.Area.SizeY,
						},
						Amount: float64(report.Amount) * float64(sizeX1) / float64(*report.Area.SizeX),
					}
					sizeX2 := *report.Area.SizeX - sizeX1
					newArea2 := exploreScheduler.Area{
						Area: &openapi.Area{
							PosX:  report.Area.PosX + sizeX1,
							PosY:  report.Area.PosY,
							SizeX: &sizeX2,
							SizeY: report.Area.SizeY,
						},
						Amount: float64(report.Amount) * float64(sizeX2) / float64(*report.Area.SizeX),
					}
					//log.Printf("explore: %+v(%d,%d),%+v(%d,%d),%+v(%d,%d)\n", *report, *report.Area.SizeX, *report.Area.SizeY, *newArea1.Area, *newArea1.Area.SizeX, *newArea1.Area.SizeY, *newArea2.Area, *newArea2.Area.SizeX, *newArea2.Area.SizeY)
					insertExplore(&newArea1)
					insertExplore(&newArea2)
				} else {
					sizeY1 := *report.Area.SizeY / 2
					newArea1 := exploreScheduler.Area{
						Area: &openapi.Area{
							PosX:  report.Area.PosX,
							PosY:  report.Area.PosY,
							SizeX: report.Area.SizeX,
							SizeY: &sizeY1,
						},
						Amount: float64(report.Amount) * float64(sizeY1) / float64(*report.Area.SizeY),
					}
					sizeY2 := *report.Area.SizeY - sizeY1
					newArea2 := exploreScheduler.Area{
						Area: &openapi.Area{
							PosX:  report.Area.PosX,
							PosY:  report.Area.PosY + sizeY1,
							SizeX: report.Area.SizeX,
							SizeY: &sizeY2,
						},
						Amount: float64(report.Amount) * float64(sizeY2) / float64(*report.Area.SizeY),
					}
					//log.Printf("explore: %+v(%d,%d),%+v(%d,%d),%+v(%d,%d)\n", *report, *report.Area.SizeX, *report.Area.SizeY, *newArea1.Area, *newArea1.Area.SizeX, *newArea1.Area.SizeY, *newArea2.Area, *newArea2.Area.SizeX, *newArea2.Area.SizeY)
					insertExplore(&newArea1)
					insertExplore(&newArea2)
				}
			}
		}
	}(report)
	//log.Printf("license to channel end\n")
}
