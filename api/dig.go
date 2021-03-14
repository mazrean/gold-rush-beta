package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mazrean/gold-rush-beta/openapi"
)

type treasureDepth struct {
	depth     int
	treasures []string
}

var (
	digCalledNum int64 = 0

	digMetricsLocker = sync.RWMutex{}
	digRetryNum      = []int{}
	digTreasureNum   = []int{}
	digTreasureList  = []*treasureDepth{}

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
			//log.Printf("dig not found(request:%+v): {requestTime: %d}\n", req, requestTime)

			treasures = []string{}
			break
		}
		var apiErr openapi.GenericOpenAPIError
		if res != nil && res.StatusCode == 403 {
			//log.Printf("dig not found(request:%+v): {requestTime: %d}\n", req, requestTime)

			if errors.As(err, &apiErr) {
				return nil, fmt.Errorf("dig 403 error(request:%+v): %+v", *dig, apiErr.Model())
			}
			return nil, fmt.Errorf("dig 403 error(request:%+v): %w", *dig, err)
		}
		if errors.As(err, &apiErr) {
			log.Printf("dig error(request:%+v): %+v", *dig, apiErr.Model())
		} else {
			log.Printf("dig error(request:%+v): %+v", *dig, err)
		}
	}

	digMetricsLocker.Lock()
	digRetryNum = append(digRetryNum, i)
	digTreasureNum = append(digTreasureNum, len(treasures))
	digTreasureList = append(digTreasureList, &treasureDepth{
		depth:     int(dig.Depth),
		treasures: treasures,
	})
	digMetricsLocker.Unlock()

	return treasures, nil
}
