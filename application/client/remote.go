package client

import "strings"

func NewEndpoint(endpoint string) *Client {
	return NewRemote(strings.TrimRight(endpoint, "/"))
}