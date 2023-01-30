package main

import (
	"regexp"
	"strings"
)

func avg(in []int) (sum int) {
	if len(in) == 0 {
		return 0
	}

	for _, n := range in {
		sum += n
	}
	if sum == 0 {
		return 0
	}
	sum /= len(in)
	return
}

func parameterize(x string) string {
	x = strings.ToLower(x)
	re := regexp.MustCompile(`[^\w]`)
	re2 := regexp.MustCompile(`_+`)
	x = re.ReplaceAllString(x, "_")
	x = re2.ReplaceAllString(x, "_")
	return x
}

func unique(strings []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range strings {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
