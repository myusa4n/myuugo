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

// 次のトークンの種類が kind だった場合にはトークンを1つ読み進めて真を返す。
// それ以外の場合には偽を返す。
func consume(kind TokenKind) bool {
	if tokenizer.Fetch().kind != kind {
		return false
	}
	tokenizer.Succ()
	return true
}

// 文の終端記号であるトークンを1つ読み進めて真を返す。
// それ以外の場合には偽を返す。
func consumeEndLine() bool {
	return consume(TokenSemicolon) || consume(TokenNewLine)
}

func expectEndLine() {
	if !consumeEndLine() {
		madden("文の終端記号ではありません")
	}
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

// 次のトークンが期待しているkindのときには、トークンを1つ読み進める。
// それ以外の場合にはエラーを報告する。
func expect(kind TokenKind) {
	var token = tokenizer.Fetch()
	if !consume(kind) {
		errorAt(token.str, "'%s'ではありません", kind)
	}
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
	if consume(TokenStar) {
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
	if consume(TokenStar) {
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
	var n = NewLeafNode(NodePackageStmt)

	expect(TokenPackage)
	n.label = expectIdentifier().str

	return n
}

func stmtList() *Node {
	var stmts = make([]*Node, 0)
	var endLineRequired = false

	for !atEof() && !(currentToken().Test(TokenRbrace)) {
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
	var node = NewNode(NodeStmtList, stmts)
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

	if consume(TokenReturn) {
		return NewNode(NodeReturn, []*Node{expr()})
	} else {
		var n = expr()
		if consume(TokenEqual) {
			// 代入文
			var e = expr()
			return NewBinaryNode(NodeAssign, n, e)
		}
		return NewNode(NodeExprStmt, []*Node{n})
	}
}

func varStmt() *Node {
	expect(TokenVar)
	var v = variableDeclaration()
	ty, ok := consumeType()

	if !ok {
		// 型が明示されていないときは初期化が必須
		expect(TokenEqual)
		return NewBinaryNode(NodeVarStmt, v, expr())
	} else {
		v.lvar.varType = ty
	}
	if consume(TokenEqual) {
		return NewBinaryNode(NodeVarStmt, v, expr())
	}
	return NewNode(NodeVarStmt, []*Node{v})
}

func funcDefinition() *Node {
	expect(TokenFunc)
	identifier := expectIdentifier()

	var prevFuncLabel = currentFuncLabel
	currentFuncLabel = identifier.str
	var fn = Env.RegisterFunc(currentFuncLabel)

	var parameters = make([]*Node, 0)

	expect(TokenLparen)
	for !consume(TokenRparen) {
		if len(parameters) > 0 {
			expect(TokenComma)
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

	expect(TokenLbrace)

	var node = NewNode(NodeFunctionDef, make([]*Node, 0))
	node.label = identifier.str
	node.children = append(node.children, stmtList())
	node.children = append(node.children, parameters...)

	expect(TokenRbrace)

	currentFuncLabel = prevFuncLabel

	return node
}

// range は未対応
func forStmt() *Node {
	expect(TokenFor)
	// 初期化, ループ条件, 更新式, 繰り返す文
	var node = NewNode(NodeFor, []*Node{nil, nil, nil, nil})

	if consume(TokenLbrace) {
		// 無限ループ
		node.children[3] = stmtList()
		expect(TokenRbrace)
		return node
	}

	var s = stmt()
	if consume(TokenLbrace) {
		// while文
		if s.kind != NodeExprStmt {
			madden("for文の条件に式以外が書かれています")
		}
		node.children[1] = s.children[0] // expr
		node.children[3] = stmtList()
		expect(TokenRbrace)
		return node
	}

	// 通常のfor文
	node.children[0] = s
	expect(TokenSemicolon)
	node.children[1] = stmt().children[0] // expr
	expect(TokenSemicolon)
	node.children[2] = stmt()

	expect(TokenLbrace)
	node.children[3] = stmtList()
	expect(TokenRbrace)
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
		return NewBinaryNode(NodeMetaIf, ifNode, elseNode)
	}
	return NewBinaryNode(NodeMetaIf, ifNode, nil)
}

func ifStmt() *Node {
	expect(TokenIf)
	var lhs = expr()
	expect(TokenLbrace)
	var rhs = stmtList()
	expect(TokenRbrace)
	return NewBinaryNode(NodeIf, lhs, rhs)
}

func elseStmt() *Node {
	expect(TokenElse)
	expect(TokenLbrace)
	var stmts = stmtList()
	expect(TokenRbrace)
	return NewNode(NodeElse, []*Node{stmts})
}

func expr() *Node {
	return equality()
}

func equality() *Node {
	var n = relational()
	for {
		if consume(TokenDoubleEqual) {
			n = NewBinaryNode(NodeEql, n, relational())
		} else if consume(TokenNotEqual) {
			n = NewBinaryNode(NodeNotEql, n, relational())
		} else {
			return n
		}
	}
}

func relational() *Node {
	var n = add()
	for {
		if consume(TokenLess) {
			n = NewBinaryNode(NodeLess, n, add())
		} else if consume(TokenLessEqual) {
			n = NewBinaryNode(NodeLessEql, n, add())
		} else if consume(TokenGreater) {
			n = NewBinaryNode(NodeGreater, n, add())
		} else if consume(TokenGreaterEqual) {
			n = NewBinaryNode(NodeGreaterEql, n, add())
		} else {
			return n
		}
	}
}

func add() *Node {
	var n = mul()
	for {
		if consume(TokenPlus) {
			n = NewBinaryNode(NodeAdd, n, mul())
		} else if consume(TokenMinus) {
			n = NewBinaryNode(NodeSub, n, mul())
		} else {
			return n
		}
	}
}

func mul() *Node {
	var n = unary()
	for {
		if consume(TokenStar) {
			n = NewBinaryNode(NodeMul, n, unary())
		} else if consume(TokenSlash) {
			n = NewBinaryNode(NodeDiv, n, unary())
		} else {
			return n
		}
	}
}

func unary() *Node {
	if consume(TokenPlus) {
		return primary()
	}
	if consume(TokenMinus) {
		return NewBinaryNode(NodeSub, NewNodeNum(0), primary())
	}
	if consume(TokenStar) {
		return NewNode(NodeDeref, []*Node{unary()})
	}
	if consume(TokenAmpersand) {
		return NewNode(NodeAddr, []*Node{unary()})
	}
	return primary()
}

func primary() *Node {
	// 次のトークンが "(" なら、"(" expr ")" のはず
	if consume(TokenLparen) {
		var n = expr()
		expect(TokenRparen)
		return n
	}

	if currentToken().kind != TokenIdentifier {
		return NewNodeNum(expectNumber())
	}

	if tokenizer.Prefetch(1).Test(TokenLparen) {
		// 関数呼び出し
		var tok = expectIdentifier()
		expect(TokenLparen)
		var node = NewNode(NodeFunctionCall, make([]*Node, 0))
		node.label = tok.str
		for !consume(TokenRparen) {
			if len(node.children) > 0 {
				expect(TokenComma)
			}
			node.children = append(node.children, expr())
		}
		return node
	}
	return variableRef()
}

func variableRef() *Node {
	var tok = expectIdentifier()
	var node = NewLeafNode(NodeLocalVar)
	node.lvar = Env.FindLocalVar(currentFuncLabel, tok)
	if node.lvar == nil {
		errorAt(tok.rest, "未定義の変数です %s", tok.str)
	}
	return node
}

func variableDeclaration() *Node {
	var tok = expectIdentifier()
	var node = NewLeafNode(NodeLocalVar)
	lvar := Env.FindLocalVar(currentFuncLabel, tok)
	if lvar != nil {
		errorAt(tok.rest, "すでに定義済みの変数です %s", tok.str)
	}
	node.lvar = Env.AddLocalVar(currentFuncLabel, tok)
	return node
}
