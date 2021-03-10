package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mazrean/gold-rush-beta/openapi"
)

const (
	RequestRoutineNum = 3
	CalcRoutineNum    = 2
	RequestChanLen    = 10
	CalcChanLen       = 10
	Area              = 3500
	ExploreArea       = 1
)

var (
	startTime time.Time

	client = openapi.NewAPIClient(&openapi.Configuration{
		Servers: openapi.ServerConfigurations{
			{
				URL: fmt.Sprintf("http://%s:8000", os.Getenv("ADDRESS")),
			},
		},
		HTTPClient: http.DefaultClient,
		Debug:      false,
	})
	api = client.DefaultApi

	digQueue = make(chan func(licenseID int32) func(context.Context), 20)

	licenseLocker                  = sync.RWMutex{}
	license       *openapi.License = nil

	isLicenseQueuedLocker = sync.RWMutex{}
	isLicenseQueued       = false

	coinsLocker = sync.RWMutex{}
	coins       = []int32{}

	coinUses = [10]int{2, 2, 2, 2, 1, 1, 1, 1, 1, 0}

	cacheChan   = make(chan func(context.Context), RequestChanLen)
	licenseChan = make(chan func(context.Context), RequestChanLen)
	digChan     = make(chan func(context.Context), RequestChanLen)
	exploreChan = make(chan func(context.Context), RequestChanLen)
	calcChan    = make(chan func(context.Context), CalcChanLen)
)

func main() {
	startTime = time.Now()
	fmt.Println(startTime.String())

	ctx := context.Background()

	wg := sync.WaitGroup{}

	for i := 0; i < RequestRoutineNum; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for {
				var requestFunc func(context.Context)
				select {
				case requestFunc = <-cacheChan:
					//fmt.Println("cache")
				case requestFunc = <-digChan:
					//fmt.Println("dig")
				case requestFunc = <-licenseChan:
					//fmt.Println("license")
				case requestFunc = <-exploreChan:
				}
				fmt.Printf("request func start(routine: %d):%+v\n", i, time.Now())
				requestFunc(ctx)
			}
		}(i)
	}

	for i := 0; i < CalcRoutineNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for calcFunc := range calcChan {
				calcFunc(ctx)
			}
		}()
	}

	var sizeX int32 = ExploreArea
	var sizeY int32 = ExploreArea
	for i := 0; i < Area; i += ExploreArea {
		for j := 0; j < Area; j += ExploreArea {
			exploreChan <- explore(&openapi.Area{
				PosX:  int32(i),
				PosY:  int32(j),
				SizeX: &sizeX,
				SizeY: &sizeY,
			})
		}
	}

	wg.Wait()
}

func licenses(req []int32) func(context.Context) {
	return func(ctx context.Context) {
		var licenseVal openapi.License
		var res *http.Response
		var err error
		for {
			licenseVal, res, err = api.IssueLicense(ctx).Args(req).Execute()
			if err != nil {
				var apiErr openapi.GenericOpenAPIError = err.(openapi.GenericOpenAPIError)
				ok := errors.As(err, &apiErr)
				if ok {
					fmt.Printf("license error(%s):%+v\n", apiErr.Error(), apiErr.Model().(openapi.ModelError))
				}
				fmt.Println("license error:", err)

				continue
			}
			fmt.Printf("license request succeeded(%d):%+v\n", len(req), licenseVal)
			break
		}

		if res.StatusCode != 200 {
			return
		}

		licenseLocker.Lock()
		license = &licenseVal
		var digNum int32
		if len(digQueue) <= int(license.DigAllowed) {
			license.DigUsed = int32(len(digQueue))
			digNum = int32(len(digQueue))
		} else {
			license.DigUsed = licenseVal.DigAllowed
			digNum = licenseVal.DigAllowed
		}
		licenseLocker.Unlock()

		isLicenseQueuedLocker.Lock()
		isLicenseQueued = false
		isLicenseQueuedLocker.Unlock()

		for i := 0; i < int(digNum); i++ {
			digChan <- (<-digQueue)(license.Id)
		}
	}
}

func explore(req *openapi.Area) func(context.Context) {
	return func(ctx context.Context) {
		report, res, err := api.ExploreArea(ctx).Args(*req).Execute()
		if err != nil {
			fmt.Println("explore error:", err)
			return
		}
		calcChan <- func(ctx context.Context) {
			if res.StatusCode != 200 || report.Amount == 0 {
				return
			}

			if report.Area.GetSizeX() != 1 || report.Area.GetSizeY() != 1 {
				sizeX := (req.GetSizeX() + 1) / 2
				sizeY := (req.GetPosY() + 1) / 2
				exploreChan <- explore(&openapi.Area{
					PosX:  req.PosX,
					PosY:  req.PosY,
					SizeX: &sizeX,
					SizeY: &sizeY,
				})

				lessSizeX := req.GetSizeX() - sizeX
				if lessSizeX != 0 {
					exploreChan <- explore(&openapi.Area{
						PosX:  req.PosX + sizeX,
						PosY:  req.PosY,
						SizeX: &lessSizeX,
						SizeY: &sizeY,
					})
				}

				lessSizeY := req.GetSizeY() - sizeY
				if lessSizeY != 0 {
					exploreChan <- explore(&openapi.Area{
						PosX:  req.PosX,
						PosY:  req.PosY + sizeY,
						SizeX: &sizeX,
						SizeY: &lessSizeY,
					})
				}

				if lessSizeX != 0 && lessSizeY != 0 {
					exploreChan <- explore(&openapi.Area{
						PosX:  req.PosX + sizeX,
						PosY:  req.PosY + sizeY,
						SizeX: &lessSizeX,
						SizeY: &lessSizeY,
					})
				}
			} else {
				digFunc := dig(&openapi.Dig{
					PosX:  req.PosX,
					PosY:  req.PosY,
					Depth: 1,
				}, int(report.Amount))
				if digFunc != nil {
					digChan <- digFunc
				}
			}
		}
	}
}

func dig(req *openapi.Dig, amount int) func(context.Context) {
	isLicenseQueuedLocker.RLock()
	newIsLicenseQueued := isLicenseQueued
	isLicenseQueuedLocker.RUnlock()
	if newIsLicenseQueued {
		digQueue <- func(licenseID int32) func(context.Context) {
			return func(ctx context.Context) {
				req.LicenseID = licenseID
				var treasures []string
				var res *http.Response
				var err error
				for {
					startTime := time.Now()
					treasures, res, err = api.Dig(ctx).Args(*req).Execute()
					requestTime := time.Since(startTime).Milliseconds()
					if res.StatusCode == 404 {
						fmt.Printf("dig not found(depth:%d): {requestTime: %d}\n", req.Depth, requestTime)

						calcChan <- func(ctx context.Context) {
							fmt.Println("set next dig")
							req.Depth++
							digChan <- dig(req, amount)
						}
						return
					}
					if err != nil {
						var apiErr openapi.GenericOpenAPIError = err.(openapi.GenericOpenAPIError)
						ok := errors.As(err, &apiErr)
						if ok {
							fmt.Printf("dig error(%s):%+v\n", apiErr.Error(), apiErr.Model().(openapi.ModelError))
						}
						fmt.Println("dig error:", err)
						continue
					}
					fmt.Printf("dig succeeded(depth:%d): {treasures: %s, requestTime: %d}\n", req.Depth, strings.Join(treasures, ", "), requestTime)
					break
				}

				if res.StatusCode != 200 {
					return
				}

				for _, treasure := range treasures {
					cacheChan <- cache(treasure)
				}

				if len(treasures) < amount {
					req.Depth++
					digChan <- dig(req, amount-len(treasures))
				}
			}
		}

		return nil
	}

	licenseLocker.RLock()
	var remain int32 = 0
	if license != nil {
		remain = license.DigAllowed - license.DigUsed
	}
	licenseLocker.RUnlock()

	if remain < 1 {
		reqCoinLen := coinUses[9-int(time.Since(startTime).Minutes())]
		coinsLen := len(coins)
		coinsLocker.Lock()
		var reqCoins []int32
		if reqCoinLen <= coinsLen {
			reqCoins = coins[:reqCoinLen]
			coins = coins[reqCoinLen:]
		} else {
			reqCoins = coins[:]
			coins = coins[coinsLen:]
		}
		coinsLocker.Unlock()

		isLicenseQueuedLocker.Lock()
		isLicenseQueued = true
		isLicenseQueuedLocker.Unlock()

		licenseChan <- licenses(reqCoins)
		digQueue <- func(licenseID int32) func(ctx context.Context) {
			return func(ctx context.Context) {
				req.LicenseID = license.Id
				var treasures []string
				var res *http.Response
				var err error
				for {
					startTime := time.Now()
					treasures, res, err = api.Dig(ctx).Args(*req).Execute()
					requestTime := time.Since(startTime).Milliseconds()
					if res.StatusCode == 404 {
						fmt.Printf("dig not found(depth:%d): {requestTime: %d}\n", req.Depth, requestTime)

						calcChan <- func(ctx context.Context) {
							fmt.Println("set next dig")
							req.Depth++
							digChan <- dig(req, amount)
						}
						return
					}
					if err != nil {
						var apiErr openapi.GenericOpenAPIError = err.(openapi.GenericOpenAPIError)
						ok := errors.As(err, &apiErr)
						if ok {
							fmt.Printf("dig error(%s):%+v\n", apiErr.Error(), apiErr.Model().(openapi.ModelError))
						}
						fmt.Println("dig error:", err)
						continue
					}
					fmt.Printf("dig succeeded(depth:%d): {treasures: %s, requestTime: %d}\n", req.Depth, strings.Join(treasures, ", "), requestTime)
					break
				}

				calcChan <- func(ctx context.Context) {
					if res.StatusCode != 200 {
						return
					}

					for _, treasure := range treasures {
						cacheChan <- cache(treasure)
					}

					if len(treasures) < amount {
						req.Depth++
						digChan <- dig(req, amount-len(treasures))
					}
				}
			}
		}

		return nil
	}

	licenseLocker.RLock()
	req.LicenseID = license.Id
	atomic.AddInt32(&license.DigUsed, 1)
	licenseLocker.RUnlock()

	return func(ctx context.Context) {
		var treasures []string
		var res *http.Response
		var err error
		for {
			startTime := time.Now()
			treasures, res, err = api.Dig(ctx).Args(*req).Execute()
			requestTime := time.Since(startTime).Milliseconds()
			if res.StatusCode == 404 {
				fmt.Printf("dig not found(depth:%d): {requestTime: %d}\n", req.Depth, requestTime)

				calcChan <- func(ctx context.Context) {
					fmt.Println("set next dig")
					req.Depth++
					digChan <- dig(req, amount)
				}
				return
			}
			if err != nil {
				var apiErr openapi.GenericOpenAPIError = err.(openapi.GenericOpenAPIError)
				ok := errors.As(err, &apiErr)
				if ok {
					fmt.Printf("dig error(%s):%+v\n", apiErr.Error(), apiErr.Model().(openapi.ModelError))
				}
				fmt.Println("dig error:", err)
				continue
			}
			fmt.Printf("dig succeeded(depth:%d): {treasures: %s, requestTime: %d}\n", req.Depth, strings.Join(treasures, ", "), requestTime)
			break
		}

		calcChan <- func(ctx context.Context) {
			if res.StatusCode != 200 {
				return
			}

			for _, treasure := range treasures {
				cacheChan <- cache(treasure)
			}

			if len(treasures) < amount {
				digChan <- dig(req, amount-len(treasures))
			}
		}
	}
}

func cache(req string) func(context.Context) {
	req = fmt.Sprintf(`"%s"`, req)
	return func(ctx context.Context) {
		var newCoins []int32
		var res *http.Response
		var err error
		for {
			newCoins, res, err = api.Cash(ctx).Args(req).Execute()
			if err != nil {
				var apiErr openapi.GenericOpenAPIError = err.(openapi.GenericOpenAPIError)
				ok := errors.As(err, &apiErr)
				if ok {
					fmt.Printf("cache error(%s):%+v\n", apiErr.Error(), apiErr.Model().(openapi.ModelError))
				}
				fmt.Println("cache error:", err)
				continue
			}
			fmt.Printf("cash request succeeded(%s):%d\n", req, len(newCoins))
			break
		}
		if res.StatusCode != 200 {
			return
		}

		coinsLocker.Lock()
		coins = append(coins, newCoins...)
		coinsLocker.Unlock()
	}
}
