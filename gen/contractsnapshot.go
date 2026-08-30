package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// snapshotField is one field of a snapshot type, mirroring Field exactly as
// parseField computes it - no reinterpretation of the required/optional
// sigil convention here.
type snapshotField struct {
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Optional bool   `json:"optional"`
}

type snapshotType struct {
	Fields map[string]snapshotField `json:"fields"`
}

// contractSnapshot is the flat, language-agnostic view of api.yaml's type
// and enum definitions, consumed by mountos-servers' structural drift
// checker (scripts/tools/admincontract). It carries no request/response/
// resource shaping - just the raw Types/Enums maps already on Spec.
type contractSnapshot struct {
	Version string                  `json:"version"`
	Types   map[string]snapshotType `json:"types"`
	Enums   map[string][]string     `json:"enums"`
}

// generateContractSnapshot walks every type in spec.Types and every enum in
// spec.Enums and writes one flat JSON snapshot to outPath, for a downstream
// consumer (mountos-servers' admincontract tool) to diff hand-written Go
// structs against. Field required/optional/type come straight from
// parseField - this function only serializes, it never reinterprets them.
func generateContractSnapshot(spec *Spec, outPath string) error {
	snap := contractSnapshot{
		Version: spec.Version,
		Types:   make(map[string]snapshotType, len(spec.Types)),
		Enums:   make(map[string][]string, len(spec.Enums)),
	}

	for typeName, fields := range spec.Types {
		st := snapshotType{Fields: make(map[string]snapshotField, len(fields))}
		for _, s := range fields {
			f := parseField(s)
			st.Fields[f.Name] = snapshotField{
				Type:     f.Type,
				Required: f.Required,
				Optional: f.Optional,
			}
		}
		snap.Types[typeName] = st
	}

	for enumName, values := range spec.Enums {
		sorted := make([]string, len(values))
		copy(sorted, values)
		snap.Enums[enumName] = sorted
	}

	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal contract snapshot: %w", err)
	}
	data = append(data, '\n')

	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(outPath), err)
	}
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}
	return nil
}
