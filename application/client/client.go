package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Muhammad-Jay/neuron/application/connection"

	"github.com/Muhammad-Jay/neuron/shared/types/core"
)

type Client struct {
	conn connection.Connection
}

func New(conn connection.Connection) *Client {
	return &Client{conn: conn}
}

func NewLocal(socketPath string) *Client {
	return New(connection.NewLocal(socketPath))
}

func NewRemote(endpoint string) *Client {
	return New(connection.NewRemote(endpoint))
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Health(ctx context.Context) error {
	return c.conn.Health(ctx)
}

func (c *Client) ListInstances(ctx context.Context) ([]connection.InstanceResponse, error) {
	var response struct {
		Data []connection.InstanceResponse `json:"data"`
	}
	if err := c.conn.Do(ctx, http.MethodGet, connection.InstancesPath, nil, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}


func (c *Client) EnsureInstance(ctx context.Context, key connection.InstanceKey, system *core.System) (connection.InstanceResponse, error) {
	if system == nil {
		return connection.InstanceResponse{}, fmt.Errorf("system is required")
	}

	var response struct {
		Data connection.InstanceResponse `json:"data"`
	}
	req := connection.CreateInstanceRequest{Key: key, System: system}

	if err := c.conn.Do(ctx, http.MethodPost, connection.InstancesPath, req, &response); err != nil {
		return connection.InstanceResponse{}, err
	}
	return response.Data, nil
}

func (c *Client) Execute(ctx context.Context, instanceID string, input map[string]any) (connection.ExecuteResponse, error) {
	instanceID = strings.TrimSpace(instanceID)
	if instanceID == "" {
		return connection.ExecuteResponse{}, fmt.Errorf("instance ID is required")
	}

	var response struct {
		Data connection.ExecuteResponse `json:"data"`
	}

	path := fmt.Sprintf(connection.ExecutePath, instanceID)
	if err := c.conn.Do(ctx, http.MethodPost, path, connection.ExecuteRequest{Input: input}, &response); err != nil {
		return connection.ExecuteResponse{}, err
	}
	return response.Data, nil
}