package sdk

import (
	"context"
	"fmt"
	"net/url"
)

type AccountsService struct{ c *Client }

func (s *AccountsService) Create(ctx context.Context, req *CreateAccountRequest) (*IDResponse, error) {
	data, err := s.c.post(ctx, "/api/v1/accounts/create", req)
	if err != nil {
		return nil, err
	}
	return decodeJSON[IDResponse](data)
}

func (s *AccountsService) List(ctx context.Context, opts *ListOptions) (*PaginatedResponse[Account], error) {
	q := url.Values{}
	if opts != nil {
		addPagination(q, opts.Page, opts.Limit)
	}
	path := "/api/v1/accounts/list"
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	data, err := s.c.get(ctx, path)
	if err != nil {
		return nil, err
	}
	return decodeJSON[PaginatedResponse[Account]](data)
}

func (s *AccountsService) Get(ctx context.Context, accountID string) (*Account, error) {
	data, err := s.c.get(ctx, fmt.Sprintf("/api/v1/accounts/%s", accountID))
	if err != nil {
		return nil, err
	}
	return decodeJSON[Account](data)
}

func (s *AccountsService) Edit(ctx context.Context, accountID string, req *EditAccountRequest) (*IDResponse, error) {
	data, err := s.c.put(ctx, fmt.Sprintf("/api/v1/accounts/%s/edit", accountID), req)
	if err != nil {
		return nil, err
	}
	return decodeJSON[IDResponse](data)
}

func (s *AccountsService) Lock(ctx context.Context, accountID string) error {
	_, err := s.c.post(ctx, fmt.Sprintf("/api/v1/accounts/%s/lock", accountID), nil)
	return err
}

func (s *AccountsService) Unlock(ctx context.Context, accountID string) error {
	_, err := s.c.post(ctx, fmt.Sprintf("/api/v1/accounts/%s/unlock", accountID), nil)
	return err
}

func (s *AccountsService) Activate(ctx context.Context, accountID string) error {
	_, err := s.c.post(ctx, fmt.Sprintf("/api/v1/accounts/%s/activate", accountID), nil)
	return err
}

func (s *AccountsService) Deactivate(ctx context.Context, accountID string) error {
	_, err := s.c.post(ctx, fmt.Sprintf("/api/v1/accounts/%s/deactivate", accountID), nil)
	return err
}
