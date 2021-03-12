package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mazrean/gold-rush-beta/openapi"
)

var (
	cashCalledNum int64 = 0

	cashMetricsLocker = sync.Mutex{}
	coinNum           = []int{}
	cashRetryNum      = []int{}

	cashTreasureCoinLocker = sync.RWMutex{}
	cashTreasureCoinMap    = map[string]int{}

	cashRequestTimeLocker = sync.Mutex{}
	cashRequestTime       = []int64{}

	coinsLocker = sync.RWMutex{}
	coins       = []int32{}
)

func Cash(ctx context.Context, treasure string) {
	atomic.AddInt64(&cashCalledNum, 1)

	var (
		i        int
		coinList []int32
		//res      *http.Response
		err error
	)
	for i = 0; ; i++ {
		startTime := time.Now()
		coinList, _, err = api.Cash(ctx).Args(fmt.Sprintf(`"%s"`, treasure)).Execute()
		requestTime := time.Since(startTime).Milliseconds()
		cashRequestTimeLocker.Lock()
		cashRequestTime = append(cashRequestTime, requestTime)
		cashRequestTimeLocker.Unlock()

		if err == nil {
			//log.Printf("cash succeeded\n")
			break
		}
		var apiErr openapi.GenericOpenAPIError
		ok := errors.As(err, &apiErr)
		if ok {
			//log.Printf("cache error(%s):%+v\n", apiErr.Error(), apiErr.Model().(openapi.ModelError))
		} else {
			log.Printf("cache error(%s):%+v(time:%s)\n", treasure, err, time.Now().String())
		}
	}

	cashMetricsLocker.Lock()
	coinNum = append(coinNum, len(coinList))
	cashRetryNum = append(cashRetryNum, i)
	cashMetricsLocker.Unlock()

	cashTreasureCoinLocker.Lock()
	cashTreasureCoinMap[treasure] = len(coinList)
	cashTreasureCoinLocker.Unlock()

	coinsLocker.Lock()
	coins = append(coins, coinList...)
	coinsLocker.Unlock()
}

func PreserveCoin(coinNum int) []int32 {
	var res []int32
	coinsLocker.Lock()
	if coinNum > len(coins) {
		if len(coins) < 1 {
			coinNum = 0
		} else if len(coins) < 6 {
			coinNum = 1
		} else {
			coinNum = 6
		}
	}
	res = coins[:coinNum]
	coins = coins[coinNum:]
	coinsLocker.Unlock()

	return res
}
