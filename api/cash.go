package api

import (
	"context"
	"errors"
	"fmt"
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
		coinList, _, err = api.Cash(ctx).Args(treasure).Execute()
		requestTime := time.Since(startTime).Nanoseconds()
		cashRequestTimeLocker.Lock()
		cashRequestTime = append(cashRequestTime, requestTime)
		cashRequestTimeLocker.Unlock()

		if err == nil {
			break
		}
		var apiErr openapi.GenericOpenAPIError
		ok := errors.As(err, &apiErr)
		if ok {
			//fmt.Printf("cache error(%s):%+v\n", apiErr.Error(), apiErr.Model().(openapi.ModelError))
		} else {
			fmt.Println("cache error:", err)
		}
	}

	cashMetricsLocker.Lock()
	coinNum = append(coinNum, len(coinList))
	cashRetryNum = append(cashRetryNum, i)
	cashMetricsLocker.Unlock()

	coinsLocker.Lock()
	coins = append(coins, coinList...)
	coinsLocker.Unlock()
}
