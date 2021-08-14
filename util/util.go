package util

import (
	"fmt"
	"os"
	"unicode"
)

func Strtoi(s string) (int, string) {
	var res = 0
	for i, c := range s {
		if !unicode.IsDigit(c) {
			return res, s[i:]
		}
		res = res*10 + int(c) - int('0')
	}
	return res, ""
}

func IsAlnum(c rune) bool {
	return IsAlpha(c) || unicode.IsDigit(c)
}

func IsAlpha(c rune) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

func RuneAt(s string, i int) rune {
	return []rune(s)[i]
}

func Alarm(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintln(os.Stderr, "")
	os.Exit(1)
}
