package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mazrean/gold-rush-beta/openapi"
)

var (
	digCalledNum int64 = 0

	digMetricsLocker = sync.RWMutex{}
	digRetryNum      = []int{}
	digTreasureNum   = []int{}

	digRequestTimeLocker = sync.Mutex{}
	digRequestTime       = [10][]int64{
		{},
		{},
		{},
		{},
		{},
		{},
		{},
		{},
		{},
		{},
	}
)

func Dig(ctx context.Context, dig *openapi.Dig) ([]string, error) {
	atomic.AddInt64(&digCalledNum, 1)

	var (
		i         int
		treasures []string
		res       *http.Response
		err       error
	)
	for i = 0; ; i++ {
		startTime := time.Now()
		treasures, res, err = api.Dig(ctx).Args(*dig).Execute()
		requestTime := time.Since(startTime).Milliseconds()
		digRequestTimeLocker.Lock()
		digRequestTime[dig.Depth-1] = append(digRequestTime[dig.Depth-1], requestTime)
		digRequestTimeLocker.Unlock()

		if err == nil {
			break
		}
		if res != nil && res.StatusCode == 404 {
			//fmt.Printf("dig not found(request:%+v): {requestTime: %d}\n", req, requestTime)

			treasures = []string{}
			break
		}
		if res != nil && res.StatusCode == 403 {
			//fmt.Printf("dig not found(request:%+v): {requestTime: %d}\n", req, requestTime)

			return nil, fmt.Errorf("dig 403 error: %+v", err)
		}
	}

	digMetricsLocker.Lock()
	digRetryNum = append(digRetryNum, i)
	digTreasureNum = append(digTreasureNum, len(treasures))
	digMetricsLocker.Unlock()

	return treasures, nil
}
