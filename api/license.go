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

func IssueLicense(ctx context.Context, coins []int32) *openapi.License {
	atomic.AddInt64(&issueLicenseCalledNum, 1)

	var (
		i       int
		license openapi.License
		//res     *http.Response
		err error
	)
	for i = 0; ; i++ {
		startTime := time.Now()
		license, _, err = api.IssueLicense(ctx).Args(coins).Execute()
		requestTime := time.Since(startTime).Milliseconds()
		issueLicenseRequestTimeLocker.Lock()
		issueLicenseRequestTime = append(issueLicenseRequestTime, requestTime)
		issueLicenseRequestTimeLocker.Unlock()

		if err == nil {
			break
		}

		var apiErr openapi.GenericOpenAPIError
		ok := errors.As(err, &apiErr)
		if ok {
			//log.Printf("license error(%s):%+v\n", apiErr.Error(), apiErr.Model().(openapi.ModelError))
		} else {
			log.Printf("license error:%+v\n", err)
		}
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

	return &license
}
