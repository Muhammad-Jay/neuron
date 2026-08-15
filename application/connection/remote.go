package connection

import (
	"net/http"
	"strings"
	"time"
)

func NewRemote(endpoint string) Connection {
	endpoint = strings.TrimRight(endpoint, "/")
	return New(NewHTTPTransport(&http.Client{Timeout: 60 * time.Second}, endpoint))
}
