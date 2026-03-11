package sdk

import "context"

// CacheService provides cache management operations.
type CacheService struct{ c *Client }

// Refresh triggers a service verifier cache refresh on appserv
// and broadcasts to all registered service nodes.
func (s *CacheService) Refresh(ctx context.Context) error {
	_, err := s.c.post(ctx, "/api/v1/cache/refresh", nil)
	return err
}
