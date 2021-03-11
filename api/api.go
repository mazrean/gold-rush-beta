package api

import (
	"fmt"
	"net/http"
	"os"

	"github.com/mazrean/gold-rush-beta/openapi"
)

type API struct {
	*openapi.DefaultApiService
}

func NewAPI() *API {
	client := openapi.NewAPIClient(&openapi.Configuration{
		Servers: openapi.ServerConfigurations{
			{
				URL: fmt.Sprintf("http://%s:8000", os.Getenv("ADDRESS")),
			},
		},
		HTTPClient: http.DefaultClient,
		Debug:      false,
	})

	return &API{client.DefaultApi}
}
