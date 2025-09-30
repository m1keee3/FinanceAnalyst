package utils

import "unicode"

func IsLetterOnly(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func IsAllUpper(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) && !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}
