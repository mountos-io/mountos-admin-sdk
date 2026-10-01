package main

import "testing"

// TestCursorResponseType protects the wrapper choice for cursor endpoints:
// only a "cursor" query field typed string selects the string wrapper.
func TestCursorResponseType(t *testing.T) {
	cases := []struct {
		name  string
		query []string
		want  string
	}{
		{"int64 cursor", []string{"path: string", "cursor: int64", "limit: int=20"}, "CursorPaginatedResponse"},
		{"string cursor", []string{"path: string", "cursor: string", "limit: int=20"}, "StringCursorPaginatedResponse"},
		{"string field not named cursor", []string{"token: string", "cursor: int64"}, "CursorPaginatedResponse"},
		{"no cursor field", []string{"limit: int=20"}, "CursorPaginatedResponse"},
	}
	for _, c := range cases {
		got := cursorResponseType(Endpoint{Pagination: "cursor", Query: c.query})
		if got != c.want {
			t.Errorf("%s: cursorResponseType = %q, want %q", c.name, got, c.want)
		}
	}
}
