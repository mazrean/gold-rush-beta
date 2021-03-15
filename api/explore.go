package api

import (
	"bytes"
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
)

var (
	exploreCalledNum int64 = 0

	exploreMetricsLocker = sync.RWMutex{}
	exploreRetryNum      = []int{}

	exploreRequestTimeLocker = sync.Mutex{}
	exploreRequestTime       = []int64{}
)

func Explore(ctx context.Context, area *openapi.Area) (*openapi.Report, error) {
	atomic.AddInt64(&exploreCalledNum, 1)

	sb := strings.Builder{}
	sb.WriteString(baseURL)
	sb.WriteString("/explore")

	buf := bytes.Buffer{}
	err := json.NewEncoder(&buf).Encode(area)
	if err != nil {
		return nil, fmt.Errorf("failed to encord response body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", sb.String(), &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	var (
		i      int
		report openapi.Report
	)
	for i = 0; ; i++ {
		startTime := time.Now()
		res, err := client.Do(req)
		requestTime := time.Since(startTime).Milliseconds()
		if err != nil {
			return nil, fmt.Errorf("failed to do http request: %w", err)
		}

		exploreRequestTimeLocker.Lock()
		exploreRequestTime = append(exploreRequestTime, requestTime)
		exploreRequestTimeLocker.Unlock()

		if res.StatusCode == 200 {
			err = json.NewDecoder(res.Body).Decode(&report)
			if err != nil {
				return nil, fmt.Errorf("failed to decord response body: %w", err)
			}
			break
		}

		if res != nil && res.StatusCode == 429 {
			continue
		}
		/*var apiErr openapi.ModelError
		err = json.NewDecoder(res.Body).Decode(&apiErr)
		log.Printf("explore error(%d):%+v\n", res.Status, apiErr)*/

		buf = bytes.Buffer{}
		err = json.NewEncoder(&buf).Encode(area)
		if err != nil {
			return nil, fmt.Errorf("failed to encord response body: %w", err)
		}
		req.Body = io.NopCloser(&buf)
	}

	exploreMetricsLocker.Lock()
	exploreRetryNum = append(exploreRetryNum, i)
	exploreMetricsLocker.Unlock()

	return &report, nil
}
