package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestRustInputStructsAreConstructible fails when a non_exhaustive Rust struct
// that a caller can build lacks a way to build it. Outside the crate a
// non_exhaustive struct has no literal, so each input struct needs Default or
// new(required...), plus a with_<field> setter for every optional field.
func TestRustInputStructsAreConstructible(t *testing.T) {
	spec := loadSpec("../api.yaml")
	dir := t.TempDir()
	generateRust(spec, dir)
	raw, err := os.ReadFile(filepath.Join(dir, "types_gen.rs"))
	if err != nil {
		t.Fatalf("read generated Rust: %v", err)
	}
	src := string(raw)

	inputs := rustInputTypes(spec)
	inputs["Config"] = true
	for _, res := range spec.Resources {
		for _, ep := range res.Endpoints {
			if len(ep.Request) > 0 {
				inputs[requestTypeName(res.Name, ep.Action)] = true
			}
			if ep.Pagination == "cursor" || (ep.Pagination == "page" && hasExtraQueryParam(ep.Query)) {
				inputs[listOptionsTypeName(res.Name)] = true
			}
		}
	}
	for _, must := range []string{"VolumeRetentionPolicy", "VolumeVersioningPolicy", "DashboardUser", "EditVolumeRequest"} {
		if !inputs[must] {
			t.Errorf("%s is not classified as an input struct", must)
		}
	}

	for name := range inputs {
		structRe := regexp.MustCompile(`(?s)#\[derive\(([^)]*)\)\]\n#\[non_exhaustive\]\npub struct ` + name + ` \{(.*?)\n\}\n`)
		m := structRe.FindStringSubmatch(src)
		if m == nil {
			t.Errorf("%s: struct not found or not non_exhaustive", name)
			continue
		}
		implRe := regexp.MustCompile(`(?s)\nimpl ` + name + ` \{(.*?)\n\}\n`)
		impl := ""
		if im := implRe.FindStringSubmatch(src); im != nil {
			impl = im[1]
		}
		required, optional := 0, []string{}
		for _, line := range strings.Split(m[2], "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "pub ") {
				continue
			}
			field, ty, _ := strings.Cut(strings.TrimPrefix(line, "pub "), ": ")
			if strings.HasPrefix(ty, "Option<") {
				optional = append(optional, strings.TrimPrefix(field, "r#"))
			} else {
				required++
			}
		}
		hasDefault := strings.Contains(m[1], "Default")
		if !hasDefault && !strings.Contains(impl, "pub fn new(") {
			t.Errorf("%s: input struct has neither Default nor new()", name)
		}
		if required > 0 && !strings.Contains(impl, "pub fn new(") {
			t.Errorf("%s: required fields but no new()", name)
		}
		for _, f := range optional {
			if !strings.Contains(impl, "pub fn with_"+f+"(") {
				t.Errorf("%s: optional field %s has no with_%s setter", name, f, f)
			}
		}
	}
}
