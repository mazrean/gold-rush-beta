package api

import (
	"fmt"
	"net/http"
	"os"

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
		HTTPClient: http.DefaultClient,
		Debug:      false,
	})
	api = client.DefaultApi
}