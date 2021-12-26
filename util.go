package stonkcritter

import (
	"encoding/base64"
	"strconv"
	"strings"
	"time"
)

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

func normalizeDate(ts string) time.Time {
	t, _ := time.Parse("2006-01-02", ts)

	if strings.Contains(ts, "/") {
		t, _ = time.Parse("01/02/2006", ts)
	}

	return t
}

// deduper is a quick and dirty deduplicator, returns two closures, one to test
// if the given pair is locked, and one to mark the given pair as locked
func deduper() (func(int64, string) bool, func(int64, string)) {
	locks := []string{}

	shouldsend := func(chatID int64, msg string) bool {
		msghash := base64.RawStdEncoding.EncodeToString([]byte(msg))
		lock := strconv.Itoa(int(chatID)) + "_" + msghash
		for _, l := range locks {
			if l == lock {
				return false
			}
		}
		return true
	}

	marksent := func(chatID int64, msg string) {
		msghash := base64.RawStdEncoding.EncodeToString([]byte(msg))
		locks = append(locks, strconv.Itoa(int(chatID))+"_"+msghash)
	}

	return shouldsend, marksent
}
