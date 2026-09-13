package convert

import (
	"encoding/json"
	"sort"
	"strings"
)

// Field is one frontmatter entry, kept in order.
type Field struct {
	Key   string
	Value any // string, bool, int, []string, map[string]string
}

// Frontmatter renders fields as a YAML block. Strings are JSON-quoted, which is valid YAML
// and sidesteps every quoting rule.
func Frontmatter(fields []Field) string {
	var b strings.Builder
	b.WriteString("---\n")
	for _, f := range fields {
		switch v := f.Value.(type) {
		case string:
			if v == "" {
				continue
			}
			b.WriteString(f.Key + ": " + yamlString(v) + "\n")
		case bool:
			if v {
				b.WriteString(f.Key + ": true\n")
			} else {
				b.WriteString(f.Key + ": false\n")
			}
		case int:
			b.WriteString(f.Key + ": " + itoa(v) + "\n")
		case []string:
			if len(v) == 0 {
				continue
			}
			b.WriteString(f.Key + ":\n")
			for _, s := range v {
				b.WriteString("  - " + yamlString(s) + "\n")
			}
		case map[string]string:
			if len(v) == 0 {
				continue
			}
			b.WriteString(f.Key + ":\n")
			keys := make([]string, 0, len(v))
			for k := range v {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				b.WriteString("  " + yamlString(k) + ": " + yamlString(v[k]) + "\n")
			}
		}
	}
	b.WriteString("---\n")
	return b.String()
}

func yamlString(s string) string {
	out, _ := json.Marshal(s)
	return string(out)
}

func itoa(n int) string {
	out, _ := json.Marshal(n)
	return string(out)
}
