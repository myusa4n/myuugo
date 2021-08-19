package codegen

import (
	"strconv"

	"github.com/myuu222/myuugo/lang"
	"github.com/myuu222/myuugo/parse"
	"github.com/myuu222/myuugo/util"
)

var labelNumber = 0
var program *parse.Program

func declare(node *parse.Node) {
	var variable = node.Variable

	if variable.Kind == lang.VariableTopLevel {
		p(".data")
		p(variable.Name + ":")

		emit(".zero %d\n", entitySizeOf(variable.Type))
		p(".text")
		return
	}
	// 基本的に何もしないが配列または構造体の場合は動的にメモリを確保し、あらかじめ割り当てる
	if variable.Type.Kind == lang.TypeArray || variable.Type.Kind == lang.TypeUserDefined {
		genLvalue(node)
		emit("mov rdi, %d", entitySizeOf(variable.Type))
		emit("call malloc")
		emit("pop rdi")
		emit("mov [rdi], rax")
		return
	}
}

func assign(lhs *parse.Node, rhs *parse.Node) {
	if lhs.ExprType.Kind == lang.TypeArray {
		gen(lhs)
		gen(rhs)

		emit("pop rdi")
		emit("pop rax")

		var size = lang.Sizeof(*lhs.ExprType.PtrTo)

		for i := 0; i < lhs.ExprType.ArraySize; i++ {
			emit("mov r10, %s PTR [rdi+%d]", word(size), size*i)
			emit("mov %s PTR [rax+%d], r10", word(size), size*i)
		}
		return
	}

	genLvalue(lhs)
	gen(rhs)

	emit("pop rdi")
	emit("pop rax")

	emit("mov [rax], " + register(1, lang.Sizeof(lhs.ExprType)))
}

// 多値を返す関数の返り値を左辺にある複数の変数に代入する
func assignMultiple(lhss []*parse.Node, rhs *parse.Node) {
	gen(rhs)

	// 分解する
	// 右端の変数から代入されることになる
	for i := len(lhss) - 1; i >= 0; i-- {
		var l = lhss[i]
		genLvalue(l)

		emit("pop rax")
		emit("pop rdi")
		emit("mov [rax], " + register(1, lang.Sizeof(l.ExprType)))
	}
}

func genLvalue(node *parse.Node) {
	if node.Kind == parse.NodeDeref {
		gen(node.Target)
		return
	} else if node.Kind == parse.NodeVariable {
		var variable = node.Variable
		if variable.Kind == lang.VariableLocal {
			emit("mov rax, rbp")
			emit("sub rax, %d", node.Variable.Offset)
			emit("push rax")
		} else {
			emit("mov rax, OFFSET FLAT:%s", node.Variable.Name)
			emit("push rax")
		}
		return
	} else if node.Kind == parse.NodeIndex {
		gen(node.Seq)
		gen(node.Index)
		emit("pop rdi")
		emit("imul rdi, %d", lang.Sizeof(node.ExprType))

		if node.Seq.ExprType.Kind == lang.TypeSlice {
			emit("add rdi, 8") // 要素数を表す値のオフセットの分だけずらしておく
		}

		emit("pop rax")
		emit("add rax, rdi")
		emit("push rax")
		return
	} else if node.Kind == parse.NodeDot {
		gen(node.Owner)
		entityType := *node.Owner.ExprType.PtrTo

		for i := 0; i < len(entityType.MemberNames); i++ {
			if entityType.MemberNames[i] == node.MemberName {
				emit("pop rax")
				emit("add rax, %d", entityType.MemberOffsets[i])
				emit("push rax")
				return
			}
		}
		panic("到達しないはず")
	}
	util.Alarm("代入の左辺値が変数またはポインタ参照ではありません")
}

func gen(node *parse.Node) {
	if node.Kind == parse.NodePackageStmt {
		// 何もしない
		return
	}
	if node.Kind == parse.NodeTypeStmt {
		// 何もしない
		return
	}
	if node.Kind == parse.NodeNum {
		emit("push %d", node.Val)
		return
	}
	if node.Kind == parse.NodeBool {
		emit("push %d", node.Val)
		return
	}
	if node.Kind == parse.NodeStmtList {
		for _, stmt := range node.Children {
			gen(stmt)
		}
		return
	}
	if node.Kind == parse.NodeReturn {
		if node.Target != nil {
			var exprs = node.Target.Children
			for _, e := range exprs {
				gen(e)
			}

			for i := range exprs {
				emit("pop " + register(len(exprs)-i-1, 8))
			}
		} else {
			// void型
			emit("mov rax, 0")
		}
		emit("mov rsp, rbp")
		emit("pop rbp")
		emit("ret")
		return
	}
	if node.Kind == parse.NodeVariable {
		genLvalue(node)

		if node.Variable.Kind == lang.VariableTopLevel && node.ExprType.Kind == lang.TypeArray {
			return
		}

		emit("pop rax")

		if lang.Sizeof(node.ExprType) == 1 {
			emit("movzx rax, BYTE PTR [rax]")
		} else { // 8
			emit("mov rax, [rax]")
		}
		emit("push rax")
		return
	}
	if node.Kind == parse.NodeAssign {
		var lhs = node.Children[0]
		var rhs = node.Children[1]

		if rhs.ExprType.Kind == lang.TypeMultiple && len(rhs.Children) == 1 {
			assignMultiple(lhs.Children, rhs.Children[0])
			return
		}

		for i, l := range lhs.Children {
			r := rhs.Children[i]
			assign(l, r)
		}
		return
	}
	if node.Kind == parse.NodeMetaIf {
		var endLabel = ".Lend" + strconv.Itoa(labelNumber)
		var elseLabel = ".Lelse" + strconv.Itoa(labelNumber)

		gen(node.If)
		p("%s:", elseLabel)

		if node.Else != nil {
			gen(node.Else)
		}
		p("%s:", endLabel)
		return
	}
	if node.Kind == parse.NodeIf {
		var endLabel = ".Lend" + strconv.Itoa(labelNumber)
		var elseLabel = ".Lelse" + strconv.Itoa(labelNumber)
		labelNumber += 1

		gen(node.Condition)
		emit("pop rax")
		emit("cmp rax, 0")
		emit("je " + elseLabel)
		gen(node.Body)
		emit("jmp " + endLabel)
		return
	}
	if node.Kind == parse.NodeElse {
		gen(node.Body)
		return
	}
	if node.Kind == parse.NodeFor {
		var beginLabel = ".Lbegin" + strconv.Itoa(labelNumber)
		var endLabel = ".Lend" + strconv.Itoa(labelNumber)
		labelNumber += 1

		if node.Init != nil {
			gen(node.Init)
		}
		p("%s:", beginLabel)
		if node.Condition != nil {
			gen(node.Condition) // 条件
			emit("pop rax")
			emit("cmp rax, 0")
			emit("je " + endLabel)
		}
		gen(node.Body)
		if node.Update != nil {
			gen(node.Update)
		}
		emit("jmp %s", beginLabel)
		p("%s:", endLabel)
		return
	}
	if node.Kind == parse.NodeFunctionCall {
		// TODO: rune型と配列型の扱いについて考える
		for _, argument := range node.Arguments {
			gen(argument)
		}
		for i := range node.Arguments {
			// 配列や構造体は先頭のアドレスだけ渡しておいてNodeFunctionDef側でうまいこと代入してもらう
			emit("pop " + register(len(node.Arguments)-i, 8))
		}
		emit("mov al, 0") // 可変長引数の関数を呼び出すためのルール
		emit("call " + node.Label)

		// 今見ている関数が多値だった場合は、rax, rdi, rsi, ...から取り出していく
		fn := program.FindFunction(node.Label)
		if fn != nil && fn.ReturnValueType.Kind == lang.TypeMultiple {
			// raxから順にスタックに突っ込んでいく
			for i := range fn.ReturnValueType.Components {
				emit("push " + register(i, 8))
			}
			return
		}
		emit("push rax")
		return
	}
	if node.Kind == parse.NodeFunctionDef {
		p("%s:", node.Label)

		// プロローグ
		emit("push rbp")
		emit("mov rbp, rsp")

		emit("sub rsp, %d", getFrameSize(program, node.Label))

		for i, param := range node.Parameters { // 引数
			genLvalue(param)
			emit("pop rax")

			emit("mov [rax], " + register(i+1, lang.Sizeof(param.ExprType)))
		}

		gen(node.Body) // 関数本体

		// エピローグ
		// 関数の返り値の型が void 型だと仮定する
		emit("mov rax, 0")
		emit("mov rsp, rbp")
		emit("pop rbp")
		emit("ret")

		return
	}
	if node.Kind == parse.NodeNot {
		gen(node.Target)
		emit("pop rax")
		emit("xor rax, 1")
		emit("push rax")
		return
	}
	if node.Kind == parse.NodeAddr {
		genLvalue(node.Target)
		return
	}
	if node.Kind == parse.NodeDeref {
		gen(node.Target)
		emit("pop rax")
		emit("mov rax, [rax]")
		emit("push rax")
		return
	}
	if node.Kind == parse.NodeShortVarDeclStmt {
		var lhs = node.Children[0]
		var rhs = node.Children[1]

		for _, v := range lhs.Children {
			declare(v)
		}

		if rhs.ExprType.Kind == lang.TypeMultiple && len(rhs.Children) == 1 {
			assignMultiple(lhs.Children, rhs.Children[0])
			return
		}

		for i, l := range lhs.Children {
			r := rhs.Children[i]
			assign(l, r)
		}
		return
	}
	if node.Kind == parse.NodeLocalVarStmt {
		declare(node.Children[0])
		if len(node.Children) == 2 {
			assign(node.Children[0], node.Children[1])
		}
		return
	}
	if node.Kind == parse.NodeTopLevelVarStmt {
		declare(node.Children[0])
		return
	}
	if node.Kind == parse.NodeExprStmt {
		gen(node.Children[0])
		if node.Children[0].ExprType.Kind == lang.TypeMultiple {
			for range node.Children[0].ExprType.Components {
				emit("pop rax")
			}
			return
		}
		emit("pop rax")
		return
	}
	if node.Kind == parse.NodeIndex {
		genLvalue(node)
		emit("pop rax")
		if lang.Sizeof(node.ExprType) == 1 {
			emit("movzx rax, BYTE PTR [rax]")
		} else {
			emit("mov rax, [rax]")
		}
		emit("push rax")
		return
	}
	if node.Kind == parse.NodeDot {
		genLvalue(node)
		emit("pop rax")
		if lang.Sizeof(node.ExprType) == 1 {
			emit("movzx rax, BYTE PTR [rax]")
		} else {
			emit("mov rax, [rax]")
		}
		emit("push rax")
		return
	}
	if node.Kind == parse.NodeString {
		emit("mov rax, OFFSET FLAT:%s", node.Str.Label)
		emit("push rax")
		return
	}
	if node.Kind == parse.NodeLogicalAnd {
		gen(node.Lhs)
		emit("pop rax")
		emit("push 0")
		emit("cmp rax, 0")

		var label = ".Land" + strconv.Itoa(labelNumber)
		labelNumber++

		// 短絡評価する
		emit("je " + label)

		emit("pop rax") // スタックから0を削除する
		gen(node.Rhs)

		emit("pop rax")
		emit("cmp rax, 1")
		emit("sete al")
		emit("movzb rax, al")
		emit("push rax")

		p("%s:", label)

		return
	}
	if node.Kind == parse.NodeLogicalOr {
		gen(node.Lhs)
		emit("pop rax")
		emit("push 1")
		emit("cmp rax, 1")

		var label = ".Lor" + strconv.Itoa(labelNumber)
		labelNumber++

		// 短絡評価する
		emit("je " + label)

		emit("pop rax") // スタックから1を削除する
		gen(node.Rhs)

		emit("pop rax")
		emit("cmp rax, 1")
		emit("sete al")
		emit("movzb rax, al")
		emit("push rax")

		p("%s:", label)

		return
	}
	if node.Kind == parse.NodeSliceLiteral {
		var elemType = *node.LiteralType.PtrTo

		emit("mov rdi, %d", 8+lang.Sizeof(elemType)*len(node.Children))
		emit("call malloc")
		emit("mov QWORD PTR [rax], %d", len(node.Children)) // 要素数を表す領域
		emit("push rax")

		for i := 0; i < len(node.Children); i++ {
			gen(node.Children[i])
			emit("pop rdi")
			emit("pop rax")
			emit("mov %s PTR [rax+%d], rdi", word(lang.Sizeof(elemType)), 8+i*lang.Sizeof(elemType))
			emit("push rax")
		}
		return
	}
	if node.Kind == parse.NodeAppendCall {
		gen(node.Arguments[1])
		gen(node.Arguments[0])

		var elemType = node.Arguments[1].ExprType

		emit("pop rdi")
		emit("add QWORD PTR [rdi], 1") // 要素数を増やす
		emit("mov rsi, [rdi]")
		emit("imul rsi, %d", lang.Sizeof(elemType))
		emit("add rsi, 8") // 要素数分のアドレス

		emit("call realloc") // 8 + 要素数 x 要素サイズ分のメモリを確保

		emit("mov r10, rax") // 退避

		emit("mov rdi, [rax]") // rdiに要素数を代入
		emit("sub rdi, 1")
		emit("imul rdi, %d", lang.Sizeof(elemType))
		emit("add rdi, 8")   // 要素数用のオフセットを加算
		emit("add rax, rdi") // 代入するべき要素のアドレス

		emit("pop rdi") // 追加する要素の値
		emit("mov %s PTR [rax], rdi", word(lang.Sizeof(elemType)))
		emit("push r10")
		return
	}

	gen(node.Lhs)
	gen(node.Rhs)

	emit("pop rdi")
	emit("pop rax")

	switch node.Kind {
	case parse.NodeAdd:
		emit("add rax, rdi")
	case parse.NodeSub:
		emit("sub rax, rdi")
	case parse.NodeMul:
		emit("imul rax, rdi")
	case parse.NodeDiv:
		emit("cqo")
		emit("idiv rdi")
	case parse.NodeEql:
		emit("cmp rax, rdi")
		emit("sete al")
		emit("movzb rax, al")
	case parse.NodeNotEql:
		emit("cmp rax, rdi")
		emit("setne al")
		emit("movzb rax, al")
	case parse.NodeLess:
		emit("cmp rax, rdi")
		emit("setl al")
		emit("movzb rax, al")
	case parse.NodeLessEql:
		emit("cmp rax, rdi")
		emit("setle al")
		emit("movzb rax, al")
	case parse.NodeGreater:
		emit("cmp rdi, rax")
		emit("setl al")
		emit("movzb rax, al")
	case parse.NodeGreaterEql:
		emit("cmp rdi, rax")
		emit("setle al")
		emit("movzb rax, al")
	}
	emit("push rax")
}

func GenX86_64(prog *parse.Program) {
	program = prog
	// アセンブリの前半部分
	p(".intel_syntax noprefix")
	p(".globl main")

	p(".data")
	for _, str := range prog.StringLiterals {
		p(str.Label + ":")
		emit(".string %s", str.Value)
	}
	p(".text")

	for _, c := range prog.Code {
		// 抽象構文木を下りながらコード生成
		gen(c)
	}
}
