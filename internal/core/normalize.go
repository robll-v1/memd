package core

import (
	"regexp"
	"strings"
)

var whitespaceRE = regexp.MustCompile(`\s+`)

func NormalizeContent(input string) string {
	text := strings.ReplaceAll(input, "```", " ")
	text = strings.ReplaceAll(text, "`", "")
	text = strings.ToLower(text)
	text = whitespaceRE.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

func TokenizeNormalized(input string) []string {
	if input == "" {
		return nil
	}
	parts := strings.Fields(NormalizeContent(input))
	seen := make(map[string]struct{}, len(parts))
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if len(part) < 2 {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		result = append(result, part)
	}
	return result
}
