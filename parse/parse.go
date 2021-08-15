package parse

import (
	"fmt"
	"os"
	"strings"

	"github.com/myuu222/myuugo/lang"
	"github.com/myuu222/myuugo/util"
)

var tokenizer *Tokenizer
var userInput string
var filename string

// エラーの起きた場所を報告するための関数
// 下のようなフォーマットでエラーメッセージを表示する
//
// foo.c:10: x = y + + 5;
//                   ^ 式ではありません
func errorAt(rest string, message string) {
	// 行番号と、restがその行の何番目から始まるかを見つける
	var lineNumber = 1
	var startIndex = 0
	for _, c := range userInput[:len(userInput)-len(rest)] {
		if c == '\n' {
			lineNumber += 1
			startIndex = 0
		} else if c == '\t' {
			startIndex += 4 // タブは空白4文字扱いとする
		} else {
			startIndex += 1
		}
	}
	for i, line := range strings.Split(userInput, "\n") {
		if i+1 == lineNumber {
			// 見つかった行をファイル名と行番号と一緒に表示
			var indent, _ = fmt.Fprintf(os.Stderr, "%s:%d: ", filename, lineNumber)
			fmt.Fprintln(os.Stderr, line)
			fmt.Fprintf(os.Stderr, "%*s^ %s\n", indent+startIndex, " ", message)
		}
	}
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
		util.Alarm("文の終端記号ではありません")
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
		errorAt(token.rest, "識別子ではありません")
	}
	return token
}

// 次のトークンが数値の場合、トークンを1つ読み進めてその数値を返す。
// それ以外の場合にはエラーを報告する。
func (t *Tokenizer) expectNumber() int {
	token := t.Fetch()
	if !t.Test(TokenNumber) {
		errorAt(token.rest, "数ではありません")
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
		errorAt(token.rest, "真偽値ではありません")
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
		errorAt(token.rest, "文字列ではありません")
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
		util.Alarm("型ではありません")
	}
	return ty
}

func (t *Tokenizer) consumeType() (lang.Type, bool) {
	var varType lang.Type = lang.Type{}
	if t.Consume(TokenStar) {
		ty := t.expectType()
		return lang.NewPointerType(&ty), true
	}
	if t.Consume(TokenLSBrace) {
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
	return varType, true
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

	Env.program.Code = append(Env.program.Code, topLevelStmtList().Children...)
	return Env.program
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
			errorAt(tokenizer.Fetch().rest, "文の区切り文字が必要です")
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
			errorAt(tokenizer.Fetch().rest, "文の区切り文字が必要です")
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

func topLevelStmt() *Node {
	// 関数定義
	if tokenizer.Test(TokenFunc) {
		return funcDefinition()
	}
	// var文
	if tokenizer.Test(TokenVar) {
		return topLevelVarStmt()
	}

	// 許可されていないもの
	if tokenizer.Test(TokenIf) {
		util.Alarm("if文はトップレベルでは使用できません")
	}
	if tokenizer.Test(TokenFor) {
		util.Alarm("for文はトップレベルでは使用できません")
	}
	if tokenizer.Test(TokenReturn) {
		util.Alarm("return文はトップレベルでは使用できません")
	}
	panic("トップレベルの文として許可されていません")
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

	var s = simpleStmt()
	if tokenizer.Consume(TokenLbrace) {
		// while文
		if s.Kind != NodeExprStmt {
			util.Alarm("for文の条件に式以外が書かれています")
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
		errorAt(token.rest, "'"+string(TokenIf)+"'ではありません")
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
	return equality()
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
	return primary()
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
	if tokenizer.Prefetch(1).Test(TokenLSBrace) {
		// 添字アクセス
		var arr = variableRef()
		tokenizer.Expect(TokenLSBrace)
		var index = expr()
		tokenizer.Expect(TokenRSBrace)
		return NewIndexNode(arr, index)
	}
	return variableRef()
}

func variableRef() *Node {
	var tok = tokenizer.expectIdentifier()
	var node = NewLeafNode(NodeVariable)
	node.Variable = Env.FindVar(tok.str)
	if node.Variable == nil {
		errorAt(tok.rest, "未定義の変数です")
	}
	return node
}

func localVariableDeclaration() *Node {
	var tok = tokenizer.expectIdentifier()
	var node = NewLeafNode(NodeVariable)
	lvar := Env.FindLocalVar(tok.str)
	if lvar != nil {
		errorAt(tok.rest, "すでに定義済みの変数です")
	}
	node.Variable = Env.AddLocalVar(lang.NewUndefinedType(), tok.str)
	return node
}

func topLevelVariableDeclaration() *Node {
	var tok = tokenizer.expectIdentifier()
	var node = NewLeafNode(NodeVariable)
	lvar := Env.program.FindTopLevelVariable(tok.str)
	if lvar != nil {
		errorAt(tok.rest, "すでに定義済みの変数です")
	}
	node.Variable = Env.program.AddTopLevelVariable(lang.NewUndefinedType(), tok.str)
	return node
}
