package sdk

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)


type UsersService struct{ c *Client }

func (s *UsersService) Add(ctx context.Context, req *AddUserRequest) (*IDResponse, error) {
	data, err := s.c.post(ctx, "/api/v1/users/add", req)
	if err != nil {
		return nil, err
	}
	return decodeJSON[IDResponse](data)
}

func (s *UsersService) List(ctx context.Context, opts *UserListOptions) (*PaginatedResponse[User], error) {
	q := url.Values{}
	if opts != nil {
		q.Set("accountId", strconv.FormatInt(opts.AccountID, 10))
		addPagination(q, opts.Page, opts.Limit)
	}
	data, err := s.c.get(ctx, "/api/v1/users/list?"+q.Encode())
	if err != nil {
		return nil, err
	}
	return decodeJSON[PaginatedResponse[User]](data)
}

func (s *UsersService) Get(ctx context.Context, userID string) (*User, error) {
	data, err := s.c.get(ctx, fmt.Sprintf("/api/v1/users/%s", userID))
	if err != nil {
		return nil, err
	}
	return decodeJSON[User](data)
}

func (s *UsersService) Edit(ctx context.Context, userID string, req *EditUserRequest) (*IDResponse, error) {
	data, err := s.c.put(ctx, fmt.Sprintf("/api/v1/users/%s/edit", userID), req)
	if err != nil {
		return nil, err
	}
	return decodeJSON[IDResponse](data)
}

func (s *UsersService) Activate(ctx context.Context, userID string) error {
	_, err := s.c.post(ctx, fmt.Sprintf("/api/v1/users/%s/activate", userID), nil)
	return err
}

func (s *UsersService) Deactivate(ctx context.Context, userID string) error {
	_, err := s.c.post(ctx, fmt.Sprintf("/api/v1/users/%s/deactivate", userID), nil)
	return err
}
