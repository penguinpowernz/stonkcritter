package models

import (
	"regexp"
	"strings"
	"time"
)

func normalizeDate(ts string) time.Time {
	t, _ := time.Parse("2006-01-02", ts)

	if strings.Contains(ts, "/") {
		t, _ = time.Parse("01/02/2006", ts)
	}

	return t
}

func parameterize(x string) string {
	x = strings.ToLower(x)
	re := regexp.MustCompile(`[^\w]`)
	re2 := regexp.MustCompile(`_+`)
	x = re.ReplaceAllString(x, "_")
	x = re2.ReplaceAllString(x, "_")
	return x
}
