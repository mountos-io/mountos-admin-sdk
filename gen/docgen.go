package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func generateDoc(spec *Spec, outDir string) {
	var w strings.Builder

	w.WriteString("# mountOS Admin API Reference\n\n")
	fmt.Fprintf(&w, "Base path: `%s`\n", spec.BasePath)
	fmt.Fprintf(&w, "Auth: `Authorization: Bearer <JWT>` (%s/%s, sub=%s, aud=%s)\n\n",
		"ED25519", spec.JWT.Algorithm, spec.JWT.Subject, spec.JWT.Audience)

	// Response Envelope
	w.WriteString("## Response Envelope\n\n")
	w.WriteString("All responses use `StandardResponse`:\n")
	w.WriteString("```\n")
	w.WriteString("{ \"status\": \"success\"|\"failure\", \"message\": string, \"data\"?: object, \"errorCode\"?: int }\n")
	w.WriteString("```\n\n")
	w.WriteString("Paginated responses nest in `data`:\n")
	w.WriteString("```\n")
	w.WriteString("{ \"items\": T[], \"pagination\": { \"page\": int, \"limit\": int, \"total\": int64, \"totalPages\": int64 } }\n")
	w.WriteString("```\n\n")
	w.WriteString("Cursor-paginated responses nest in `data`:\n")
	w.WriteString("```\n")
	w.WriteString("{ \"items\": T[], \"nextCursor\": int64|null }\n")
	w.WriteString("```\n\n")

	// Error Codes
	w.WriteString("## Error Codes (AppServ 1XXXX)\n\n")
	w.WriteString("| Code  | Name                   |\n")
	w.WriteString("|-------|------------------------|\n")
	for _, ec := range spec.ErrorCodes {
		fmt.Fprintf(&w, "| %d | %-22s |\n", ec.Code, ec.Name)
	}
	w.WriteString("\n---\n")

	// Resources
	for _, res := range spec.Resources {
		w.WriteString("\n## " + res.Name + "\n")
		fullBasePath := spec.BasePath + res.Path
		for _, ep := range res.Endpoints {
			fullPath := fullBasePath + ep.Path
			fullPath = strings.TrimSuffix(fullPath, "/")
			w.WriteString("\n")
			fmt.Fprintf(&w, "### %s %s\n", ep.Method, fullPath)

			// Path params
			allParams := extractPathParams(fullPath)
			if len(allParams) > 0 {
				for _, p := range allParams {
					name := strings.TrimPrefix(p, ":")
					fmt.Fprintf(&w, "Param: `%s`\n", name)
				}
			}

			// Query params
			if len(ep.Query) > 0 {
				var parts []string
				for _, qs := range ep.Query {
					f := parseField(qs)
					desc := fmt.Sprintf("`%s=%s", f.Name, docType(f.Type))
					if f.Required {
						desc += "(required)"
					}
					if f.Default != "" {
						desc += fmt.Sprintf("(default %s)", f.Default)
					}
					desc += "`"
					parts = append(parts, desc)
				}
				fmt.Fprintf(&w, "Query: %s\n", strings.Join(parts, ", "))
			}

			// Request body
			if len(ep.Request) > 0 {
				w.WriteString("Request:\n```\n{\n")
				for i, rs := range ep.Request {
					f := parseField(rs)
					line := fmt.Sprintf("  \"%s\"", f.Name)
					if !f.Required {
						line += "?"
					}
					line += ": " + docType(f.Type)
					if f.Required {
						line += "(required)"
					}
					if i < len(ep.Request)-1 {
						line += ","
					}
					w.WriteString(line + "\n")
				}
				w.WriteString("}\n```\n")
			}

			// Response
			if len(ep.Response) > 0 {
				parts := make([]string, len(ep.Response))
				for i, rs := range ep.Response {
					f := parseField(rs)
					parts[i] = fmt.Sprintf("\"%s\": %s", f.Name, docType(f.Type))
				}
				fmt.Fprintf(&w, "Response data: `{ %s }`\n", strings.Join(parts, ", "))
			} else if ep.ResponseType != "" {
				if ep.ResponseArray {
					fmt.Fprintf(&w, "Response data: `%s[]`\n", ep.ResponseType)
				} else if ep.Pagination == "page" {
					fmt.Fprintf(&w, "Response data: `{ \"items\": %s[], \"pagination\": PaginationMeta }`\n", ep.ResponseType)
				} else if ep.Pagination == "cursor" {
					fmt.Fprintf(&w, "Response data: `{ \"items\": %s[], \"nextCursor\": int64|null }`\n", ep.ResponseType)
				} else {
					fmt.Fprintf(&w, "Response data: `%s`\n", ep.ResponseType)
				}
			}
		}

		// Type definition if resource has a responseType referencing a type
		typeName := findResourceType(res)
		if typeName != "" {
			if fields, ok := spec.Types[typeName]; ok {
				w.WriteString("\n### " + typeName + " Type\n")
				w.WriteString("```\n{\n")
				for i, fs := range fields {
					f := parseField(fs)
					line := "  \"" + f.Name + "\""
					if f.Optional {
						line += "?"
					}
					line += ": " + docFieldType(f.Type)
					if i < len(fields)-1 {
						line += ","
					}
					w.WriteString(line + "\n")
				}
				w.WriteString("}\n```\n")
			}
		}
		w.WriteString("\n---\n")
	}

	// JWT Construction
	w.WriteString("\n## JWT Construction\n\n")
	w.WriteString("```\n")
	w.WriteString("Header:  { \"alg\": \"" + spec.JWT.Algorithm + "\", \"typ\": \"JWT\" }\n")
	w.WriteString("Payload: {\n")
	fmt.Fprintf(&w, "  \"sub\": \"%s\",\n", spec.JWT.Subject)
	fmt.Fprintf(&w, "  \"aud\": [\"%s\"],\n", spec.JWT.Audience)
	w.WriteString("  \"iat\": unix_now,\n")
	w.WriteString("  \"nbf\": unix_now,\n")
	w.WriteString("  \"exp\": unix_now + 3600,\n")
	w.WriteString("  \"jti\": \"<nanosecond_timestamp_string>\",\n")
	fmt.Fprintf(&w, "  \"scope\": \"%s\",\n", spec.JWT.Scope)
	if kfp, ok := spec.JWT.Claims["kfp"]; ok {
		fmt.Fprintf(&w, "  \"kfp\": \"<%s>\"\n", kfp)
	}
	w.WriteString("}\n")
	w.WriteString("Signature: ED25519 sign(header.payload, privateKey)\n")
	w.WriteString("```\n\n")
	fmt.Fprintf(&w, "Key format: %s.\n\n", spec.JWT.KeyFormat)

	// PaginationMeta
	w.WriteString("## PaginationMeta Type\n")
	w.WriteString("```\n")
	w.WriteString("{ \"page\": int, \"limit\": int, \"total\": int64, \"totalPages\": int64 }\n")
	w.WriteString("```\n")

	writeFile(filepath.Join(outDir, "api.md"), w.String())
}

func docType(t string) string {
	switch t {
	case "string", "datetime":
		return "string"
	case "int64":
		return "int64"
	case "int32":
		return "int32"
	case "int":
		return "int"
	case "bool":
		return "bool"
	case "object":
		return "object"
	case "json":
		return "object"
	default:
		return t
	}
}

func docFieldType(t string) string {
	switch t {
	case "datetime":
		return "RFC3339"
	case "json":
		return "object"
	default:
		return docType(t)
	}
}

func findResourceType(res Resource) string {
	for _, ep := range res.Endpoints {
		if ep.ResponseType != "" && ep.Pagination != "" {
			return ep.ResponseType
		}
		if ep.ResponseType != "" && ep.Action == "get" {
			return ep.ResponseType
		}
	}
	// Check array response
	for _, ep := range res.Endpoints {
		if ep.ResponseArray && ep.ResponseType != "" {
			return ep.ResponseType
		}
	}
	return ""
}
