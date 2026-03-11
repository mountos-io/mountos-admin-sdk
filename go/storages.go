package sdk

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

type StoragesService struct{ c *Client }

func (s *StoragesService) Create(ctx context.Context, req *CreateStorageRequest) (*CreateStorageResponse, error) {
	data, err := s.c.post(ctx, "/api/v1/storages/create", req)
	if err != nil {
		return nil, err
	}
	return decodeJSON[CreateStorageResponse](data)
}

func (s *StoragesService) List(ctx context.Context, opts *StorageListOptions) (*PaginatedResponse[Storage], error) {
	q := url.Values{}
	if opts != nil {
		q.Set("accountId", strconv.FormatInt(opts.AccountID, 10))
		addPagination(q, opts.Page, opts.Limit)
	}
	data, err := s.c.get(ctx, "/api/v1/storages/list?"+q.Encode())
	if err != nil {
		return nil, err
	}
	return decodeJSON[PaginatedResponse[Storage]](data)
}

func (s *StoragesService) Get(ctx context.Context, storageID string) (*Storage, error) {
	data, err := s.c.get(ctx, fmt.Sprintf("/api/v1/storages/%s", storageID))
	if err != nil {
		return nil, err
	}
	return decodeJSON[Storage](data)
}

func (s *StoragesService) Edit(ctx context.Context, storageID string, req *EditStorageRequest) (*IDResponse, error) {
	data, err := s.c.put(ctx, fmt.Sprintf("/api/v1/storages/%s/edit", storageID), req)
	if err != nil {
		return nil, err
	}
	return decodeJSON[IDResponse](data)
}

func (s *StoragesService) Activate(ctx context.Context, storageID string) error {
	_, err := s.c.post(ctx, fmt.Sprintf("/api/v1/storages/%s/activate", storageID), nil)
	return err
}

func (s *StoragesService) Deactivate(ctx context.Context, storageID string) error {
	_, err := s.c.post(ctx, fmt.Sprintf("/api/v1/storages/%s/deactivate", storageID), nil)
	return err
}
