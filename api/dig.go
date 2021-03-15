package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mazrean/gold-rush-beta/openapi"
	"golang.org/x/sync/errgroup"
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

	sb := strings.Builder{}
	sb.WriteString(baseURL)
	sb.WriteString("/dig")

	pr, pw := io.Pipe()
	eg := errgroup.Group{}
	eg.Go(func() error {
		defer pr.Close()
		defer pw.Close()
		err := json.NewEncoder(pw).Encode(dig)
		if err != nil {
			return fmt.Errorf("failed to encord response body: %w", err)
		}

		return nil
	})

	req, err := http.NewRequestWithContext(ctx, "POST", sb.String(), pr)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	var (
		i         int
		treasures []string
	)
	for i = 0; ; i++ {
		startTime := time.Now()
		res, err := client.Do(req)
		requestTime := time.Since(startTime).Milliseconds()
		if err != nil {
			return nil, fmt.Errorf("failed to do http request: %w", err)
		}
		defer res.Body.Close()

		err = eg.Wait()
		if err != nil {
			return nil, fmt.Errorf("failed in error group: %w", err)
		}

		digRequestTimeLocker.Lock()
		digRequestTime[dig.Depth-1] = append(digRequestTime[dig.Depth-1], requestTime)
		digRequestTimeLocker.Unlock()

		if res.StatusCode == 200 {
			err = json.NewDecoder(res.Body).Decode(&treasures)
			if err != nil {
				return nil, fmt.Errorf("failed to decord response body: %w", err)
			}
			break
		}

		if res.StatusCode == 404 {
			//log.Printf("dig not found(request:%+v): {requestTime: %d}\n", req, requestTime)

			treasures = []string{}
			break
		}

		var apiErr openapi.ModelError
		err = json.NewDecoder(res.Body).Decode(&apiErr)
		if err != nil {
			return nil, fmt.Errorf("failed to decord response body: %w", err)
		}
		if res.StatusCode == 403 {
			//log.Printf("dig not found(request:%+v): {requestTime: %d}\n", req, requestTime)
			return nil, fmt.Errorf("dig 403 error(request:%+v): %+v", *dig, apiErr)
		}
		//log.Printf("dig error(request:%+v): %+v", *dig, apiErr)
		pr, pw = io.Pipe()
		eg := errgroup.Group{}
		eg.Go(func() error {
			defer pr.Close()
			defer pw.Close()
			err := json.NewEncoder(pw).Encode(dig)
			if err != nil {
				return fmt.Errorf("failed to encord response body: %w", err)
			}

			return nil
		})
		req.Body = pr
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
