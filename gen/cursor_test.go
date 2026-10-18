package main

import "testing"

// TestCursorResponseType protects the wrapper choice for cursor endpoints:
// only a "cursor" query field typed string selects the string wrapper.
func TestCursorResponseType(t *testing.T) {
	cases := []struct {
		name  string
		want  string
		query []string
	}{
		{"int64 cursor", "CursorPaginatedResponse", []string{"path: string", "cursor: int64", "limit: int=20"}},
		{"string cursor", "StringCursorPaginatedResponse", []string{"path: string", "cursor: string", "limit: int=20"}},
		{"string field not named cursor", "CursorPaginatedResponse", []string{"token: string", "cursor: int64"}},
		{"no cursor field", "CursorPaginatedResponse", []string{"limit: int=20"}},
	}
	for _, c := range cases {
		got := cursorResponseType(Endpoint{Pagination: "cursor", Query: c.query})
		if got != c.want {
			t.Errorf("%s: cursorResponseType = %q, want %q", c.name, got, c.want)
		}
	}
}
