package bot

import "strings"

func isTicker(s string) bool {
	return strings.HasPrefix(s, "$")
}

func tgEscape(s string) string {
	s = strings.ReplaceAll(s, ".", `\.`)
	s = strings.ReplaceAll(s, "(", `\(`)
	s = strings.ReplaceAll(s, ")", `\)`)
	s = strings.ReplaceAll(s, "-", `\-`)
	s = strings.ReplaceAll(s, "$", `\$`)
	s = strings.ReplaceAll(s, ":", `\:`)
	return s
}
