package main

import (
	"fmt"
	"os"
)

var tokenizer *Tokenizer
var userInput string

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

func currentToken() Token {
	return tokenizer.Fetch()
}

// 次のトークンが期待している記号の時には、トークンを1つ読み進めて真を返す。
// それ以外の場合には偽を返す。
func consume(op string) bool {
	if !tokenizer.Fetch().Test(TokenReserved, op) {
		return false
	}
	tokenizer.Succ()
	return true
}

// 文の終端記号であるトークンを1つ読み進めて真を返す。
// それ以外の場合には偽を返す。
func consumeEndLine() bool {
	return consume(";") || consume("\n")
}

func expectEndLine() {
	if !consumeEndLine() {
		madden("文の終端記号ではありません")
	}
}

// 次のトークンの種類が kind だった場合にはトークンを1つ読み進めて真を返す。
// それ以外の場合には偽を返す。
func consumeKind(kind TokenKind) bool {
	if tokenizer.Fetch().kind != kind {
		return false
	}
	tokenizer.Succ()
	return true
}

// 次のトークンが識別子の時には、トークンを1つ読み進めてそのトークンを返す。
// この時、返り値の二番目の値は真になる。
// 逆に識別子でない場合は、偽になる。
func consumeIdentifier() (Token, bool) {
	token := tokenizer.Fetch()
	if token.kind == TokenIdentifier {
		tokenizer.Succ()
		return token, true
	}
	return Token{}, false
}

// 次のトークンが識別子の時には、トークンを1つ読み進めてそのトークンを返す。
// そうでない場合はエラーを報告する。
func expectIdentifier() Token {
	token, ok := consumeIdentifier()
	if !ok {
		errorAt(token.str, "識別子ではありません")
	}
	return token
}

// 次のトークンが期待している記号のときには、トークンを1つ読み進める。
// それ以外の場合にはエラーを報告する。
func expect(op string) {
	var token = tokenizer.Fetch()
	if !token.Test(TokenReserved, op) {
		errorAt(token.str, "'%s'ではありません", op)
	}
	tokenizer.Succ()
}

// 次のトークンが期待している種類の時にはトークンを1つ読み進める。
// それ以外の場合にはエラーを報告する。
func expectKind(kind TokenKind) {
	token := tokenizer.Fetch()
	if token.kind != kind {
		errorAt(token.str, "'%s'ではありません", kind)
	}
	tokenizer.Succ()
}

// 次のトークンが数値の場合、トークンを1つ読み進めてその数値を返す。
// それ以外の場合にはエラーを報告する。
func expectNumber() int {
	token := tokenizer.Fetch()
	if token.kind != TokenNumber {
		errorAt(token.str, "数ではありません")
	}
	var val = token.val
	tokenizer.Succ()
	return val
}

func atEof() bool {
	return tokenizer.Fetch().kind == TokenEof
}

func expectType() Type {
	var varType Type = Type{}
	if consume("*") {
		varType.kind = TypePtr
		ty := expectType()
		varType.ptrTo = &ty
		return varType
	}
	tok := expectIdentifier()
	if tok.str == "int" {
		return Type{kind: TypeInt}
	}
	return varType
}

func consumeType() (Type, bool) {
	var varType Type = Type{}
	if consume("*") {
		varType.kind = TypePtr
		ty := expectType()
		varType.ptrTo = &ty
		return varType, true
	}
	tok, ok := consumeIdentifier()
	if !ok {
		return Type{}, false
	}
	if tok.str == "int" {
		return Type{kind: TypeInt}, true
	}
	return varType, true
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
	NodeAddr         NodeKind = "ADDR"          // &
	NodeDeref        NodeKind = "DEREF"         // *addr
	NodeVarStmt      NodeKind = "VAR STMT"      // var ...
	NodePackageStmt  NodeKind = "PACKAGE STMT"  // package ...
	NodeExprStmt     NodeKind = "EXPR STMT"     // 式文
)

type Node struct {
	kind     NodeKind  // ノードの型
	val      int       // kindがNodeNumの場合にのみ使う
	lvar     *LocalVar // kindがNodeLocalVarの場合にのみ使う
	label    string    // kindがNodeFunctionCallまたはNodePackageの場合にのみ使う
	children []*Node   // 子。lhs, rhsの順でchildrenに格納される
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
var currentFuncLabel = ""
var Env *Environment

func program() {
	Env = NewEnvironment()

	for consumeEndLine() {
	}
	code = []*Node{packageStmt()}
	expectEndLine()

	code = append(code, stmtList().children...)
}

func packageStmt() *Node {
	var n = newLeafNode(NodePackageStmt)

	expectKind(TokenPackage)
	n.label = expectIdentifier().str

	return n
}

func stmtList() *Node {
	var stmts = make([]*Node, 0)
	var endLineRequired = false

	for !atEof() && !(currentToken().kind == TokenReserved && currentToken().str == "}") {
		if endLineRequired {
			errorAt(currentToken().rest, "文の区切り文字が必要です")
		}
		if consumeEndLine() {
			continue
		}
		stmts = append(stmts, stmt())

		endLineRequired = true
		if consumeEndLine() {
			endLineRequired = false
		}
	}
	var node = newNode(NodeStmtList, stmts)
	node.children = stmts
	return node
}

func stmt() *Node {
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
	// var文
	if currentToken().kind == TokenVar {
		return varStmt()
	}

	if consumeKind(TokenReturn) {
		return newNode(NodeReturn, []*Node{expr()})
	} else {
		var n = expr()
		if consume("=") {
			// 代入文
			var e = expr()
			return newBinaryNode(NodeAssign, n, e)
		}
		return newNode(NodeExprStmt, []*Node{n})
	}
}

func varStmt() *Node {
	expectKind(TokenVar)
	var v = variableDeclaration()
	ty, ok := consumeType()

	if !ok {
		// 型が明示されていないときは初期化が必須
		expect("=")
		return newBinaryNode(NodeVarStmt, v, expr())
	} else {
		v.lvar.varType = ty
	}
	if consume("=") {
		return newBinaryNode(NodeVarStmt, v, expr())
	}
	return newNode(NodeVarStmt, []*Node{v})
}

func funcDefinition() *Node {
	expectKind(TokenFunc)
	identifier := expectIdentifier()

	var prevFuncLabel = currentFuncLabel
	currentFuncLabel = identifier.str
	var fn = Env.RegisterFunc(currentFuncLabel)

	var parameters = make([]*Node, 0)

	expect("(")
	for !consume(")") {
		if len(parameters) > 0 {
			expect(",")
		}
		lvarNode := variableDeclaration()
		parameters = append(parameters, lvarNode)
		lvarNode.lvar.varType = expectType()
		fn.ParameterTypes = append(fn.ParameterTypes, lvarNode.lvar.varType)
	}

	// 本当はvoid型が正しいけれど、テストを簡単にするためしばらくはint型で定義
	fn.ReturnValueType = NewType(TypeInt)
	var ty, ok = consumeType()
	if ok {
		fn.ReturnValueType = ty
	}

	expect("{")

	var node = newNode(NodeFunctionDef, make([]*Node, 0))
	node.label = identifier.str
	node.children = append(node.children, stmtList())
	node.children = append(node.children, parameters...)

	expect("}")

	currentFuncLabel = prevFuncLabel

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
		return node
	}

	var s = stmt()
	if consume("{") {
		// while文
		if s.kind != NodeExprStmt {
			madden("for文の条件に式以外が書かれています")
		}
		node.children[1] = s.children[0] // expr
		node.children[3] = stmtList()
		expect("}")
		return node
	}

	// 通常のfor文
	node.children[0] = s
	expect(";")
	node.children[1] = stmt().children[0] // expr
	expect(";")
	node.children[2] = stmt()

	expect("{")
	node.children[3] = stmtList()
	expect("}")
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
	return newBinaryNode(NodeIf, lhs, rhs)
}

func elseStmt() *Node {
	expectKind(TokenElse)
	expect("{")
	var stmts = stmtList()
	expect("}")
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
	if consume("*") {
		return newNode(NodeDeref, []*Node{unary()})
	}
	if consume("&") {
		return newNode(NodeAddr, []*Node{unary()})
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

	if tokenizer.Prefetch(1).Test(TokenReserved, "(") {
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
	return variableRef()
}

func variableRef() *Node {
	var tok = expectIdentifier()
	var node = newLeafNode(NodeLocalVar)
	node.lvar = Env.FindLocalVar(currentFuncLabel, tok)
	if node.lvar == nil {
		errorAt(tok.rest, "未定義の変数です %s", tok.str)
	}
	return node
}

func variableDeclaration() *Node {
	var tok = expectIdentifier()
	var node = newLeafNode(NodeLocalVar)
	lvar := Env.FindLocalVar(currentFuncLabel, tok)
	if lvar != nil {
		errorAt(tok.rest, "すでに定義済みの変数です %s", tok.str)
	}
	node.lvar = Env.AddLocalVar(currentFuncLabel, tok)
	return node
}
