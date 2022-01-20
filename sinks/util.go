package sinks

import (
	"encoding/base64"
	"strconv"
)

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
