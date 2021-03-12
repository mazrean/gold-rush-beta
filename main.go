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

	fmt.Println(sb.String())
}

const (
	exploreWorkerNum    = 3
	requestWorkerNum    = 3
	middleWorkerNum     = 1
	normalWorkerNum     = 5
	channelBuf          = 100
	licenseSub          = 3
	exploreSubWorkerNum = 5
	reserveNum          = 10
)

var (
	cashChan    chan string
	digChan     chan *scheduler.Point
	licenseChan chan []int32
	exploreChan chan *openapi.Area

	digLicenseChan chan struct{}

	normalChan chan func()

	reservedLicenseNum int32 = 0

	coinUses = [11]int{6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6}
)

func schedule(ctx context.Context) {
	cashChan = make(chan string, channelBuf)
	digChan = make(chan *scheduler.Point, channelBuf)
	licenseChan = make(chan []int32, channelBuf)
	exploreChan = make(chan *openapi.Area, channelBuf)

	digLicenseChan = make(chan struct{}, channelBuf)

	normalChan = make(chan func(), channelBuf)

	insertLicense()

	for i := 0; i < exploreWorkerNum; i++ {
		go func() {
			for arg := range exploreChan {
				explore(ctx, arg)
			}
		}()
	}

	for i := 0; i < requestWorkerNum; i++ {
		go func() {
			for {
				if time.Since(startTime).Minutes() < 9 {
					select {
					case arg := <-licenseChan:
						license(ctx, arg)
						continue
					default:
					}

					select {
					case arg := <-licenseChan:
						license(ctx, arg)
						continue
					case arg := <-digChan:
						dig(ctx, arg)
						continue
					default:
					}

					select {
					case arg := <-licenseChan:
						license(ctx, arg)
					case arg := <-digChan:
						dig(ctx, arg)
					case arg := <-cashChan:
						cash(ctx, arg)
					}
				} else {
					select {
					case arg := <-cashChan:
						cash(ctx, arg)
						continue
					default:
					}

					select {
					case arg := <-cashChan:
						cash(ctx, arg)
						continue
					case arg := <-licenseChan:
						license(ctx, arg)
						continue
					default:
					}

					select {
					case arg := <-cashChan:
						cash(ctx, arg)
					case arg := <-licenseChan:
						license(ctx, arg)
					case arg := <-digChan:
						dig(ctx, arg)
					}
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
					case licenseID := <-api.LicenseChan:
						point := scheduler.Pop()
						point.Dig.LicenseID = licenseID
						digChan <- point
						if len(api.LicenseChan)+int(reservedLicenseNum) < licenseSub {
							insertLicense()
						}
					}
				case licenseID := <-api.LicenseChan:
					select {
					case <-ctx.Done():
						break DIG_SCHEDULER
					case <-digLicenseChan:
						point := scheduler.Pop()
						point.Dig.LicenseID = licenseID
						digChan <- point
						if len(api.LicenseChan)+int(reservedLicenseNum) < licenseSub {
							insertLicense()
						}
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
	digLicenseChan <- struct{}{}
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
