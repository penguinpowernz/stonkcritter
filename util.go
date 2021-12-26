package politstonk

import (
	"strings"
	"time"
)

func isTicker(s string) bool {
	return strings.HasPrefix(s, "$")
}
func normalizeDate(ts string) time.Time {
	t, _ := time.Parse("2006-01-02", ts)

	if strings.Contains(ts, "/") {
		t, _ = time.Parse("01/02/2006", ts)
	}

	return t
}
