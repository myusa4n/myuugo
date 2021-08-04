package main

import (
	"fmt"
	"os"
	"unicode"
)

func strtoi(s string) (int, string) {
	var res = 0
	for i, c := range s {
		if !unicode.IsDigit(c) {
			return res, s[i:]
		}
		res = res*10 + int(c) - int('0')
	}
	return res, ""
}

func isAlnum(c rune) bool {
	return isAlpha(c) || unicode.IsDigit(c)
}

func isAlpha(c rune) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

func runeAt(s string, i int) rune {
	return []rune(s)[i]
}

func madden(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintln(os.Stderr, "")
	os.Exit(1)
}
