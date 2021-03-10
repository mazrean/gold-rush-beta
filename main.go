package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mazrean/gold-rush-beta/openapi"
)

const (
	RequestRoutineNum = 4
	CalcRoutineNum    = 2
	RequestChanLen    = 50
	CalcChanLen       = 50
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

	licenseLocker                    = sync.RWMutex{}
	licenseList   []*openapi.License = []*openapi.License{}
	remain                           = 0

	isLicenseQueuedLocker = sync.RWMutex{}
	isLicenseQueued       = false

	coinsLocker = sync.RWMutex{}
	coins       = []int32{}

	coinUses = [10]int{8, 8, 8, 8, 8, 8, 8, 8, 1, 0}

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
				var state string
				select {
				case requestFunc = <-cacheChan:
					state = "cache"
				case requestFunc = <-digChan:
					state = "dig"
				case requestFunc = <-licenseChan:
					state = "license"
				case requestFunc = <-exploreChan:
					state = "explore"
				}
				if requestFunc == nil {
					fmt.Printf("nil func(state:%s)", state)
					continue
				}
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
				var apiErr openapi.GenericOpenAPIError
				ok := errors.As(err, &apiErr)
				if ok {
					//fmt.Printf("license error(%s):%+v\n", apiErr.Error(), apiErr.Model().(openapi.ModelError))
				} else {
					fmt.Println("license error:", err)
				}
				if res != nil && res.StatusCode == 409 {
					licenses, _, err := api.ListLicenses(ctx).Execute()
					if err != nil {
						var apiErr openapi.GenericOpenAPIError
						ok := errors.As(err, &apiErr)
						if ok {
							fmt.Printf("get license error(%s):%+v\n", apiErr.Error(), apiErr.Model().(openapi.ModelError))
						} else {
							fmt.Println("get license error:", err)
						}
					}

					fmt.Printf("licenses: %+v\n", licenses)
				}

				continue
			}
			fmt.Printf("license request succeeded(%d):%+v\n", len(req), licenseVal)
			break
		}

		if res != nil && res.StatusCode != 200 {
			return
		}

		licenseLocker.Lock()
		var digNum int32
		queueLen := len(digQueue)
		if queueLen < int(licenseVal.DigAllowed) {
			licenseVal.DigUsed = int32(queueLen)
			digNum = int32(queueLen)
			licenseList = append(licenseList, &licenseVal)
		} else {
			licenseVal.DigUsed = licenseVal.DigAllowed
			digNum = licenseVal.DigAllowed
		}
		remain += int(licenseVal.DigAllowed) - int(digNum)
		licenseLocker.Unlock()

		isLicenseQueuedLocker.Lock()
		isLicenseQueued = false
		isLicenseQueuedLocker.Unlock()

		for i := 0; i < int(digNum); i++ {
			digChan <- (<-digQueue)(licenseVal.Id)
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
			if !(res != nil && res.StatusCode == 200) || report.Amount == 0 {
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
					if res != nil && res.StatusCode == 404 {
						fmt.Printf("dig not found(request:%+v): {requestTime: %d}\n", req, requestTime)

						calcChan <- func(ctx context.Context) {
							//fmt.Println("set next dig")
							req.Depth++
							digFunc := dig(req, amount)
							if digFunc != nil {
								digChan <- digFunc
							}
						}
						return
					}
					if res != nil && res.StatusCode == 403 {
						fmt.Printf("dig invalid license(request: %+v)\n", req)
						licenses, _, err := api.ListLicenses(ctx).Execute()
						if err != nil {
							var apiErr openapi.GenericOpenAPIError
							ok := errors.As(err, &apiErr)
							if ok {
								fmt.Printf("get license error(%s):%+v\n", apiErr.Error(), apiErr.Model().(openapi.ModelError))
							} else {
								fmt.Println("get license error:", err)
							}
						}

						fmt.Printf("licenses: %+v\n", licenses)

						calcChan <- func(ctx context.Context) {
							//fmt.Println("set next dig")
							digFunc := dig(req, amount)
							if digFunc != nil {
								digChan <- digFunc
							}
						}
						return
					}
					if err != nil {
						var apiErr openapi.GenericOpenAPIError
						ok := errors.As(err, &apiErr)
						if ok {
							//fmt.Printf("dig error(request: %+v):%+v\n", req, apiErr.Model().(openapi.ModelError))
						}
						//fmt.Println("dig error:", err)
						continue
					}
					fmt.Printf("dig succeeded(request:%+v): {treasures: %s, requestTime: %d}\n", req, strings.Join(treasures, ", "), requestTime)
					break
				}

				if res != nil && res.StatusCode != 200 {
					return
				}

				for _, treasure := range treasures {
					cacheChan <- cache(treasure)
				}

				if len(treasures) < amount {
					req.Depth++
					digFunc := dig(req, amount-len(treasures))
					if digFunc != nil {
						digChan <- digFunc
					}
				}
			}
		}

		return nil
	}

	if remain < 1 {
		//fmt.Printf("coin use index:%d\n", int(time.Since(startTime).Minutes()))
		reqCoinLen := coinUses[int(time.Since(startTime).Minutes())]
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
				req.LicenseID = licenseID
				var treasures []string
				var res *http.Response
				var err error
				for {
					startTime := time.Now()
					treasures, res, err = api.Dig(ctx).Args(*req).Execute()
					requestTime := time.Since(startTime).Milliseconds()
					if res != nil && res.StatusCode == 404 {
						fmt.Printf("dig not found(request:%+v): {requestTime: %d}\n", req, requestTime)

						calcChan <- func(ctx context.Context) {
							//fmt.Println("set next dig")
							req.Depth++
							digFunc := dig(req, amount)
							if digFunc != nil {
								digChan <- digFunc
							}
						}
						return
					}
					if res != nil && res.StatusCode == 403 {
						fmt.Printf("dig invalid license(request: %+v)\n", req)
						licenses, _, err := api.ListLicenses(ctx).Execute()
						if err != nil {
							var apiErr openapi.GenericOpenAPIError
							ok := errors.As(err, &apiErr)
							if ok {
								fmt.Printf("get license error(%s):%+v\n", apiErr.Error(), apiErr.Model().(openapi.ModelError))
							} else {
								fmt.Println("get license error:", err)
							}
						}

						fmt.Printf("licenses: %+v\n", licenses)

						calcChan <- func(ctx context.Context) {
							//fmt.Println("set next dig")
							digFunc := dig(req, amount)
							if digFunc != nil {
								digChan <- digFunc
							}
						}
						return
					}
					if err != nil {
						var apiErr openapi.GenericOpenAPIError
						ok := errors.As(err, &apiErr)
						if ok {
							//fmt.Printf("dig error:%+v\n", apiErr.Model().(openapi.ModelError))
						}
						//fmt.Println("dig error:", err)
						continue
					}
					fmt.Printf("dig succeeded(request:%+v): {treasures: %s, requestTime: %d}\n", req, strings.Join(treasures, ", "), requestTime)
					break
				}

				calcChan <- func(ctx context.Context) {
					if res != nil && res.StatusCode != 200 {
						return
					}

					for _, treasure := range treasures {
						cacheChan <- cache(treasure)
					}

					if len(treasures) < amount {
						req.Depth++
						digFunc := dig(req, amount-len(treasures))
						if digFunc != nil {
							digChan <- digFunc
						}
					}
				}
			}
		}

		return nil
	}

	licenseLocker.Lock()
	req.LicenseID = licenseList[0].Id
	remain--
	licenseList[0].DigUsed++
	if licenseList[0].DigUsed == licenseList[0].DigAllowed {
		licenseList = licenseList[1:]
	}
	licenseLocker.Unlock()

	return func(ctx context.Context) {
		var treasures []string
		var res *http.Response
		var err error
		for {
			startTime := time.Now()
			treasures, res, err = api.Dig(ctx).Args(*req).Execute()
			requestTime := time.Since(startTime).Milliseconds()
			if res != nil && res.StatusCode == 404 {
				fmt.Printf("dig not found(request:%+v): {requestTime: %d}\n", req, requestTime)

				calcChan <- func(ctx context.Context) {
					//fmt.Println("set next dig")
					req.Depth++
					digFunc := dig(req, amount)
					if digFunc != nil {
						digChan <- digFunc
					}
				}
				return
			}
			if res != nil && res.StatusCode == 403 {
				fmt.Printf("dig invalid license(request: %+v)\n", req)
				licenses, _, err := api.ListLicenses(ctx).Execute()
				if err != nil {
					var apiErr openapi.GenericOpenAPIError
					ok := errors.As(err, &apiErr)
					if ok {
						fmt.Printf("get license error(%s):%+v\n", apiErr.Error(), apiErr.Model().(openapi.ModelError))
					} else {
						fmt.Println("get license error:", err)
					}
				}

				fmt.Printf("licenses: %+v\n", licenses)

				calcChan <- func(ctx context.Context) {
					//fmt.Println("set next dig")
					digFunc := dig(req, amount)
					if digFunc != nil {
						digChan <- digFunc
					}
				}
				return
			}
			if err != nil {
				var apiErr openapi.GenericOpenAPIError
				ok := errors.As(err, &apiErr)
				if ok {
					//fmt.Printf("dig error(request: %+v):%+v\n", req, apiErr.Model().(openapi.ModelError))
				}
				//fmt.Println("dig error:", err)
				continue
			}
			fmt.Printf("dig succeeded(request:%+v): {treasures: %s, requestTime: %d}\n", req, strings.Join(treasures, ", "), requestTime)
			break
		}

		calcChan <- func(ctx context.Context) {
			if res != nil && res.StatusCode != 200 {
				return
			}

			for _, treasure := range treasures {
				cacheChan <- cache(treasure)
			}

			if len(treasures) < amount {
				digFunc := dig(req, amount-len(treasures))
				if digFunc != nil {
					digChan <- digFunc
				}
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
				var apiErr openapi.GenericOpenAPIError
				ok := errors.As(err, &apiErr)
				if ok {
					fmt.Printf("cache error(%s):%+v\n", apiErr.Error(), apiErr.Model().(openapi.ModelError))
				}
				fmt.Println("cache error:", err)
				continue
			}
			//fmt.Printf("cash request succeeded(%s):%d\n", req, len(newCoins))
			break
		}
		if res != nil && res.StatusCode != 200 {
			return
		}

		coinsLocker.Lock()
		coins = append(coins, newCoins...)
		coinsLocker.Unlock()
	}
}
