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
	TokenFor        TokenKind = "for"
	TokenFunc       TokenKind = "func"
)

type Token struct {
	kind TokenKind // トークンの型
	val  int       // kindがNumberの場合、その数値
	str  string    // トークン文字列
	rest string    // 自信を含めた残りすべてのトークン文字列
}

func (t Token) test(kind TokenKind, str string) bool {
	return t.kind == kind && t.str == str
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

// n個先のトークンを先読みする
func prefetch(n int) Token {
	return tokens[n]
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

// 次のトークンが識別子の時には、トークンを1つ読み進めてそのトークンを返す。
// そうでない場合はエラーを報告する。
func expectIdentifier() Token {
	token := tokens[0]
	if token.kind == TokenIdentifier {
		tokens = tokens[1:]
		return token
	}
	errorAt(token.str, "識別子ではありません")
	return token // 到達しない
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
			} else if token.str == string(TokenFor) {
				token.kind = TokenFor
			} else if token.str == string(TokenFunc) {
				token.kind = TokenFunc
			}
			tokens = append(tokens, token)
			continue
		}
		if c == '+' || c == '-' || c == '*' || c == '/' || c == '(' || c == ')' || c == '<' ||
			c == '>' || c == ';' || c == '\n' || c == '=' || c == '{' || c == '}' || c == ',' {
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
	NodeAdd          NodeKind = "ADD"           // +
	NodeSub          NodeKind = "SUB"           // -
	NodeMul          NodeKind = "MUL"           // *
	NodeDiv          NodeKind = "DIV"           // /
	NodeEql          NodeKind = "EQL"           // ==
	NodeNotEql       NodeKind = "NOT EQL"       // !=
	NodeLess         NodeKind = "LESS"          // <
	NodeLessEql      NodeKind = "LESS EQL"      // <=
	NodeGreater      NodeKind = "GREATER"       // >
	NodeGreaterEql   NodeKind = "GREATER EQL"   // >=
	NodeAssign       NodeKind = "ASSIGN"        // =
	NodeReturn       NodeKind = "RETURN"        // return
	NodeLocalVar     NodeKind = "Local Var"     // ローカル変数
	NodeNum          NodeKind = "NUM"           // 整数
	NodeMetaIf       NodeKind = "META IF"       // if ... else ...
	NodeIf           NodeKind = "IF"            // if
	NodeElse         NodeKind = "ELSE"          // else
	NodeStmtList     NodeKind = "STMT LIST"     // stmt*
	NodeFor          NodeKind = "FOR"           // for
	NodeFunctionCall NodeKind = "FUNCTION CALL" // fn()
	NodeFunctionDef  NodeKind = "FUNCTION DEF"  // func fn() { ... }
)

type Node struct {
	kind     NodeKind // ノードの型
	val      int      // kindがNodeNumの場合にのみ使う
	offset   int      // kindがNodeLocalVarの場合にのみ使う
	label    string   // kindがNodeFunctionCallの場合にのみ使う
	children []*Node  // 子。lhs, rhsの順でchildrenに格納される
}

func newNode(kind NodeKind, children []*Node) *Node {
	return &Node{kind: kind, children: children}
}

func newBinaryNode(kind NodeKind, lhs *Node, rhs *Node) *Node {
	return &Node{kind: kind, children: []*Node{lhs, rhs}}
}

func newLeafNode(kind NodeKind) *Node {
	return &Node{kind: kind}
}

func newNodeNum(val int) *Node {
	return &Node{kind: NodeNum, val: val}
}

var code []*Node

func program() {
	code = stmtList().children
}

func stmtList() *Node {
	var stmts = make([]*Node, 0)

	for !atEof() && !(currentToken().kind == TokenReserved && currentToken().str == "}") {
		var s = stmt()
		if s != nil {
			stmts = append(stmts, s)
		}
	}
	var node = newNode(NodeStmtList, stmts)
	node.children = stmts
	return node
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
	// for文
	if currentToken().kind == TokenFor {
		return forStmt()
	}
	// 関数定義
	if currentToken().kind == TokenFunc {
		return funcDefinition()
	}

	var n *Node
	if consumeKind(TokenReturn) {
		n = newNode(NodeReturn, []*Node{expr()})
	} else {
		n = expr()
		if consume("=") {
			// 代入文
			var e = expr()
			n = newBinaryNode(NodeAssign, n, e)
		}
	}
	consumeEndLine()
	return n
}

func funcDefinition() *Node {
	expectKind(TokenFunc)
	identifier := expectIdentifier()

	var parameters = make([]*Node, 0)

	expect("(")
	for !consume(")") {
		if len(parameters) > 0 {
			expect(",")
		}
		parameters = append(parameters, variable())
	}
	expect("{")

	var node = newNode(NodeFunctionDef, make([]*Node, 0))
	node.label = identifier.str
	node.children = append(node.children, stmtList())
	node.children = append(node.children, parameters...)

	expect("}")
	consumeEndLine()
	return node
}

// range は未対応
func forStmt() *Node {
	expectKind(TokenFor)
	// 初期化, ループ条件, 更新式, 繰り返す文
	var node = newNode(NodeFor, []*Node{nil, nil, nil, nil})

	if consume("{") {
		// 無限ループ
		node.children[3] = stmtList()
		expect("}")
		consumeEndLine()
		return node
	}

	var s = stmt()
	if consume("{") {
		// while文
		node.children[1] = s
		node.children[3] = stmtList()
		expect("}")
		consumeEndLine()
		return node
	}

	// 通常のfor文
	node.children[0] = s
	node.children[1] = stmt()
	node.children[2] = stmt()

	expect("{")
	node.children[3] = stmtList()
	expect("}")
	consumeEndLine()
	return node
}

func metaIfStmt() *Node {
	token := currentToken()
	if token.kind != TokenIf {
		errorAt(token.str, "'%s'ではありません", TokenIf)
	}

	var ifNode = ifStmt()
	if currentToken().kind == TokenElse {
		var elseNode = elseStmt()
		return newBinaryNode(NodeMetaIf, ifNode, elseNode)
	}
	return newBinaryNode(NodeMetaIf, ifNode, nil)
}

func ifStmt() *Node {
	expectKind(TokenIf)
	var lhs = expr()
	expect("{")
	var rhs = stmtList()
	expect("}")
	consumeEndLine()
	return newBinaryNode(NodeIf, lhs, rhs)
}

func elseStmt() *Node {
	expectKind(TokenElse)
	expect("{")
	var stmts = stmtList()
	expect("}")
	consumeEndLine()
	return newNode(NodeElse, []*Node{stmts})
}

func expr() *Node {
	return equality()
}

func equality() *Node {
	var n = relational()
	for {
		if consume("==") {
			n = newBinaryNode(NodeEql, n, relational())
		} else if consume("!=") {
			n = newBinaryNode(NodeNotEql, n, relational())
		} else {
			return n
		}
	}
}

func relational() *Node {
	var n = add()
	for {
		if consume("<") {
			n = newBinaryNode(NodeLess, n, add())
		} else if consume("<=") {
			n = newBinaryNode(NodeLessEql, n, add())
		} else if consume(">") {
			n = newBinaryNode(NodeGreater, n, add())
		} else if consume(">=") {
			n = newBinaryNode(NodeGreaterEql, n, add())
		} else {
			return n
		}
	}
}

func add() *Node {
	var n = mul()
	for {
		if consume("+") {
			n = newBinaryNode(NodeAdd, n, mul())
		} else if consume("-") {
			n = newBinaryNode(NodeSub, n, mul())
		} else {
			return n
		}
	}
}

func mul() *Node {
	var n = unary()
	for {
		if consume("*") {
			n = newBinaryNode(NodeMul, n, unary())
		} else if consume("/") {
			n = newBinaryNode(NodeDiv, n, unary())
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
		return newBinaryNode(NodeSub, newNodeNum(0), primary())
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

	if currentToken().kind != TokenIdentifier {
		return newNodeNum(expectNumber())
	}

	if prefetch(1).test(TokenReserved, "(") {
		// 関数呼び出し
		var tok = expectIdentifier()
		expect("(")
		var node = newNode(NodeFunctionCall, make([]*Node, 0))
		node.label = tok.str
		for !consume(")") {
			if len(node.children) > 0 {
				expect(",")
			}
			node.children = append(node.children, expr())
		}
		return node
	}
	return variable()
}

func variable() *Node {
	var tok = expectIdentifier()
	var node = newLeafNode(NodeLocalVar)
	lvar := addLocalVar("main", tok)
	node.offset = lvar.offset
	return node
}
