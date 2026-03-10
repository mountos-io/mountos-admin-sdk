package sdk

import (
	"context"
	"fmt"
)

type ServiceNodesService struct{ c *Client }

func (s *ServiceNodesService) List(ctx context.Context, regionID int64) ([]ServiceNode, error) {
	data, err := s.c.get(ctx, fmt.Sprintf("/api/v1/regions/%d/nodes", regionID))
	if err != nil {
		return nil, err
	}
	result, err := decodeJSON[[]ServiceNode](data)
	if err != nil {
		return nil, err
	}
	return *result, nil
}

func (s *ServiceNodesService) Drain(ctx context.Context, regionID int64, nodeID string) error {
	_, err := s.c.post(ctx, fmt.Sprintf("/api/v1/regions/%d/nodes/%s/drain", regionID, nodeID), nil)
	return err
}

func (s *ServiceNodesService) Activate(ctx context.Context, regionID int64, nodeID string) error {
	_, err := s.c.post(ctx, fmt.Sprintf("/api/v1/regions/%d/nodes/%s/activate", regionID, nodeID), nil)
	return err
}

func (s *ServiceNodesService) Remove(ctx context.Context, regionID int64, nodeID string) error {
	_, err := s.c.delete(ctx, fmt.Sprintf("/api/v1/regions/%d/nodes/%s", regionID, nodeID))
	return err
}
