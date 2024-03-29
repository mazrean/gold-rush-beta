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

type License struct {
	ID     int32
	IsLast bool
}

var (
	issueLicenseCalledNum int64 = 0
	licenseMetricsLocker        = sync.RWMutex{}
	coinNumLicenses             = [11][]int8{
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
		{},
	}
	issueLicenseRetryNum = []int{}

	issueLicenseRequestTimeLocker = sync.Mutex{}
	issueLicenseRequestTime       = []int64{}

	LicenseChan = make(chan *License, 100)
)

func IssueLicense(ctx context.Context, coins []int32) (*openapi.License, error) {
	atomic.AddInt64(&issueLicenseCalledNum, 1)

	sb := strings.Builder{}
	sb.WriteString(baseURL)
	sb.WriteString("/dig")

	buf := bytes.Buffer{}
	err := json.NewEncoder(&buf).Encode(coins)
	if err != nil {
		return nil, fmt.Errorf("failed to encord response body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", sb.String(), &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	var (
		i       int
		license openapi.License
		//res     *http.Response
	)
	for i = 0; ; i++ {
		startTime := time.Now()
		res, err := client.Do(req)
		requestTime := time.Since(startTime).Milliseconds()
		if err != nil {
			return nil, fmt.Errorf("failed to do http request: %w", err)
		}
		defer res.Body.Close()

		issueLicenseRequestTimeLocker.Lock()
		issueLicenseRequestTime = append(issueLicenseRequestTime, requestTime)
		issueLicenseRequestTimeLocker.Unlock()

		if res.StatusCode == 200 {
			err = json.NewDecoder(res.Body).Decode(&license)
			if err != nil {
				return nil, fmt.Errorf("failed to decord response body: %w", err)
			}
			break
		}

		/*var apiErr openapi.ModelError
		err = json.NewDecoder(res.Body).Decode(&apiErr)
		if err != nil {
			return nil, fmt.Errorf("failed to decord response body: %w", err)
		}*/

		buf = bytes.Buffer{}
		err = json.NewEncoder(&buf).Encode(coins)
		if err != nil {
			return nil, fmt.Errorf("failed to encord response body: %w", err)
		}

		req.Body = io.NopCloser(&buf)
	}

	for i := 0; i < int(license.DigAllowed); i++ {
		LicenseChan <- &License{
			ID:     license.Id,
			IsLast: i == int(license.DigAllowed)-1,
		}
	}

	licenseMetricsLocker.Lock()
	coinNumLicenses[len(coins)] = append(coinNumLicenses[len(coins)], int8(license.DigAllowed))
	issueLicenseRetryNum = append(issueLicenseRetryNum, i)
	licenseMetricsLocker.Unlock()

	return &license, nil
}
