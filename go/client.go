package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// Client is the mountOS Admin API client.
type Client struct {
	baseURL string
	http    *http.Client
	auth    *tokenCache

	Accounts     *AccountsService
	Users        *UsersService
	Regions      *RegionsService
	Storages     *StoragesService
	Volumes      *VolumesService
	AuditLogs    *AuditLogsService
	ServiceNodes *ServiceNodesService
	Discover     *DiscoverService
}

// NewClient creates a new SDK client.
func NewClient(cfg Config) (*Client, error) {
	tc, err := newTokenCache(cfg.PrivateKey)
	if err != nil {
		return nil, err
	}

	c := &Client{
		baseURL: cfg.BaseURL,
		http:    http.DefaultClient,
		auth:    tc,
	}
	c.Accounts = &AccountsService{c: c}
	c.Users = &UsersService{c: c}
	c.Regions = &RegionsService{c: c}
	c.Storages = &StoragesService{c: c}
	c.Volumes = &VolumesService{c: c}
	c.AuditLogs = &AuditLogsService{c: c}
	c.ServiceNodes = &ServiceNodesService{c: c}
	c.Discover = &DiscoverService{c: c}
	return c, nil
}

type envelope struct {
	Status    string          `json:"status"`
	Message   string          `json:"message"`
	Data      json.RawMessage `json:"data,omitempty"`
	ErrorCode int             `json:"errorCode,omitempty"`
}

func (c *Client) do(ctx context.Context, method, path string, body any) (json.RawMessage, error) {
	token, err := c.auth.getToken()
	if err != nil {
		return nil, err
	}

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("mountos: marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("mountos: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mountos: request failed: %w", err)
	}
	defer resp.Body.Close()

	var env envelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return nil, &Error{Message: resp.Status, Status: resp.StatusCode}
	}

	if env.Status != "success" {
		return nil, &Error{Message: env.Message, Status: resp.StatusCode, ErrorCode: env.ErrorCode}
	}
	return env.Data, nil
}

func (c *Client) get(ctx context.Context, path string) (json.RawMessage, error) {
	return c.do(ctx, http.MethodGet, path, nil)
}

func (c *Client) post(ctx context.Context, path string, body any) (json.RawMessage, error) {
	return c.do(ctx, http.MethodPost, path, body)
}

func (c *Client) put(ctx context.Context, path string, body any) (json.RawMessage, error) {
	return c.do(ctx, http.MethodPut, path, body)
}

func (c *Client) delete(ctx context.Context, path string) (json.RawMessage, error) {
	return c.do(ctx, http.MethodDelete, path, nil)
}

func decodeJSON[T any](data json.RawMessage) (*T, error) {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("mountos: decode data: %w", err)
	}
	return &v, nil
}

func addPagination(q url.Values, page, limit int) {
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
}
