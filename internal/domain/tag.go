package domain

import "strings"

const TagPrefix = "tag:"

func NormalizeTag(tag string) string {
	return strings.TrimPrefix(strings.ToLower(strings.TrimSpace(tag)), TagPrefix)
}

func NormalizeTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		normalized := NormalizeTag(tag)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}
