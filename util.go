package politstonk

import "strings"

func isTicker(s string) bool {
	return strings.HasPrefix(s, "$")
}
