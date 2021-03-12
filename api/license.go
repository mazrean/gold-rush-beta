package api

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mazrean/gold-rush-beta/openapi"
)

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

	LicenseChan = make(chan int32, 100)
)

func IssueLicense(ctx context.Context, coins []int32) *openapi.License {
	atomic.AddInt64(&issueLicenseCalledNum, 1)

	var (
		i       int
		license openapi.License
		res     *http.Response
		err     error
	)
	for i = 0; ; i++ {
		startTime := time.Now()
		license, res, err = api.IssueLicense(ctx).Args(coins).Execute()
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
		if res != nil && res.StatusCode == 409 {
			licenseList, _, err := api.ListLicenses(ctx).Execute()
			if err != nil {
				var apiErr openapi.GenericOpenAPIError
				ok := errors.As(err, &apiErr)
				if ok {
					log.Printf("get license error(%s):%+v\n", apiErr.Error(), apiErr.Model().(openapi.ModelError))
				} else {
					log.Printf("get license error:%+v\n", err)
				}
			} else {
				log.Printf("licenses: %+v\n", licenseList)
			}
		}
	}

	for i := 0; i < int(license.DigAllowed); i++ {
		LicenseChan <- license.Id
	}

	licenseMetricsLocker.Lock()
	coinNumLicenses[len(coins)] = append(coinNumLicenses[len(coins)], int8(license.DigAllowed))
	issueLicenseRetryNum = append(issueLicenseRetryNum, i)
	licenseMetricsLocker.Unlock()

	return &license
}
