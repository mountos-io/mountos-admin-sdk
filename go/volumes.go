package sdk

import (
	"context"
	"fmt"
)

type VolumesService struct{ c *Client }

func (s *VolumesService) UpdateQuota(ctx context.Context, volumeID string, req *UpdateVolumeQuotaRequest) (*StringIDResponse, error) {
	data, err := s.c.put(ctx, fmt.Sprintf("/api/v1/volumes/%s/quota", volumeID), req)
	if err != nil {
		return nil, err
	}
	return decodeJSON[StringIDResponse](data)
}
