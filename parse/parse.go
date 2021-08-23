package parse

import (
	"github.com/myuu222/myuugo/lang"
)

var tokenizer *Tokenizer
var userInput string
var filename string

// トークナイザ拡張

// 文の終端記号であるトークンを1つ読み進めて真を返す。
// それ以外の場合には偽を返す。
func (t *Tokenizer) consumeEndLine() bool {
	return t.Consume(TokenSemicolon) || t.Consume(TokenNewLine)
}

func (t *Tokenizer) expectEndLine() {
	if !t.consumeEndLine() {
		BadToken(t.Fetch(), "文の終端記号ではありません")
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
		BadToken(token, "識別子ではありません")
	}
	return token
}

// 次のトークンが数値の場合、トークンを1つ読み進めてその数値を返す。
// それ以外の場合にはエラーを報告する。
func (t *Tokenizer) expectNumber() int {
	token := t.Fetch()
	if !t.Test(TokenNumber) {
		BadToken(token, "数ではありません")
	}
	var val = token.val
	tokenizer.Succ()
	return val
}

// 次のトークンが真偽値の場合、トークンを1つ読み進めて1か0を返す。
// それ以外の場合にはエラーを報告する。
func (t *Tokenizer) expectBool() int {
	token := t.Fetch()
	if !t.Test(TokenBool) {
		BadToken(token, "真偽値ではありません")
	}
	var val = token.val
	tokenizer.Succ()
	return val
}

// 次のトークンが文字列の場合、トークンを1つ読み進めてその文字列を返す。
// それ以外の場合にはエラーを報告する。
func (t *Tokenizer) expectString() string {
	token := t.Fetch()
	if !t.Test(TokenString) {
		BadToken(token, "文字列ではありません")
	}
	var val = token.str
	tokenizer.Succ()
	return val
}

func (t *Tokenizer) atEof() bool {
	return t.Test(TokenEof)
}

func (t *Tokenizer) expectType() lang.Type {
	ty, ok := t.consumeType()
	if !ok {
		BadToken(t.Fetch(), "型ではありません")
	}
	return ty
}

func (t *Tokenizer) consumeType() (lang.Type, bool) {
	if t.Consume(TokenStar) {
		ty := t.expectType()
		return lang.NewPointerType(&ty), true
	}
	if t.Consume(TokenLSBrace) {
		if t.Consume(TokenRSBrace) {
			// スライス
			ty := t.expectType()
			return lang.NewSliceType(ty), true
		}
		var arraySize = t.expectNumber()
		t.Expect(TokenRSBrace)
		ty := t.expectType()
		return lang.NewArrayType(ty, arraySize), true
	}
	tok, ok := t.consumeIdentifier()
	if !ok {
		return lang.Type{}, false
	}
	if tok.str == "int" {
		return lang.NewType(lang.TypeInt), true
	}
	if tok.str == "rune" {
		return lang.NewType(lang.TypeRune), true
	}
	if tok.str == "bool" {
		return lang.NewType(lang.TypeBool), true
	}
	if tok.str == "string" {
		var r = lang.NewType(lang.TypeRune)
		return lang.NewPointerType(&r), true
	}
	if tok.str == "struct" {
		tokenizer.Expect(TokenLbrace)
		tokenizer.Expect(TokenNewLine)
		names, types := []string{}, []lang.Type{}
		for !tokenizer.Consume(TokenRbrace) {
			name := tokenizer.expectIdentifier().str
			ty := tokenizer.expectType()
			tokenizer.Expect(TokenNewLine)

			names = append(names, name)
			types = append(types, ty)
		}
		return lang.NewStructType(names, types), true
	}
	return Env.program.FindType(tok.str)
}

var Env *Environment

func stepIn() {
	Env = Env.Fork()
}

func stepInFunction(name string) {
	Env = Env.Fork()
	Env.FunctionName = name
}

func stepOut() {
	Env = Env.parent
}

func Parse(path string) *Program {
	tokenizer = NewTokenizer()
	tokenizer.Tokenize(path)
	Env = NewEnvironment()

	for tokenizer.consumeEndLine() {
	}
	Env.program.Code = []*Node{packageStmt()}
	tokenizer.expectEndLine()

	for {
		for tokenizer.consumeEndLine() {
		}
		if tokenizer.Test(TokenImport) {
			Env.program.Code = append(Env.program.Code, importStmt())
		} else {
			break
		}
	}

	Env.program.Code = append(Env.program.Code, topLevelStmtList().Children...)
	return Env.program
}

func importStmt() *Node {
	tokenizer.Expect(TokenImport)
	packages := []string{}

	if tokenizer.Consume(TokenLparen) {
		// グループ化
		tokenizer.Expect(TokenNewLine)

		for !tokenizer.Consume(TokenRparen) {
			pkg := tokenizer.expectString()
			packages = append(packages, pkg)
			tokenizer.Expect(TokenNewLine)

			Env.program.AddPackageToImport(pkg)
		}
		return NewImportStmtNode(packages)
	}
	packages = append(packages, tokenizer.expectString())
	Env.program.AddPackageToImport(packages[0])
	return NewImportStmtNode(packages)
}

func packageStmt() *Node {
	var n = NewLeafNode(NodePackageStmt)

	tokenizer.Expect(TokenPackage)
	n.Label = tokenizer.expectIdentifier().str

	return n
}

func localStmtList() *Node {
	var stmts = make([]*Node, 0)
	var endLineRequired = false

	for !(tokenizer.Test(TokenRbrace)) {
		if endLineRequired {
			BadToken(tokenizer.Fetch(), "文の区切り文字が必要です")
		}
		if tokenizer.consumeEndLine() {
			continue
		}
		stmts = append(stmts, localStmt())

		endLineRequired = true
		if tokenizer.consumeEndLine() {
			endLineRequired = false
		}
	}
	var node = NewNode(NodeStmtList, stmts)
	node.Children = stmts
	return node
}

func topLevelStmtList() *Node {
	var stmts = make([]*Node, 0)
	var endLineRequired = false

	for !tokenizer.atEof() && !(tokenizer.Test(TokenRbrace)) {
		if endLineRequired {
			BadToken(tokenizer.Fetch(), "文の区切り文字が必要です")
		}
		if tokenizer.consumeEndLine() {
			continue
		}
		stmts = append(stmts, topLevelStmt())

		endLineRequired = true
		if tokenizer.consumeEndLine() {
			endLineRequired = false
		}
	}
	var node = NewNode(NodeStmtList, stmts)
	node.Children = stmts
	return node
}

func typeStmt() *Node {
	tokenizer.Expect(TokenType)
	typeName := tokenizer.expectIdentifier().str
	entityType := tokenizer.expectType()
	definedType := lang.NewUserDefinedType(typeName, entityType)

	Env.program.RegisterType(definedType)

	return NewNode(NodeTypeStmt, []*Node{})
}

func topLevelStmt() *Node {
	// 関数定義
	if tokenizer.Test(TokenFunc) {
		return funcDefinition()
	}
	// var文
	if tokenizer.Test(TokenVar) {
		return topLevelVarStmt()
	}
	// typ文
	if tokenizer.Test(TokenType) {
		return typeStmt()
	}

	// 許可されていないもの
	if tokenizer.Test(TokenIf) {
		BadToken(tokenizer.Fetch(), "if文はトップレベルでは使用できません")
	}
	if tokenizer.Test(TokenFor) {
		BadToken(tokenizer.Fetch(), "for文はトップレベルでは使用できません")
	}
	if tokenizer.Test(TokenReturn) {
		BadToken(tokenizer.Fetch(), "return文はトップレベルでは使用できません")
	}
	BadToken(tokenizer.Fetch(), "トップレベルの文として許可されていません")
	return nil // 到達しない
}

func simpleStmt() *Node {
	if tokenizer.Test(TokenNewLine) || tokenizer.Test(TokenSemicolon) {
		return nil
	}

	var pos = 0
	var nxtToken = tokenizer.Prefetch(pos)
	for !nxtToken.Test(TokenNewLine) && !nxtToken.Test(TokenSemicolon) {
		if tokenizer.Prefetch(pos).Test(TokenEqual) {
			// 代入文としてパース
			var n = exprList()
			tokenizer.Expect(TokenEqual)
			return NewBinaryNode(NodeAssign, n, exprList())
		}
		if tokenizer.Prefetch(pos).Test(TokenColonEqual) {
			// 短絡変数宣言としてパース
			var n = localVarList()
			tokenizer.Expect(TokenColonEqual)
			return NewBinaryNode(NodeShortVarDeclStmt, n, exprList())
		}
		pos += 1
		nxtToken = tokenizer.Prefetch(pos)
	}
	return NewNode(NodeExprStmt, []*Node{expr()})
}

func localStmt() *Node {
	// if文
	if tokenizer.Test(TokenIf) {
		return metaIfStmt()
	}
	// for文
	if tokenizer.Test(TokenFor) {
		return forStmt()
	}
	// var文
	if tokenizer.Test(TokenVar) {
		return localVarStmt()
	}
	if tokenizer.Consume(TokenReturn) {
		if tokenizer.Test(TokenNewLine) || tokenizer.Test(TokenSemicolon) {
			// 空のreturn文
			return NewUnaryOperationNode(NodeReturn, nil)
		}
		return NewUnaryOperationNode(NodeReturn, exprList())
	}
	return simpleStmt()
}

// トップレベル変数は初期化式は与えないことにする
func topLevelVarStmt() *Node {
	tokenizer.Expect(TokenVar)
	var v = topLevelVariableDeclaration()
	v.Variable.Type = tokenizer.expectType()
	return NewNode(NodeTopLevelVarStmt, []*Node{v})
}

func localVarStmt() *Node {
	tokenizer.Expect(TokenVar)
	var v = localVariableDeclaration()
	ty, ok := tokenizer.consumeType()

	if !ok {
		// 型が明示されていないときは初期化が必須
		tokenizer.Expect(TokenEqual)
		return NewBinaryNode(NodeLocalVarStmt, v, expr())
	} else {
		v.Variable.Type = ty
	}
	if tokenizer.Consume(TokenEqual) {
		return NewBinaryNode(NodeLocalVarStmt, v, expr())
	}
	return NewNode(NodeLocalVarStmt, []*Node{v})
}

func funcDefinition() *Node {
	tokenizer.Expect(TokenFunc)
	identifier := tokenizer.expectIdentifier()

	stepInFunction(identifier.str)
	var fn = lang.NewFunction(Env.FunctionName, []lang.Type{}, lang.NewUndefinedType())
	Env.program.RegisterFunction(fn)

	var parameters = make([]*Node, 0)

	tokenizer.Expect(TokenLparen)
	for !tokenizer.Consume(TokenRparen) {
		if len(parameters) > 0 {
			tokenizer.Expect(TokenComma)
		}
		lvarNode := localVariableDeclaration()
		parameters = append(parameters, lvarNode)
		lvarNode.Variable.Type = tokenizer.expectType()
		fn.ParameterTypes = append(fn.ParameterTypes, lvarNode.Variable.Type)
	}

	fn.ReturnValueType = lang.NewType(lang.TypeVoid)
	if tokenizer.Consume(TokenLparen) { // 多値
		var types = []lang.Type{tokenizer.expectType()}
		for tokenizer.Consume(TokenComma) {
			types = append(types, tokenizer.expectType())
		}
		tokenizer.Expect(TokenRparen)
		fn.ReturnValueType = lang.NewMultipleType(types)
	} else {
		var ty, ok = tokenizer.consumeType()
		if ok {
			fn.ReturnValueType = ty
		}
	}
	tokenizer.Expect(TokenLbrace)

	var functionName = identifier.str
	var body = localStmtList()

	tokenizer.Expect(TokenRbrace)

	var node = NewFunctionDefNode(functionName, parameters, body)

	stepOut()

	return node
}

// range は未対応
func forStmt() *Node {
	stepIn()
	tokenizer.Expect(TokenFor)
	// 初期化, ループ条件, 更新式, 繰り返す文

	if tokenizer.Consume(TokenLbrace) {
		// 無限ループ
		var body = localStmtList()
		tokenizer.Expect(TokenRbrace)
		stepOut()
		return NewForNode(nil, nil, nil, body)
	}

	var st = tokenizer.Fetch()
	var s = simpleStmt()
	if tokenizer.Consume(TokenLbrace) {
		// while文
		if s.Kind != NodeExprStmt {
			BadToken(st, "for文の条件に式以外が書かれています")
		}
		var cond = s.Children[0] // expr
		var body = localStmtList()
		tokenizer.Expect(TokenRbrace)
		stepOut()
		return NewForNode(nil, cond, nil, body)
	}

	// 通常のfor文
	var init = s
	tokenizer.Expect(TokenSemicolon)
	var cond = expr()
	tokenizer.Expect(TokenSemicolon)
	var update = simpleStmt()

	tokenizer.Expect(TokenLbrace)
	var body = localStmtList()
	tokenizer.Expect(TokenRbrace)
	stepOut()
	return NewForNode(init, cond, update, body)
}

func metaIfStmt() *Node {
	token := tokenizer.Fetch()
	if !token.Test(TokenIf) {
		BadToken(token, "'"+string(TokenIf)+"'ではありません")
	}

	var ifNode = ifStmt()
	if tokenizer.Test(TokenElse) {
		var elseNode = elseStmt()
		return NewMetaIfNode(ifNode, elseNode)
	}
	return NewMetaIfNode(ifNode, nil)
}

func ifStmt() *Node {
	stepIn()

	tokenizer.Expect(TokenIf)
	var cond = expr()
	tokenizer.Expect(TokenLbrace)
	var body = localStmtList()
	tokenizer.Expect(TokenRbrace)

	stepOut()
	return NewIfNode(cond, body)
}

func elseStmt() *Node {
	tokenizer.Expect(TokenElse)

	if tokenizer.Consume(TokenLbrace) {
		stepIn()
		var body = localStmtList()
		tokenizer.Expect(TokenRbrace)
		stepOut()
		return NewElseNode(body)
	}
	return metaIfStmt()
}

func localVarList() *Node {
	var lvars = []*Node{localVariableDeclaration()}
	for tokenizer.Consume(TokenComma) {
		lvars = append(lvars, localVariableDeclaration())
	}
	return NewNode(NodeLocalVarList, lvars)
}

func exprList() *Node {
	var exprs = []*Node{expr()}

	for tokenizer.Consume(TokenComma) {
		exprs = append(exprs, expr())
	}
	return NewNode(NodeExprList, exprs)
}

func expr() *Node {
	return logicalOr()
}

func logicalOr() *Node {
	var n = logicalAnd()
	for {
		if tokenizer.Consume(TokenDoubleVerticalLine) {
			n = NewBinaryOperationNode(NodeLogicalOr, n, logicalAnd())
		} else {
			return n
		}
	}
}

func logicalAnd() *Node {
	var n = equality()
	for {
		if tokenizer.Consume(TokenDoubleAmpersand) {
			n = NewBinaryOperationNode(NodeLogicalAnd, n, equality())
		} else {
			return n
		}
	}
}

func equality() *Node {
	var n = relational()
	for {
		if tokenizer.Consume(TokenDoubleEqual) {
			n = NewBinaryOperationNode(NodeEql, n, relational())
		} else if tokenizer.Consume(TokenNotEqual) {
			n = NewBinaryOperationNode(NodeNotEql, n, relational())
		} else {
			return n
		}
	}
}

func relational() *Node {
	var n = add()
	for {
		if tokenizer.Consume(TokenLess) {
			n = NewBinaryOperationNode(NodeLess, n, add())
		} else if tokenizer.Consume(TokenLessEqual) {
			n = NewBinaryOperationNode(NodeLessEql, n, add())
		} else if tokenizer.Consume(TokenGreater) {
			n = NewBinaryOperationNode(NodeGreater, n, add())
		} else if tokenizer.Consume(TokenGreaterEqual) {
			n = NewBinaryOperationNode(NodeGreaterEql, n, add())
		} else {
			return n
		}
	}
}

func add() *Node {
	var n = mul()
	for {
		if tokenizer.Consume(TokenPlus) {
			n = NewBinaryOperationNode(NodeAdd, n, mul())
		} else if tokenizer.Consume(TokenMinus) {
			n = NewBinaryOperationNode(NodeSub, n, mul())
		} else {
			return n
		}
	}
}

func mul() *Node {
	var n = unary()
	for {
		if tokenizer.Consume(TokenStar) {
			n = NewBinaryOperationNode(NodeMul, n, unary())
		} else if tokenizer.Consume(TokenSlash) {
			n = NewBinaryOperationNode(NodeDiv, n, unary())
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
		return NewBinaryOperationNode(NodeSub, NewNodeNum(0), primary())
	}
	if tokenizer.Consume(TokenStar) {
		return NewUnaryOperationNode(NodeDeref, unary())
	}
	if tokenizer.Consume(TokenAmpersand) {
		return NewUnaryOperationNode(NodeAddr, unary())
	}
	if tokenizer.Consume(TokenBang) {
		return NewUnaryOperationNode(NodeNot, unary())
	}
	return primary()
}

func structTypeLiteral() *Node {
	tok := tokenizer.expectIdentifier()
	ty, ok := Env.program.FindType(tok.str)
	if !ok {
		BadToken(tok, "未定義の型のリテラルです")
	}
	names, values := []string{}, []*Node{}
	tokenizer.Expect(TokenLbrace)
	for !tokenizer.Consume(TokenRbrace) {
		if len(names) > 0 {
			tokenizer.Expect(TokenComma)
		}
		names = append(names, tokenizer.expectIdentifier().str)
		tokenizer.Expect(TokenColon)
		values = append(values, expr())
	}
	return NewStructLiteral(ty, names, values)
}

func primary() *Node {
	// 次のトークンが "(" なら、"(" expr ")" のはず
	if tokenizer.Consume(TokenLparen) {
		var n = expr()
		tokenizer.Expect(TokenRparen)
		return n
	}
	if tokenizer.Test(TokenNumber) {
		return NewNodeNum(tokenizer.expectNumber())
	}
	if tokenizer.Test(TokenBool) {
		return NewNodeBool(tokenizer.expectBool())
	}
	if tokenizer.Test(TokenString) {
		var n = NewLeafNode(NodeString)
		n.Str = Env.program.AddStringLiteral(tokenizer.Fetch().str)
		tokenizer.Succ()
		return n
	}

	if tokenizer.Test(TokenLSBrace) {
		ty := tokenizer.expectType()

		if ty.Kind == lang.TypeSlice {
			elements := []*Node{}
			tokenizer.Expect(TokenLbrace)
			for !tokenizer.Consume(TokenRbrace) {
				if len(elements) > 0 {
					tokenizer.Expect(TokenComma)
				}
				elements = append(elements, expr())
			}
			return NewSliceLiteral(ty, elements)
		}
		panic("未実装の型のリテラルです")
	}

	var tok = tokenizer.Fetch()
	ty, ok := Env.program.FindType(tok.str)
	// struct型のリテラル
	if ok {
		names, values := []string{}, []*Node{}
		tokenizer.expectType()
		tokenizer.Expect(TokenLbrace)
		for !tokenizer.Consume(TokenRbrace) {
			if len(names) > 0 {
				tokenizer.Expect(TokenComma)
			}
			names = append(names, tokenizer.expectIdentifier().str)
			tokenizer.Expect(TokenColon)
			values = append(values, expr())
		}
		return NewStructLiteral(ty, names, values)
	}

	// append関数の呼び出し
	if tokenizer.Fetch().str == "append" && tokenizer.Prefetch(1).Test(TokenLparen) {
		tokenizer.Expect(TokenIdentifier)
		tokenizer.Expect(TokenLparen)
		var arg1 = expr()
		tokenizer.Expect(TokenComma)
		var arg2 = expr()
		tokenizer.Expect(TokenRparen)
		return NewAppendCallNode(arg1, arg2)
	}

	var name = tokenizer.Fetch().str
	_, ok = Env.program.FindPackageToImport(name)

	if ok {
		tokenizer.expectIdentifier()
		tokenizer.Expect(TokenDot)
	}

	var n *Node = named()
	for {
		if tokenizer.Consume(TokenLSBrace) {
			n = NewIndexNode(n, expr())
			tokenizer.Expect(TokenRSBrace)
			continue
		}
		if tokenizer.Consume(TokenDot) {
			// メソッド呼び出しは一旦無視
			name := tokenizer.expectIdentifier()
			n = NewDotNode(n, name.str)
			continue
		}
		break
	}
	return n
}

func named() *Node {
	if tokenizer.Prefetch(1).Test(TokenLparen) {
		// 関数呼び出し
		var tok = tokenizer.expectIdentifier()
		tokenizer.Expect(TokenLparen)
		var functionName = tok.str
		var arguments = []*Node{}
		for !tokenizer.Consume(TokenRparen) {
			if len(arguments) > 0 {
				tokenizer.Expect(TokenComma)
			}
			arguments = append(arguments, expr())
		}
		return NewFunctionCallNode(functionName, arguments)
	}
	return variableRef()
}

func variableRef() *Node {
	var tok = tokenizer.expectIdentifier()
	var node = NewLeafNode(NodeVariable)
	node.Variable = Env.FindVar(tok.str)
	if node.Variable == nil {
		BadToken(tok, "未定義の変数です")
	}
	return node
}

func localVariableDeclaration() *Node {
	var tok = tokenizer.expectIdentifier()
	var node = NewLeafNode(NodeVariable)
	lvar := Env.FindLocalVar(tok.str)
	if lvar != nil {
		BadToken(tok, "すでに定義済みの変数です")
	}
	node.Variable = Env.AddLocalVar(lang.NewUndefinedType(), tok.str)
	return node
}

func topLevelVariableDeclaration() *Node {
	var tok = tokenizer.expectIdentifier()
	var node = NewLeafNode(NodeVariable)
	lvar := Env.program.FindTopLevelVariable(tok.str)
	if lvar != nil {
		BadToken(tok, "すでに定義済みの変数です")
	}
	node.Variable = Env.program.AddTopLevelVariable(lang.NewUndefinedType(), tok.str)
	return node
}
