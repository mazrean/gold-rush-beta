package api

import (
	"context"
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mazrean/gold-rush-beta/openapi"
)

var (
	exploreCalledNum int64 = 0

	exploreMetricsLocker = sync.RWMutex{}
	exploreRetryNum      = []int{}

	exploreRequestTimeLocker = sync.Mutex{}
	exploreRequestTime       = []int64{}
)

func Explore(ctx context.Context, area *openapi.Area) *openapi.Report {
	atomic.AddInt64(&exploreCalledNum, 1)

	var (
		i      int
		report openapi.Report
		//res    *http.Response
		err error
	)
	for i = 0; ; i++ {
		startTime := time.Now()
		report, _, err = api.ExploreArea(ctx).Args(*area).Execute()
		requestTime := time.Since(startTime).Milliseconds()
		exploreRequestTimeLocker.Lock()
		exploreRequestTime = append(exploreRequestTime, requestTime)
		exploreRequestTimeLocker.Unlock()

		if err == nil {
			break
		}

		var apiErr openapi.GenericOpenAPIError
		ok := errors.As(err, &apiErr)
		if ok {
			log.Printf("explore error(%s):%+v\n", apiErr.Error(), apiErr.Model().(openapi.ModelError))
		} else {
			log.Printf("explore error:%+v\n", err)
		}
	}

	exploreMetricsLocker.Lock()
	exploreRetryNum = append(exploreRetryNum, i)
	exploreMetricsLocker.Unlock()

	return &report
}
