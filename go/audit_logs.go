package sdk

import (
	"context"
	"net/url"
	"strconv"
)

type AuditLogsService struct{ c *Client }

func (s *AuditLogsService) List(ctx context.Context, opts *AuditLogListOptions) (*CursorPaginatedResponse[AuditLog], error) {
	q := url.Values{}
	if opts != nil {
		if opts.AccountID > 0 {
			q.Set("accountId", strconv.FormatInt(opts.AccountID, 10))
		}
		if opts.Cursor > 0 {
			q.Set("cursor", strconv.FormatInt(opts.Cursor, 10))
		}
		if opts.Limit > 0 {
			q.Set("limit", strconv.Itoa(opts.Limit))
		}
		if opts.Subject != "" {
			q.Set("subject", opts.Subject)
		}
	}
	path := "/api/v1/audit-logs/list"
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	data, err := s.c.get(ctx, path)
	if err != nil {
		return nil, err
	}
	return decodeJSON[CursorPaginatedResponse[AuditLog]](data)
}
