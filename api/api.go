package api

import (
	"fmt"
	"net/http"
	"os"
)

var (
	baseURL = fmt.Sprintf("http://%s:8000", os.Getenv("ADDRESS"))
	client  = http.DefaultClient
)
