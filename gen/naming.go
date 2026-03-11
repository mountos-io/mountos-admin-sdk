package main

import (
	"strings"
	"unicode"
)

var abbreviations = map[string]string{
	"id":   "ID",
	"url":  "URL",
	"http": "HTTP",
	"api":  "API",
	"ip":   "IP",
	"dns":  "DNS",
	"uri":  "URI",
	"uuid": "UUID",
	"sql":  "SQL",
	"ssh":  "SSH",
	"tcp":  "TCP",
	"udp":  "UDP",
	"jwt":  "JWT",
	"tls":  "TLS",
	"ssl":  "SSL",
}

// splitWords splits on camelCase boundaries, hyphens, underscores.
func splitWords(s string) []string {
	var words []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			words = append(words, cur.String())
			cur.Reset()
		}
	}
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '-' || r == '_' {
			flush()
			continue
		}
		if i > 0 && unicode.IsUpper(r) {
			// split before uppercase if previous is lowercase
			if unicode.IsLower(runes[i-1]) {
				flush()
			} else if i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
				// e.g. "HTTPAddr" → "HTTP", "Addr"
				flush()
			}
		}
		cur.WriteRune(r)
	}
	flush()
	return words
}

func pascalCase(s string) string {
	words := splitWords(s)
	if len(words) == 0 {
		return s
	}
	var b strings.Builder
	for _, w := range words {
		lower := strings.ToLower(w)
		if abbr, ok := abbreviations[lower]; ok {
			b.WriteString(abbr)
		} else {
			b.WriteString(strings.ToUpper(w[:1]) + w[1:])
		}
	}
	return b.String()
}

func camelCase(s string) string {
	words := splitWords(s)
	if len(words) == 0 {
		return s
	}
	var b strings.Builder
	for i, w := range words {
		lower := strings.ToLower(w)
		if i == 0 {
			// first word: if it's an abbreviation, lowercase it
			if _, ok := abbreviations[lower]; ok {
				b.WriteString(lower)
			} else {
				b.WriteString(strings.ToLower(w[:1]) + w[1:])
			}
		} else {
			if abbr, ok := abbreviations[lower]; ok {
				b.WriteString(abbr)
			} else {
				b.WriteString(strings.ToUpper(w[:1]) + w[1:])
			}
		}
	}
	return b.String()
}

func goFieldName(jsonName string) string {
	return pascalCase(jsonName)
}

func goParamName(pathParam string) string {
	name := strings.TrimPrefix(pathParam, ":")
	return camelCase(name)
}

func slugify(name string) string {
	return camelCase(name)
}
