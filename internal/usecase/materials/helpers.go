package materials

import (
	"regexp"
	"strconv"
	"strings"
)

const minWeekMatchParts = 2

var weekPattern = regexp.MustCompile(`(?i)(?:неделя|week)\s*(\d+)`)

func matchesWeek(themeName string, week int) bool {
	matches := weekPattern.FindStringSubmatch(themeName)
	if len(matches) < minWeekMatchParts {
		return false
	}
	n, err := strconv.Atoi(matches[1])
	return err == nil && n == week
}

var linkPattern = regexp.MustCompile(`href=\\"([^"\\]+)\\"`)

func extractLinks(viewContent string) []string {
	matches := linkPattern.FindAllStringSubmatch(viewContent, -1)
	var links []string
	seen := make(map[string]bool)
	for _, m := range matches {
		link := m[1]
		// Skip internal CU links and anchors.
		if strings.HasPrefix(link, "#") || strings.Contains(link, "my.centraluniversity.ru") {
			continue
		}
		if !seen[link] {
			seen[link] = true
			links = append(links, link)
		}
	}
	return links
}

func sanitizeFilename(name string) string {
	replacements := map[rune]rune{
		'/': '-', '\\': '-', ':': '-', '*': '-',
		'?': '-', '"': '-', '<': '-', '>': '-', '|': '-',
	}
	runes := []rune(name)
	for i, r := range runes {
		if replacement, ok := replacements[r]; ok {
			runes[i] = replacement
		}
	}
	return string(runes)
}
