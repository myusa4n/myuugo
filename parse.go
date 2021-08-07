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

// トークナイザ拡張

// 文の終端記号であるトークンを1つ読み進めて真を返す。
// それ以外の場合には偽を返す。
func (t *Tokenizer) consumeEndLine() bool {
	return t.Consume(TokenSemicolon) || t.Consume(TokenNewLine)
}

func (t *Tokenizer) expectEndLine() {
	if !t.consumeEndLine() {
		madden("文の終端記号ではありません")
	}
}

// 次のトークンが識別子の時には、トークンを1つ読み進めてそのトークンを返す。
// この時、返り値の二番目の値は真になる。
// 逆に識別子でない場合は、偽になる。
func (t *Tokenizer) consumeIdentifier() (Token, bool) {
	token := t.Fetch()
	if token.Test(TokenIdentifier) {
		tokenizer.Succ()
		return token, true
	}
	return Token{}, false
}

// 次のトークンが識別子の時には、トークンを1つ読み進めてそのトークンを返す。
// そうでない場合はエラーを報告する。
func (t *Tokenizer) expectIdentifier() Token {
	token, ok := t.consumeIdentifier()
	if !ok {
		errorAt(token.str, "識別子ではありません")
	}
	return token
}

// 次のトークンが数値の場合、トークンを1つ読み進めてその数値を返す。
// それ以外の場合にはエラーを報告する。
func (t *Tokenizer) expectNumber() int {
	token := t.Fetch()
	if !t.Test(TokenNumber) {
		errorAt(token.str, "数ではありません")
	}
	var val = token.val
	tokenizer.Succ()
	return val
}

func (t *Tokenizer) atEof() bool {
	return t.Test(TokenEof)
}

func (t *Tokenizer) expectType() Type {
	var varType Type = Type{}
	if t.Consume(TokenStar) {
		varType.kind = TypePtr
		ty := t.expectType()
		varType.ptrTo = &ty
		return varType
	}
	tok := t.expectIdentifier()
	if tok.str == "int" {
		return Type{kind: TypeInt}
	}
	return varType
}

func (t *Tokenizer) consumeType() (Type, bool) {
	var varType Type = Type{}
	if t.Consume(TokenStar) {
		varType.kind = TypePtr
		ty := t.expectType()
		varType.ptrTo = &ty
		return varType, true
	}
	tok, ok := t.consumeIdentifier()
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

	for tokenizer.consumeEndLine() {
	}
	code = []*Node{packageStmt()}
	tokenizer.expectEndLine()

	code = append(code, stmtList().children...)
}

func packageStmt() *Node {
	var n = NewLeafNode(NodePackageStmt)

	tokenizer.Expect(TokenPackage)
	n.label = tokenizer.expectIdentifier().str

	return n
}

func stmtList() *Node {
	var stmts = make([]*Node, 0)
	var endLineRequired = false

	for !tokenizer.atEof() && !(tokenizer.Test(TokenRbrace)) {
		if endLineRequired {
			errorAt(tokenizer.Fetch().rest, "文の区切り文字が必要です")
		}
		if tokenizer.consumeEndLine() {
			continue
		}
		stmts = append(stmts, stmt())

		endLineRequired = true
		if tokenizer.consumeEndLine() {
			endLineRequired = false
		}
	}
	var node = NewNode(NodeStmtList, stmts)
	node.children = stmts
	return node
}

func stmt() *Node {
	// if文
	if tokenizer.Test(TokenIf) {
		return metaIfStmt()
	}
	// for文
	if tokenizer.Test(TokenFor) {
		return forStmt()
	}
	// 関数定義
	if tokenizer.Test(TokenFunc) {
		return funcDefinition()
	}
	// var文
	if tokenizer.Test(TokenVar) {
		return varStmt()
	}

	if tokenizer.Consume(TokenReturn) {
		return NewNode(NodeReturn, []*Node{expr()})
	} else {
		var n = expr()
		if tokenizer.Consume(TokenEqual) {
			// 代入文
			var e = expr()
			return NewBinaryNode(NodeAssign, n, e)
		}
		return NewNode(NodeExprStmt, []*Node{n})
	}
}

func varStmt() *Node {
	tokenizer.Expect(TokenVar)
	var v = variableDeclaration()
	ty, ok := tokenizer.consumeType()

	if !ok {
		// 型が明示されていないときは初期化が必須
		tokenizer.Expect(TokenEqual)
		return NewBinaryNode(NodeVarStmt, v, expr())
	} else {
		v.lvar.varType = ty
	}
	if tokenizer.Consume(TokenEqual) {
		return NewBinaryNode(NodeVarStmt, v, expr())
	}
	return NewNode(NodeVarStmt, []*Node{v})
}

func funcDefinition() *Node {
	tokenizer.Expect(TokenFunc)
	identifier := tokenizer.expectIdentifier()

	var prevFuncLabel = currentFuncLabel
	currentFuncLabel = identifier.str
	var fn = Env.RegisterFunc(currentFuncLabel)

	var parameters = make([]*Node, 0)

	tokenizer.Expect(TokenLparen)
	for !tokenizer.Consume(TokenRparen) {
		if len(parameters) > 0 {
			tokenizer.Expect(TokenComma)
		}
		lvarNode := variableDeclaration()
		parameters = append(parameters, lvarNode)
		lvarNode.lvar.varType = tokenizer.expectType()
		fn.ParameterTypes = append(fn.ParameterTypes, lvarNode.lvar.varType)
	}

	// 本当はvoid型が正しいけれど、テストを簡単にするためしばらくはint型で定義
	fn.ReturnValueType = NewType(TypeInt)
	var ty, ok = tokenizer.consumeType()
	if ok {
		fn.ReturnValueType = ty
	}

	tokenizer.Expect(TokenLbrace)

	var node = NewNode(NodeFunctionDef, make([]*Node, 0))
	node.label = identifier.str
	node.children = append(node.children, stmtList())
	node.children = append(node.children, parameters...)

	tokenizer.Expect(TokenRbrace)

	currentFuncLabel = prevFuncLabel

	return node
}

// range は未対応
func forStmt() *Node {
	tokenizer.Expect(TokenFor)
	// 初期化, ループ条件, 更新式, 繰り返す文
	var node = NewNode(NodeFor, []*Node{nil, nil, nil, nil})

	if tokenizer.Consume(TokenLbrace) {
		// 無限ループ
		node.children[3] = stmtList()
		tokenizer.Expect(TokenRbrace)
		return node
	}

	var s = stmt()
	if tokenizer.Consume(TokenLbrace) {
		// while文
		if s.kind != NodeExprStmt {
			madden("for文の条件に式以外が書かれています")
		}
		node.children[1] = s.children[0] // expr
		node.children[3] = stmtList()
		tokenizer.Expect(TokenRbrace)
		return node
	}

	// 通常のfor文
	node.children[0] = s
	tokenizer.Expect(TokenSemicolon)
	node.children[1] = stmt().children[0] // expr
	tokenizer.Expect(TokenSemicolon)
	node.children[2] = stmt()

	tokenizer.Expect(TokenLbrace)
	node.children[3] = stmtList()
	tokenizer.Expect(TokenRbrace)
	return node
}

func metaIfStmt() *Node {
	token := tokenizer.Fetch()
	if !token.Test(TokenIf) {
		errorAt(token.str, "'%s'ではありません", TokenIf)
	}

	var ifNode = ifStmt()
	if tokenizer.Test(TokenElse) {
		var elseNode = elseStmt()
		return NewBinaryNode(NodeMetaIf, ifNode, elseNode)
	}
	return NewBinaryNode(NodeMetaIf, ifNode, nil)
}

func ifStmt() *Node {
	tokenizer.Expect(TokenIf)
	var lhs = expr()
	tokenizer.Expect(TokenLbrace)
	var rhs = stmtList()
	tokenizer.Expect(TokenRbrace)
	return NewBinaryNode(NodeIf, lhs, rhs)
}

func elseStmt() *Node {
	tokenizer.Expect(TokenElse)
	tokenizer.Expect(TokenLbrace)
	var stmts = stmtList()
	tokenizer.Expect(TokenRbrace)
	return NewNode(NodeElse, []*Node{stmts})
}

func expr() *Node {
	return equality()
}

func equality() *Node {
	var n = relational()
	for {
		if tokenizer.Consume(TokenDoubleEqual) {
			n = NewBinaryNode(NodeEql, n, relational())
		} else if tokenizer.Consume(TokenNotEqual) {
			n = NewBinaryNode(NodeNotEql, n, relational())
		} else {
			return n
		}
	}
}

func relational() *Node {
	var n = add()
	for {
		if tokenizer.Consume(TokenLess) {
			n = NewBinaryNode(NodeLess, n, add())
		} else if tokenizer.Consume(TokenLessEqual) {
			n = NewBinaryNode(NodeLessEql, n, add())
		} else if tokenizer.Consume(TokenGreater) {
			n = NewBinaryNode(NodeGreater, n, add())
		} else if tokenizer.Consume(TokenGreaterEqual) {
			n = NewBinaryNode(NodeGreaterEql, n, add())
		} else {
			return n
		}
	}
}

func add() *Node {
	var n = mul()
	for {
		if tokenizer.Consume(TokenPlus) {
			n = NewBinaryNode(NodeAdd, n, mul())
		} else if tokenizer.Consume(TokenMinus) {
			n = NewBinaryNode(NodeSub, n, mul())
		} else {
			return n
		}
	}
}

func mul() *Node {
	var n = unary()
	for {
		if tokenizer.Consume(TokenStar) {
			n = NewBinaryNode(NodeMul, n, unary())
		} else if tokenizer.Consume(TokenSlash) {
			n = NewBinaryNode(NodeDiv, n, unary())
		} else {
			return n
		}
	}
}

func unary() *Node {
	if tokenizer.Consume(TokenPlus) {
		return primary()
	}
	if tokenizer.Consume(TokenMinus) {
		return NewBinaryNode(NodeSub, NewNodeNum(0), primary())
	}
	if tokenizer.Consume(TokenStar) {
		return NewNode(NodeDeref, []*Node{unary()})
	}
	if tokenizer.Consume(TokenAmpersand) {
		return NewNode(NodeAddr, []*Node{unary()})
	}
	return primary()
}

func primary() *Node {
	// 次のトークンが "(" なら、"(" expr ")" のはず
	if tokenizer.Consume(TokenLparen) {
		var n = expr()
		tokenizer.Expect(TokenRparen)
		return n
	}

	if !tokenizer.Test(TokenIdentifier) {
		return NewNodeNum(tokenizer.expectNumber())
	}

	if tokenizer.Prefetch(1).Test(TokenLparen) {
		// 関数呼び出し
		var tok = tokenizer.expectIdentifier()
		tokenizer.Expect(TokenLparen)
		var node = NewNode(NodeFunctionCall, make([]*Node, 0))
		node.label = tok.str
		for !tokenizer.Consume(TokenRparen) {
			if len(node.children) > 0 {
				tokenizer.Expect(TokenComma)
			}
			node.children = append(node.children, expr())
		}
		return node
	}
	return variableRef()
}

func variableRef() *Node {
	var tok = tokenizer.expectIdentifier()
	var node = NewLeafNode(NodeLocalVar)
	node.lvar = Env.FindLocalVar(currentFuncLabel, tok)
	if node.lvar == nil {
		errorAt(tok.rest, "未定義の変数です %s", tok.str)
	}
	return node
}

func variableDeclaration() *Node {
	var tok = tokenizer.expectIdentifier()
	var node = NewLeafNode(NodeLocalVar)
	lvar := Env.FindLocalVar(currentFuncLabel, tok)
	if lvar != nil {
		errorAt(tok.rest, "すでに定義済みの変数です %s", tok.str)
	}
	node.lvar = Env.AddLocalVar(currentFuncLabel, tok)
	return node
}
