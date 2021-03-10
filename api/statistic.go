package api

import (
	"fmt"
	"strings"
)

func Statistic(sb strings.Builder) {
	var aveTreasureNum float64 = 0
	for _, treasureNum := range digTreasureNum {
		aveTreasureNum += float64(treasureNum)
	}
	aveTreasureNum /= float64(len(digTreasureNum))

	digReqTimes := [10]float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	for i, requestTimes := range digRequestTime {
		for _, requestTime := range requestTimes {
			digReqTimes[i] += float64(requestTime)
		}
		digReqTimes[i] /= float64(len(requestTimes))
	}

	var digRetry float64 = 0
	for _, retry := range digRetryNum {
		digRetry += float64(retry)
	}
	digRetry /= float64(len(digRetryNum))

	var exploreReqTime float64 = 0
	for _, requestTime := range exploreRequestTime {
		exploreReqTime += float64(requestTime)
	}
	exploreReqTime /= float64(len(exploreRequestTime))

	var exploreRetry float64 = 0
	for _, retry := range exploreRetryNum {
		exploreRetry += float64(retry)
	}
	exploreRetry /= float64(len(exploreRetryNum))

	var licenseReqTime float64 = 0
	for _, requestTime := range issueLicenseRequestTime {
		licenseReqTime += float64(requestTime)
	}
	licenseReqTime /= float64(len(issueLicenseRequestTime))

	var licenseRetry float64 = 0
	for _, retry := range issueLicenseRetryNum {
		licenseRetry += float64(retry)
	}
	licenseRetry /= float64(len(issueLicenseRetryNum))

	coinLicenses := [11]float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	for i, allows := range coinNumLicenses {
		for _, allow := range allows {
			coinLicenses[i] += float64(allow)
		}
		coinLicenses[i] /= float64(len(allows))
	}

	var cashReqTime float64 = 0
	for _, requestTime := range cashRequestTime {
		cashReqTime += float64(requestTime)
	}
	cashReqTime /= float64(len(cashRequestTime))

	var cashRetry float64 = 0
	for _, retry := range cashRetryNum {
		cashRetry += float64(retry)
	}
	cashRetry /= float64(len(cashRetryNum))

	var aveCoinNum float64 = 0
	for _, coin := range coinNum {
		aveCoinNum += float64(coin)
	}
	aveCoinNum /= float64(len(coinNum))

	sb.WriteString(fmt.Sprintf(`api metrics:
	dig:
		called num:%d
		retry num:%g
		treasure num ave:%g
		depth-request time: %g,%g,%g,%g,%g,%g,%g,%g,%g,%g
	explore:
		called num:%d
		retry num:%g
		request time:%g
	license:
		called num:%d
		retry num:%g
		request time:%g
		coin num-allow: %g,%g,%g,%g,%g,%g,%g,%g,%g,%g,%g
	cash:
		called num:%d
		retry num:%g
		request time:%g
		coin num ave:%g
`,
		digCalledNum,
		digRetry,
		aveTreasureNum,
		digReqTimes[0],
		digReqTimes[1],
		digReqTimes[2],
		digReqTimes[3],
		digReqTimes[4],
		digReqTimes[5],
		digReqTimes[6],
		digReqTimes[7],
		digReqTimes[8],
		digReqTimes[9],
		exploreCalledNum,
		exploreRetry,
		exploreReqTime,
		issueLicenseCalledNum,
		licenseRetry,
		licenseReqTime,
		coinLicenses[0],
		coinLicenses[1],
		coinLicenses[2],
		coinLicenses[3],
		coinLicenses[4],
		coinLicenses[5],
		coinLicenses[6],
		coinLicenses[7],
		coinLicenses[8],
		coinLicenses[9],
		coinLicenses[10],
		cashCalledNum,
		cashRetry,
		cashReqTime,
		aveCoinNum))
}
