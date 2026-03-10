package sdk

import (
	"context"
	"net/url"
)

type DiscoverService struct{ c *Client }

func (s *DiscoverService) Meta(ctx context.Context, accessKeyID string) (*DiscoverMetaResponse, error) {
	q := url.Values{}
	q.Set("access_key_id", accessKeyID)
	data, err := s.c.get(ctx, "/api/v1/discover/meta?"+q.Encode())
	if err != nil {
		return nil, err
	}
	return decodeJSON[DiscoverMetaResponse](data)
}
