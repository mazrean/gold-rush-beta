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
		cancel()

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
	channelBuf       = 50
)

var (
	cashChan    chan string
	licenseChan chan []int32
	exploreChan chan *openapi.Area

	digQueue chan *scheduler.Point

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
					fmt.Printf("cash")
					cash(ctx, arg)
					continue
				case <-ctx.Done():
					break SCHEDULER
				default:
				}

				select {
				case arg := <-cashChan:
					fmt.Printf("cash")
					cash(ctx, arg)
					continue
				case arg := <-licenseChan:
					fmt.Printf("license")
					license(ctx, arg)
					continue
				case <-ctx.Done():
					break SCHEDULER
				default:
				}

				if scheduler.Len() > 0 {
					fmt.Printf("dig")
					dig(ctx, scheduler.Pop().Dig)
					continue
				}

				select {
				case arg := <-cashChan:
					fmt.Printf("cash")
					cash(ctx, arg)
					continue
				case arg := <-licenseChan:
					fmt.Printf("license")
					license(ctx, arg)
					continue
				case arg := <-exploreChan:
					fmt.Printf("explore")
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
	digQueueCheckLocker.Lock()
	if isDigQueued {
		digQueueCheckLocker.Unlock()
		digQueue <- arg
		return
	}

	licenseID, err := api.PreserveLicense()
	if err != nil {
		isDigQueued = true
		digQueueCheckLocker.Unlock()
		digQueue <- arg
		insertLicense()
		return
	}
	digQueueCheckLocker.Unlock()

	arg.Dig.LicenseID = licenseID
	scheduler.Push(arg)
}

func dig(ctx context.Context, arg *openapi.Dig) {
	treasures, err := api.Dig(ctx, arg)
	if err != nil {
		fmt.Printf("failed to dig: %+v", err)
		return
	}

	if len(treasures) > 0 {
		normalChan <- func(treasures []string) func() {
			return func() {
				for _, treasure := range treasures {
					cashChan <- treasure
				}
			}
		}(treasures)
	}

	scheduler.Push(&scheduler.Point{})
}

func insertLicense() {
	coins := api.PreserveCoin(coinUses[int(time.Since(startTime).Minutes())])
	licenseChan <- coins
}

func license(ctx context.Context, arg []int32) {
	license := api.IssueLicense(ctx, arg)
	digQueueCheckLocker.Lock()
	isDigQueued = false
	digQueueCheckLocker.Unlock()

	normalChan <- func() {
		for i := 0; i < int(license.DigAllowed); i++ {
			insertDig(<-digQueue)
		}
	}
}

func explore(ctx context.Context, arg *openapi.Area) {
	report := api.Explore(ctx, arg)

	normalChan <- func(report *openapi.Report) func() {
		return func() {
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
}
