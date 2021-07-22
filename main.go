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

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "引数の個数が正しくありません")
		os.Exit(1)
	}
	fmt.Println(".intel_syntax noprefix")
	fmt.Println(".globl main")
	fmt.Println("main:")

	n, input := strtoi(os.Args[1])
	fmt.Printf("  mov rax, %d\n", n)

	for input != "" {
		var c = input[0]
		if c == '+' {
			input = input[1:]
			n, input = strtoi(input)
			fmt.Printf("  add rax, %d\n", n)
			continue
		}
		if c == '-' {
			input = input[1:]
			n, input = strtoi(input)
			fmt.Printf("  sub rax, %d\n", n)
			continue
		}
		fmt.Fprintf(os.Stderr, "予期しない文字です: '%c'\n", c)
		os.Exit(1)
	}

	fmt.Println("  ret")
}
