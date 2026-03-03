package sdk

import (
	"context"
	"fmt"
	"net/url"
)

type RegionsService struct{ c *Client }

func (s *RegionsService) Create(ctx context.Context, req *CreateRegionRequest) (*IDResponse, error) {
	data, err := s.c.post(ctx, "/api/v1/regions/create", req)
	if err != nil {
		return nil, err
	}
	return decodeJSON[IDResponse](data)
}

func (s *RegionsService) List(ctx context.Context, opts *ListOptions) (*PaginatedResponse[Region], error) {
	q := url.Values{}
	if opts != nil {
		addPagination(q, opts.Page, opts.Limit)
	}
	path := "/api/v1/regions/list"
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	data, err := s.c.get(ctx, path)
	if err != nil {
		return nil, err
	}
	return decodeJSON[PaginatedResponse[Region]](data)
}

func (s *RegionsService) Get(ctx context.Context, regionID int64) (*Region, error) {
	data, err := s.c.get(ctx, fmt.Sprintf("/api/v1/regions/%d", regionID))
	if err != nil {
		return nil, err
	}
	return decodeJSON[Region](data)
}

func (s *RegionsService) Edit(ctx context.Context, regionID int64, req *EditRegionRequest) (*IDResponse, error) {
	data, err := s.c.put(ctx, fmt.Sprintf("/api/v1/regions/%d/edit", regionID), req)
	if err != nil {
		return nil, err
	}
	return decodeJSON[IDResponse](data)
}

func (s *RegionsService) Activate(ctx context.Context, regionID int64) error {
	_, err := s.c.post(ctx, fmt.Sprintf("/api/v1/regions/%d/activate", regionID), nil)
	return err
}

func (s *RegionsService) Deactivate(ctx context.Context, regionID int64) error {
	_, err := s.c.post(ctx, fmt.Sprintf("/api/v1/regions/%d/deactivate", regionID), nil)
	return err
}
