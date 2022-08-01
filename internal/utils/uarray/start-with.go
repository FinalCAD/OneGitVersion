package uarray

import "strings"

func StartWith(a []string, b string) bool {
	for _, s := range a {
		if strings.HasPrefix(s, b) {
			return true
		}
	}
	return false
}
