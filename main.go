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

func runeAt(s string, i int) rune {
	return []rune(s)[i]
}

type TokenKind string

const (
	Reserved TokenKind = "RESERVED"
	Number   TokenKind = "NUMBER"
	Eof      TokenKind = "EOF"
)

type Token struct {
	kind TokenKind // トークンの型
	val  int       // kindがNumberの場合、その数値
	str  string    // トークン文字列
}

// 現在着目しているトークン以降のトークン列
var tokens []Token

func madden(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintln(os.Stderr, "")
	os.Exit(1)
}

// 次のトークンが期待している記号の時には、トークンを1つ読み進めて真を返す。
// それ以外の場合には偽を返す。
func consume(op rune) bool {
	token := tokens[0]
	if token.kind != Reserved || []rune(token.str)[0] != op {
		return false
	}
	tokens = tokens[1:]
	return true
}

// 次のトークンが期待している記号のときには、トークンを1つ読み進める。
// それ以外の場合にはエラーを報告する。
func expect(op rune) {
	token := tokens[0]
	if token.kind != Reserved || runeAt(token.str, 0) != op {
		madden("'%c'ではありません", op)
	}
	tokens = tokens[1:]
}

// 次のトークンが数値の場合、トークンを1つ読み進めてその数値を返す。
// それ以外の場合にはエラーを報告する。
func expectNumber() int {
	token := tokens[0]
	if token.kind != Number {
		madden("数ではありません")
	}
	var val = token.val
	tokens = tokens[1:]
	return val
}

func atEof() bool {
	return tokens[0].kind == Eof
}

func newToken(kind TokenKind, str string) Token {
	return Token{kind: kind, str: str}
}

func tokenize(input string) []Token {
	var tokens []Token = make([]Token, 0)

	for input != "" {
		var c = runeAt(input, 0)
		if unicode.IsSpace(c) {
			input = input[1:]
			continue
		}
		if c == '+' || c == '-' {
			tokens = append(tokens, newToken(Reserved, input))
			input = input[1:]
			continue
		}
		if unicode.IsDigit(c) {
			var token = newToken(Number, input)
			token.val, input = strtoi(input)
			tokens = append(tokens, token)
			continue
		}
		madden("トークナイズできません")
	}
	tokens = append(tokens, newToken(Eof, ""))
	return tokens
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "引数の個数が正しくありません")
		os.Exit(1)
	}

	tokens = tokenize(os.Args[1])

	// アセンブリの前半部分
	fmt.Println(".intel_syntax noprefix")
	fmt.Println(".globl main")
	fmt.Println("main:")

	fmt.Printf("  mov rax, %d\n", expectNumber())

	for !atEof() {
		if consume('+') {
			fmt.Printf("  add rax, %d\n", expectNumber())
		}
		expect('-')
		fmt.Printf("  sub rax, %d\n", expectNumber())
	}

	fmt.Println("  ret")
}
