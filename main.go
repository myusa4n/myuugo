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

type NodeKind string

const (
	NodeAdd NodeKind = "ADD" // +
	NodeSub NodeKind = "SUB" // -
	NodeMul NodeKind = "MUL" // *
	NodeDiv NodeKind = "DIV" // /
	NodeNum NodeKind = "NUM" // 整数
)

type Node struct {
	kind NodeKind // ノードの型
	lhs  *Node    // 左辺
	rhs  *Node    // 右辺
	val  int      // kindがNodeNumの場合にのみ使う
}

func newNode(kind NodeKind, lhs *Node, rhs *Node) *Node {
	return &Node{kind: kind, lhs: lhs, rhs: rhs}
}

func newNodeNum(val int) *Node {
	return &Node{kind: NodeNum, val: val}
}

func expr() *Node {
	var n = mul()
	for {
		if consume('+') {
			n = newNode(NodeAdd, n, mul())
		} else if consume('-') {
			n = newNode(NodeSub, n, mul())
		} else {
			return n
		}
	}
}

func mul() *Node {
	var n = unary()
	for {
		if consume('*') {
			n = newNode(NodeMul, n, unary())
		} else if consume('/') {
			n = newNode(NodeDiv, n, unary())
		} else {
			return n
		}
	}
}

func unary() *Node {
	if consume('+') {
		return primary()
	}
	if consume('-') {
		return newNode(NodeSub, newNodeNum(0), primary())
	}
	return primary()
}

func primary() *Node {
	// 次のトークンが "(" なら、"(" expr ")" のはず
	if consume('(') {
		var n = expr()
		expect(')')
		return n
	}
	return newNodeNum(expectNumber())
}

// ユーザーからの入力プログラム
var userInput string

// 現在着目しているトークン以降のトークン列
var tokens []Token

func madden(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintln(os.Stderr, "")
	os.Exit(1)
}

func errorAt(str string, format string, args ...interface{}) {
	fmt.Fprintln(os.Stderr, userInput)
	pos := len(userInput) - len(str)
	if pos > 0 {
		fmt.Fprintf(os.Stderr, "%*s", pos, " ")
	}
	fmt.Fprintf(os.Stderr, "^ ")
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
		errorAt(token.str, "'%c'ではありません", op)
	}
	tokens = tokens[1:]
}

// 次のトークンが数値の場合、トークンを1つ読み進めてその数値を返す。
// それ以外の場合にはエラーを報告する。
func expectNumber() int {
	token := tokens[0]
	if token.kind != Number {
		errorAt(token.str, "数ではありません")
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
		if c == '+' || c == '-' || c == '*' || c == '/' || c == '(' || c == ')' {
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
		errorAt(input, "トークナイズできません")
	}
	tokens = append(tokens, newToken(Eof, ""))
	return tokens
}

func gen(node *Node) {
	if node.kind == NodeNum {
		fmt.Printf("  push %d\n", node.val)
		return
	}
	gen(node.lhs)
	gen(node.rhs)

	fmt.Println("  pop rdi")
	fmt.Println("  pop rax")

	switch node.kind {
	case NodeAdd:
		fmt.Println("  add rax, rdi")
	case NodeSub:
		fmt.Println("  sub rax, rdi")
	case NodeMul:
		fmt.Println("  imul rax, rdi")
	case NodeDiv:
		fmt.Println("  cqo")
		fmt.Println("  idiv rdi")
	}
	fmt.Println("  push rax")
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "引数の個数が正しくありません")
		os.Exit(1)
	}

	userInput = os.Args[1]
	tokens = tokenize(userInput)
	var node = expr()

	// アセンブリの前半部分
	fmt.Println(".intel_syntax noprefix")
	fmt.Println(".globl main")
	fmt.Println("main:")

	// 抽象構文木を下りながらコード生成
	gen(node)

	// スタックトップに式全体の値が残っているはずなので
	// それをRAXにロードして関数からの返り値とする
	fmt.Println("  pop rax")
	fmt.Println("  ret")
}
