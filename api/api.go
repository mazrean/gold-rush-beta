package api

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/mazrean/gold-rush-beta/openapi"
)

var (
	client *openapi.APIClient
	api    *openapi.DefaultApiService
)

func Setup() {
	client = openapi.NewAPIClient(&openapi.Configuration{
		Servers: openapi.ServerConfigurations{
			{
				URL: fmt.Sprintf("http://%s:8000", os.Getenv("ADDRESS")),
			},
		},
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
					DualStack: true,
				}).DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				MaxConnsPerHost:       100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				DisableKeepAlives:     true,
			},
			Timeout: 60 * time.Second,
		},
		Debug: false,
	})
	api = client.DefaultApi
}
