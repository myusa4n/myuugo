package main

import (
	"fmt"
	"os"
	"strconv"
	"unicode"
)

// (先頭の識別子, 識別子を切り出して得られた残りの文字列)  を返す
func getIdentifier(s string) (string, string) {
	var res = ""
	for i, c := range s {
		if (i == 0 && unicode.IsDigit(c)) || !(isAlpha(c) || (c == '_')) {
			return res, s[i:]
		}
		res += string(c)
	}
	return res, ""
}

type TokenKind string

const (
	TokenReserved   TokenKind = "RESERVED"
	TokenNumber     TokenKind = "NUMBER"
	TokenIdentifier TokenKind = "IDENTIFIER"
	TokenEof        TokenKind = "EOF"
	TokenReturn     TokenKind = "return"
	TokenIf         TokenKind = "if"
	TokenElse       TokenKind = "else"
)

type Token struct {
	kind TokenKind // トークンの型
	val  int       // kindがNumberの場合、その数値
	str  string    // トークン文字列
	rest string    // 自信を含めた残りすべてのトークン文字列
}

// ユーザーからの入力プログラム
var userInput string

// 現在着目しているトークン以降のトークン列
var tokens []Token

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

// 現在のトークンを返す
func currentToken() Token {
	return tokens[0]
}

// 次のトークンを先読みする
func prefetch() Token {
	return tokens[1]
}

// 次のトークンが期待している記号の時には、トークンを1つ読み進めて真を返す。
// それ以外の場合には偽を返す。
func consume(op string) bool {
	token := tokens[0]
	if token.kind != TokenReserved || token.str != op {
		return false
	}
	tokens = tokens[1:]
	return true
}

// 文の終端記号であるトークンを1つ読み進めて真を返す。
// それ以外の場合には偽を返す。
func consumeEndLine() bool {
	return consume(";") || consume("\n")
}

// 次のトークンの種類が kind だった場合にはトークンを1つ読み進めて真を返す。
// それ以外の場合には偽を返す。
func consumeKind(kind TokenKind) bool {
	token := tokens[0]
	if token.kind != kind {
		return false
	}
	tokens = tokens[1:]
	return true
}

// 次のトークンが識別子の時には、トークンを1つ読み進めてそのトークンを返す。
// この時、返り値の二番目の値は真になる。
// 逆に識別子でない場合は、偽になる。
func consumeIdentifier() (Token, bool) {
	token := tokens[0]
	if token.kind == TokenIdentifier {
		tokens = tokens[1:]
		return token, true
	}
	return Token{}, false
}

// 次のトークンが期待している記号のときには、トークンを1つ読み進める。
// それ以外の場合にはエラーを報告する。
func expect(op string) {
	token := tokens[0]
	if token.kind != TokenReserved || token.str != op {
		errorAt(token.str, "'%s'ではありません", op)
	}
	tokens = tokens[1:]
}

// 次のトークンが期待している種類の時にはトークンを1つ読み進める。
// それ以外の場合にはエラーを報告する。
func expectKind(kind TokenKind) {
	token := tokens[0]
	if token.kind != kind {
		errorAt(token.str, "'%s'ではありません", kind)
	}
	tokens = tokens[1:]
}

// 次のトークンが数値の場合、トークンを1つ読み進めてその数値を返す。
// それ以外の場合にはエラーを報告する。
func expectNumber() int {
	token := tokens[0]
	if token.kind != TokenNumber {
		errorAt(token.str, "数ではありません")
	}
	var val = token.val
	tokens = tokens[1:]
	return val
}

func atEof() bool {
	return tokens[0].kind == TokenEof
}

func newToken(kind TokenKind, str string, rest string) Token {
	return Token{kind: kind, str: str, rest: rest}
}

func tokenize(input string) []Token {
	var tokens []Token = make([]Token, 0)

	for input != "" {
		if len(input) >= 2 {
			var head2 = input[:2]
			if head2 == "==" || head2 == "!=" || head2 == "<=" || head2 == ">=" {
				tokens = append(tokens, newToken(TokenReserved, head2, input))
				input = input[2:]
				continue
			}
		}

		var c = runeAt(input, 0)
		if isAlpha(c) || (c == '_') {
			// input から 識別子を取り出す
			var token = newToken(TokenIdentifier, "", input)
			token.str, input = getIdentifier(input)
			if token.str == string(TokenReturn) {
				token.kind = TokenReturn
			} else if token.str == string(TokenIf) {
				token.kind = TokenIf
			} else if token.str == string(TokenElse) {
				token.kind = TokenElse
			}

			tokens = append(tokens, token)
			continue
		}
		if c == '+' || c == '-' || c == '*' || c == '/' || c == '(' || c == ')' || c == '<' ||
			c == '>' || c == ';' || c == '\n' || c == '=' || c == '{' || c == '}' {
			tokens = append(tokens, newToken(TokenReserved, string(c), input))
			input = input[1:]
			continue
		}
		if unicode.IsSpace(c) {
			input = input[1:]
			continue
		}
		if unicode.IsDigit(c) {
			var token = newToken(TokenNumber, "", input)
			token.val, input = strtoi(input)
			token.str = strconv.Itoa(token.val)
			tokens = append(tokens, token)
			continue
		}
		errorAt(input, "トークナイズできません")
	}
	tokens = append(tokens, newToken(TokenEof, "", ""))
	return tokens
}

type NodeKind string

const (
	NodeAdd        NodeKind = "ADD"         // +
	NodeSub        NodeKind = "SUB"         // -
	NodeMul        NodeKind = "MUL"         // *
	NodeDiv        NodeKind = "DIV"         // /
	NodeEql        NodeKind = "EQL"         // ==
	NodeNotEql     NodeKind = "NOT EQL"     // !=
	NodeLess       NodeKind = "LESS"        // <
	NodeLessEql    NodeKind = "LESS EQL"    // <=
	NodeGreater    NodeKind = "GREATER"     // >
	NodeGreaterEql NodeKind = "GREATER EQL" // >=
	NodeAssign     NodeKind = "ASSIGN"      // =
	NodeReturn     NodeKind = "RETURN"      // return
	NodeLocalVar   NodeKind = "Local Var"   // ローカル変数
	NodeNum        NodeKind = "NUM"         // 整数
	NodeMetaIf     NodeKind = "META IF"     // if ... else ...
	NodeIf         NodeKind = "IF"          // if
	NodeElse       NodeKind = "ELSE"        // else
)

type Node struct {
	kind   NodeKind // ノードの型
	lhs    *Node    // 左辺
	rhs    *Node    // 右辺
	val    int      // kindがNodeNumの場合にのみ使う
	offset int      // kindがNodeLocalVarの場合にのみ使う
}

func newNode(kind NodeKind, lhs *Node, rhs *Node) *Node {
	return &Node{kind: kind, lhs: lhs, rhs: rhs}
}

func newNodeNum(val int) *Node {
	return &Node{kind: NodeNum, val: val}
}

var code [100]*Node

func program() {
	var i = 0
	for !atEof() {
		var s = stmt()
		if s != nil {
			code[i] = s
			i += 1
		}
	}
	code[i] = nil
}

func stmt() *Node {
	// 空文
	if consumeEndLine() {
		return nil
	}
	// if文
	if currentToken().kind == TokenIf {
		return metaIfStmt()
	}

	var n *Node
	if consumeKind(TokenReturn) {
		n = newNode(NodeReturn, expr(), nil)
	} else {
		n = expr()
		if consume("=") {
			// 代入文
			var e = expr()
			n = newNode(NodeAssign, n, e)
		}
	}
	consumeEndLine()
	return n
}

func metaIfStmt() *Node {
	token := currentToken()
	if token.kind != TokenIf {
		errorAt(token.str, "'%s'ではありません", TokenIf)
	}

	var ifNode = ifStmt()
	if currentToken().kind == TokenElse {
		var elseNode = elseStmt()
		return newNode(NodeMetaIf, ifNode, elseNode)
	}
	return newNode(NodeMetaIf, ifNode, nil)
}

func ifStmt() *Node {
	var node = newNode(NodeIf, nil, nil)
	expectKind(TokenIf)
	node.lhs = expr()
	expect("{")
	if !consume("}") {
		consumeEndLine()
		node.rhs = stmt()
		expect("}")
	}
	consumeEndLine()
	return node
}

func elseStmt() *Node {
	var node = newNode(NodeElse, nil, nil)
	expectKind(TokenElse)
	expect("{")
	if !consume("}") {
		consumeEndLine()
		node.lhs = stmt()
		expect("}")
	}
	consumeEndLine()
	return node
}

func expr() *Node {
	return equality()
}

func equality() *Node {
	var n = relational()
	for {
		if consume("==") {
			n = newNode(NodeEql, n, relational())
		} else if consume("!=") {
			n = newNode(NodeNotEql, n, relational())
		} else {
			return n
		}
	}
}

func relational() *Node {
	var n = add()
	for {
		if consume("<") {
			n = newNode(NodeLess, n, add())
		} else if consume("<=") {
			n = newNode(NodeLessEql, n, add())
		} else if consume(">") {
			n = newNode(NodeGreater, n, add())
		} else if consume(">=") {
			n = newNode(NodeGreaterEql, n, add())
		} else {
			return n
		}
	}
}

func add() *Node {
	var n = mul()
	for {
		if consume("+") {
			n = newNode(NodeAdd, n, mul())
		} else if consume("-") {
			n = newNode(NodeSub, n, mul())
		} else {
			return n
		}
	}
}

func mul() *Node {
	var n = unary()
	for {
		if consume("*") {
			n = newNode(NodeMul, n, unary())
		} else if consume("/") {
			n = newNode(NodeDiv, n, unary())
		} else {
			return n
		}
	}
}

func unary() *Node {
	if consume("+") {
		return primary()
	}
	if consume("-") {
		return newNode(NodeSub, newNodeNum(0), primary())
	}
	return primary()
}

func primary() *Node {
	// 次のトークンが "(" なら、"(" expr ")" のはず
	if consume("(") {
		var n = expr()
		expect(")")
		return n
	}
	var tok, ok = consumeIdentifier()
	if ok {
		var node = newNode(NodeLocalVar, nil, nil)
		lvar, ok := findLocalVar(tok)

		if ok {
			node.offset = lvar.offset
			return node
		}

		lvar = LocalVar{name: tok.str}
		if len(locals) == 0 {
			lvar.offset = 0 + 8
		} else {
			lvar.offset = locals[len(locals)-1].offset + 8
		}
		node.offset = lvar.offset
		locals = append(locals, lvar)
		return node
	}
	return newNodeNum(expectNumber())
}
