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
	"github.com/mazrean/gold-rush-beta/openapi"
	"github.com/mazrean/gold-rush-beta/scheduler"
)

var startTime time.Time

func main() {
	startTime = time.Now()

	api.Setup()
	scheduler.Setup()

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
	log.Printf("finish called\n")
	sb := &strings.Builder{}

	api.Statistic(sb)
	scheduler.Statistic(sb)

	log.Println(sb.String())
	log.Printf("cashChan:%d,digChan:%d,licenseChan:%d,exploreChan:%d,digLicenseChan:%d, api.LicenseChan:%d,normalChan:%d\n",
		len(cashChan), len(digChan), len(licenseChan), len(exploreChan), len(digReadyChan), len(api.LicenseChan), len(normalChan))
}

const (
	totalWorkerNum      = 9
	exploreWorkerNum    = 2
	licenseWorkerNum    = 1
	digWorkerNum        = 3
	cashWorkerNum       = 2
	middleWorkerNum     = 5
	normalWorkerNum     = 3
	channelBuf          = 100
	licenseSub          = 3
	exploreSubWorkerNum = 1
	reserveNum          = 10
)

var (
	cashChan    chan string
	digChan     chan *scheduler.Point
	licenseChan chan []int32
	exploreChan chan *openapi.Area

	//digLicenseChan chan struct{}
	digReadyChan chan struct{}

	normalChan chan func()

	reservedLicenseNum int32 = 0

	coinUses = [11]int{6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6}
)

func schedule(ctx context.Context) {
	cashChan = make(chan string, channelBuf)
	digChan = make(chan *scheduler.Point, channelBuf)
	licenseChan = make(chan []int32, channelBuf)
	exploreChan = make(chan *openapi.Area, channelBuf)

	//digLicenseChan = make(chan struct{}, channelBuf*100)
	digReadyChan = make(chan struct{}, channelBuf)

	normalChan = make(chan func(), channelBuf)

	insertLicense()

	//sem := semaphore.NewWeighted(int64(totalWorkerNum))

	for i := 0; i < exploreWorkerNum; i++ {
		go func() {
			for arg := range exploreChan {
				//sem.Acquire(ctx, 1)
				explore(ctx, arg)
				//sem.Release(1)
			}
		}()
	}

	for i := 0; i < licenseWorkerNum; i++ {
		first := true
		go func() {
		LICENSE_WORKER:
			for {
				if time.Since(startTime) < 9*time.Minute+50*time.Second {
					select {
					case <-ctx.Done():
						break LICENSE_WORKER
					case arg := <-licenseChan:
						//sem.Acquire(ctx, 1)
						license(ctx, arg)
						//sem.Release(1)
					}
					if first {
						first = false
						digReadyChan <- struct{}{}
					}
				} else {
					select {
					case <-ctx.Done():
						break LICENSE_WORKER
					case arg := <-cashChan:
						//sem.Acquire(ctx, 1)
						cash(ctx, arg)
						//sem.Release(1)
					}
				}
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

	for i := 0; i < middleWorkerNum; i++ {
		go func() {
		DIG_SCHEDULER:
			for {
				select {
				case <-ctx.Done():
					break DIG_SCHEDULER
				/*case <-digLicenseChan:
				select {
				case <-ctx.Done():
					break DIG_SCHEDULER
				case licenseID := <-api.LicenseChan:
					point := scheduler.Pop()
					point.Dig.LicenseID = licenseID
					digChan <- point
					if len(api.LicenseChan)+int(reservedLicenseNum) < licenseSub {
						insertLicense()
					}
				}*/
				case licenseID := <-api.LicenseChan:
					point := scheduler.Pop()
					point.Dig.LicenseID = licenseID
					digChan <- point
					if len(api.LicenseChan)+int(reservedLicenseNum) < licenseSub {
						insertLicense()
					}
				}
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

	var size int32 = 1

	for k := 0; k < exploreSubWorkerNum; k++ {
		go func(k int) {
			for i := 3500 * k / exploreSubWorkerNum; i < 3500*(k+1)/exploreSubWorkerNum; i++ {
				for j := 0; j < 3500; j++ {
					exploreChan <- &openapi.Area{
						PosX:  int32(i),
						PosY:  int32(j),
						SizeX: &size,
						SizeY: &size,
					}
				}
			}
		}(k)
	}

	for i := 0; i < digWorkerNum; i++ {
		<-digReadyChan
		go func() {
		REQUEST_WORKER:
			for {
				if time.Since(startTime) < 9*time.Minute+50*time.Second {
					select {
					case <-ctx.Done():
						break REQUEST_WORKER
					case arg := <-digChan:
						//sem.Acquire(ctx, 1)
						dig(ctx, arg)
						//sem.Release(1)
					}
				} else {
					select {
					case <-ctx.Done():
						break REQUEST_WORKER
					case arg := <-cashChan:
						//sem.Acquire(ctx, 1)
						cash(ctx, arg)
						//sem.Release(1)
					}
				}
			}
		}()
	}
}

func cash(ctx context.Context, arg string) {
	api.Cash(ctx, arg)
}

func insertDig(arg *scheduler.Point) {
	if arg.Amount <= 0 {
		return
	}
	//log.Printf("depth:%d", arg.Depth)
	scheduler.Push(arg)
	//digLicenseChan <- struct{}{}
}

func dig(ctx context.Context, arg *scheduler.Point) {
	treasures, err := api.Dig(ctx, arg.Dig)
	if err != nil {
		//log.Printf("failed to dig: %+v", err)
		return
	}

	arg.Depth++
	arg.Amount -= int32(len(treasures))
	insertDig(arg)
	pop()

	if len(treasures) > 0 {
		normalChan <- func(treasures []string) func() {
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

func insertLicense() {
	//log.Printf("insertLicense start\n")
	//defer log.Printf("insertLicense end\n")
	coins := api.PreserveCoin(coinUses[int(time.Since(startTime).Minutes())])
	atomic.AddInt32(&reservedLicenseNum, reserveNum)
	//log.Printf("coins:%+v\n", coins)
	licenseChan <- coins
	//log.Printf("license channel\n")
}

func license(ctx context.Context, arg []int32) {
	//log.Printf("license start\n")
	//defer log.Printf("license end\n")
	push()
	api.IssueLicense(ctx, arg)
	atomic.AddInt32(&reservedLicenseNum, -reserveNum)
	//log.Printf("license:%+v\n", license)
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
			insertDig(&scheduler.Point{
				Dig: &openapi.Dig{
					PosX:  report.Area.PosX,
					PosY:  report.Area.PosY,
					Depth: 1,
				},
				Amount: report.Amount,
			})
		}
	}(report)
	//log.Printf("license to channel end\n")
}
