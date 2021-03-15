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

func Cash(ctx context.Context, treasure string) error {
	atomic.AddInt64(&cashCalledNum, 1)

	sb := strings.Builder{}
	sb.WriteString(baseURL)
	sb.WriteString("/cash")
	body := fmt.Sprintf(`"%s"`, treasure)
	req, err := http.NewRequestWithContext(ctx, "POST", sb.String(), strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}

	var (
		i        int
		coinList []int32
		//res      *http.Response
	)
	for i = 0; ; i++ {
		startTime := time.Now()
		res, err := client.Do(req)
		requestTime := time.Since(startTime).Milliseconds()
		if err != nil {
			return fmt.Errorf("failed to do http request: %w", err)
		}
		defer res.Body.Close()

		cashRequestTimeLocker.Lock()
		cashRequestTime = append(cashRequestTime, requestTime)
		cashRequestTimeLocker.Unlock()

		if res.StatusCode == 200 {
			err = json.NewDecoder(res.Body).Decode(&coinList)
			if err != nil {
				return fmt.Errorf("failed to decord response body: %w", err)
			}
			break
		}
		/*var apiErr openapi.ModelError
		err = json.NewDecoder(res.Body).Decode(&apiErr)
		if err != nil {
			return fmt.Errorf("failed to decord response body: %w", err)
		}*/
		req.Body = io.NopCloser(strings.NewReader(body))
		//log.Printf("cache error(%s):%+v\n", res.StatusCode, apiErr)
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

	return nil
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
