// Package client provides the official SDK for interacting with the N.O.R.E. API.
// It abstracts the underlying transport (local socket or remote HTTP) and
// handles protocol serialization for systems, instances, and executions.
package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Muhammad-Jay/neuron/application/connection"
	"github.com/Muhammad-Jay/neuron/shared/types/core"
	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
)

// Client facilitates communication with the neuron server.
// It utilizes a connection interface to allow for seamless switching
// between local socket and remote HTTP communication.
type Client struct {
	conn connection.Connection
}

// New creates a new Client using the provided Connection interface.
func New(conn connection.Connection) *Client {
	return &Client{conn: conn}
}

// NewLocal initializes a Client configured to communicate over a local Unix domain socket.
func NewLocal(socketPath string) *Client {
	return New(connection.NewLocal(socketPath))
}

// NewRemote initializes a Client configured to communicate with a remote HTTP endpoint.
func NewRemote(endpoint string) *Client {
	return New(connection.NewRemote(endpoint))
}

// Close terminates the underlying connection and releases associated resources.
func (c *Client) Close() error {
	return c.conn.Close()
}

// Health pings the server to verify it is reachable and operational.
func (c *Client) Health(ctx context.Context) error {
	return c.conn.Health(ctx)
}

// ListInstances retrieves all currently registered system instances from the server.
func (c *Client) ListInstances(ctx context.Context, queryPath string) ([]protocol.InstanceResponse, error) {
	var response struct {
		Data []protocol.InstanceResponse `json:"data"`
	}

	var path string

	if queryPath == "" || queryPath == "/" {
		path = protocol.InstancesPath
	}else {
		path = fmt.Sprintf(protocol.InstancesPath + queryPath)
	}

	if err := c.conn.Do(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

// EnsureInstance verifies the existence of a system instance by its key,
// creating a new instance on the server if it does not already exist.
func (c *Client) EnsureInstance(ctx context.Context, key protocol.InstanceKey, system *core.System) (protocol.InstanceResponse, error) {
	if system == nil {
		return protocol.InstanceResponse{}, fmt.Errorf("system is required")
	}

	var response struct {
		Data protocol.InstanceResponse `json:"data"`
	}

	req := protocol.CreateInstanceRequest{
		Key:    key,
		System: system,
	}

	if err := c.conn.Do(ctx, http.MethodPost, protocol.InstancesPath, req, &response); err != nil {
		return protocol.InstanceResponse{}, err
	}

	return response.Data, nil
}

// Execute triggers a workflow execution on a specific instance using the provided input data.
func (c *Client) Execute(ctx context.Context, instanceID string, input map[string]any) (protocol.ExecuteResponse, error) {
	instanceID = strings.TrimSpace(instanceID)
	if instanceID == "" {
		return protocol.ExecuteResponse{}, fmt.Errorf("instance ID is required")
	}

	var response struct {
		Data protocol.ExecuteResponse `json:"data"`
	}

	path := fmt.Sprintf(protocol.ExecutePath, instanceID)
	req := protocol.ExecuteRequest{
		Input: input,
	}

	if err := c.conn.Do(ctx, http.MethodPost, path, req, &response); err != nil {
		return protocol.ExecuteResponse{}, err
	}

	return response.Data, nil
}