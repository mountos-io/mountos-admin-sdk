package main

import (
	"strings"
	"testing"
)

// TestWriteGoStructOptionalWrapping protects goTag/writeGoStruct's
// pointer-wrapping decision (R3-010): a model/response field marked "?"
// must pointer-wrap the same scalar Go types the request-field branch
// already wraps, while a required or plain field, and a slice/map/enum/
// named type, must not.
func TestWriteGoStructOptionalWrapping(t *testing.T) {
	fields := []string{
		"count: int32?",     // optional scalar model field -> pointer + omitempty
		"total: int64!",     // required scalar -> value type, no omitempty
		"label: string",     // plain (non-optional, non-required) -> value type
		"tags: string[]?",   // optional slice -> no double pointer, just omitempty
		"meta: object?",     // optional map -> no pointer, just omitempty
		"state: PairState?", // optional named/enum type -> no pointer, just omitempty
	}

	var w strings.Builder
	writeGoStruct(&w, "Sample", fields, false)
	got := w.String()

	cases := []struct {
		name string
		want string
	}{
		{"count", "Count *int32 `json:\"count,omitempty\"`"},
		{"total", "Total int64 `json:\"total\"`"},
		{"label", "Label string `json:\"label\"`"},
		{"tags", "Tags []string `json:\"tags,omitempty\"`"},
		{"meta", "Meta map[string]any `json:\"meta,omitempty\"`"},
		{"state", "State PairState `json:\"state,omitempty\"`"},
	}
	for _, c := range cases {
		if !strings.Contains(collapseSpace(got), collapseSpace(c.want)) {
			t.Errorf("field %s: generated struct missing %q, got:\n%s", c.name, c.want, got)
		}
	}
}

// TestWriteGoStructRequestOptionalUnchanged proves the fix is additive: the
// request-field branch's existing pointer-wrap behavior (isRequest &&
// !f.Required) is unaffected by gating model fields on f.Optional.
func TestWriteGoStructRequestOptionalUnchanged(t *testing.T) {
	fields := []string{
		"count: int32",  // bare, non-required request field -> still pointer-wrapped
		"total: int64!", // required request field -> value type
	}

	var w strings.Builder
	writeGoStruct(&w, "SampleRequest", fields, true)
	got := collapseSpace(w.String())

	if !strings.Contains(got, collapseSpace("Count *int32 `json:\"count,omitempty\"`")) {
		t.Errorf("request optional field not pointer-wrapped, got:\n%s", w.String())
	}
	if !strings.Contains(got, collapseSpace("Total int64 `json:\"total\"`")) {
		t.Errorf("request required field unexpectedly wrapped, got:\n%s", w.String())
	}
}

// collapseSpace normalizes writeGoStruct's column-aligned whitespace (which
// pads names/types with variable runs of spaces) so test expectations don't
// have to hardcode alignment widths.
func collapseSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
