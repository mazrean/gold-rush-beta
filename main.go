package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
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

	timer := time.Tick(9*time.Minute + 50*time.Second)
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
	sb := strings.Builder{}

	api.Statistic(sb)
	scheduler.Statistic(sb)

	fmt.Println(sb.String())
}

const (
	requestWorkerNum = 4
	normalWorkerNum  = 2
	channelBuf       = 100
)

var (
	cashChan    chan string
	licenseChan chan []int32
	exploreChan chan *openapi.Area

	digQueue    chan *scheduler.Point
	digQueueLen int

	normalChan chan func()

	digQueueCheckLocker = sync.Mutex{}
	isDigQueued         = false

	coinUses = [10]int{8, 8, 8, 8, 8, 8, 8, 8, 1, 0}
)

func schedule(ctx context.Context) {
	cashChan = make(chan string, channelBuf)
	licenseChan = make(chan []int32, channelBuf)
	exploreChan = make(chan *openapi.Area, channelBuf)

	digQueue = make(chan *scheduler.Point, channelBuf)

	normalChan = make(chan func(), channelBuf)

	for i := 0; i < requestWorkerNum; i++ {
		go func() {
		SCHEDULER:
			for {
				select {
				case arg := <-cashChan:
					//fmt.Printf("cash\n")
					cash(ctx, arg)
					continue
				case <-ctx.Done():
					break SCHEDULER
				default:
				}

				select {
				case arg := <-cashChan:
					//fmt.Printf("cash\n")
					cash(ctx, arg)
					continue
				case arg := <-licenseChan:
					//fmt.Printf("license\n")
					license(ctx, arg)
					continue
				case <-ctx.Done():
					break SCHEDULER
				default:
				}

				if scheduler.Len() > 0 {
					//fmt.Printf("dig\n")
					dig(ctx, scheduler.Pop())
					continue
				}

				select {
				case arg := <-cashChan:
					//fmt.Printf("cash\n")
					cash(ctx, arg)
					continue
				case arg := <-licenseChan:
					//fmt.Printf("license\n")
					license(ctx, arg)
					continue
				case arg := <-exploreChan:
					//fmt.Printf("explore\n")
					explore(ctx, arg)
				case <-ctx.Done():
					break SCHEDULER
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
	for i := 0; i < 3500; i++ {
		for j := 0; j < 3500; j++ {
			exploreChan <- &openapi.Area{
				PosX:  int32(i),
				PosY:  int32(j),
				SizeX: &size,
				SizeY: &size,
			}
		}
	}
}

func cash(ctx context.Context, arg string) {
	api.Cash(ctx, arg)
}

func insertDig(arg *scheduler.Point) {
	//fmt.Printf("insertDig start\n")
	//defer fmt.Printf("insertDig end\n")
	if arg.Amount == 0 {
		return
	}

	digQueueCheckLocker.Lock()
	if isDigQueued {
		//fmt.Printf("queued\n")
		digQueueLen += 1
		digQueueCheckLocker.Unlock()
		digQueue <- arg
		//fmt.Printf("queue setted\n")
		return
	}

	//fmt.Printf("preserve license\n")
	licenseID, err := api.PreserveLicense()
	if err != nil {
		//fmt.Printf("cannot preserve license\n")
		isDigQueued = true
		digQueueLen += 1
		digQueueCheckLocker.Unlock()
		digQueue <- arg
		//fmt.Printf("queue setted\n")
		insertLicense()
		//fmt.Printf("license channel setted\n")
		return
	}
	digQueueCheckLocker.Unlock()

	//fmt.Printf("licenseID:%d\n", licenseID)
	arg.Dig.LicenseID = licenseID
	scheduler.Push(arg)
}

func dig(ctx context.Context, arg *scheduler.Point) {
	treasures, err := api.Dig(ctx, arg.Dig)
	if err != nil {
		//fmt.Printf("failed to dig: %+v", err)
		return
	}

	if len(treasures) > 0 {
		normalChan <- func(treasures []string) func() {
			return func() {
				//fmt.Printf("insert to cash chan start\n")
				//defer fmt.Printf("insert to cash chan end\n")
				for _, treasure := range treasures {
					cashChan <- treasure
				}
			}
		}(treasures)
	}

	arg.Depth++
	insertDig(arg)
}

func insertLicense() {
	//fmt.Printf("insertLicense start\n")
	//defer fmt.Printf("insertLicense end\n")
	coins := api.PreserveCoin(coinUses[int(time.Since(startTime).Minutes())])
	//fmt.Printf("coins:%+v\n", coins)
	licenseChan <- coins
	//fmt.Printf("license channel\n")
}

func license(ctx context.Context, arg []int32) {
	//fmt.Printf("license start\n")
	//defer fmt.Printf("license end\n")
	license := api.IssueLicense(ctx, arg)
	//fmt.Printf("license:%+v\n", license)
	digQueueCheckLocker.Lock()
	isDigQueued = false
	//fmt.Printf("queue finish\n")
	digQueueCheckLocker.Unlock()

	//fmt.Printf("license to channel start\n")
	normalChan <- func() {
		//fmt.Printf("insertDig loop start\n")
		//defer fmt.Printf("insertDig loop end")
		var digNum int
		digQueueCheckLocker.Lock()
		if digQueueLen > int(license.DigAllowed) {
			digNum = int(license.DigAllowed)
			insertLicense()
		} else {
			digNum = digQueueLen
		}
		digQueueLen -= digNum
		digQueueCheckLocker.Unlock()

		for i := 0; i < digNum; i++ {
			insertDig(<-digQueue)
		}
	}
	//fmt.Printf("license to channel end\n")
}

func explore(ctx context.Context, arg *openapi.Area) {
	//fmt.Printf("explore start\n")
	//defer fmt.Printf("explore end\n")
	report := api.Explore(ctx, arg)
	//fmt.Printf("report:%+v\n", report)

	//fmt.Printf("license to channel start\n")
	normalChan <- func(report *openapi.Report) func() {
		return func() {
			//fmt.Printf("explore insertDig start\n")
			//defer fmt.Printf("explore insertDig end\n")
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
	//fmt.Printf("license to channel end\n")
}
