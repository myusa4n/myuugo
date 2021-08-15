package codegen

import (
	"fmt"
	"strconv"

	"github.com/myuu222/myuugo/lang"
	"github.com/myuu222/myuugo/parse"
	"github.com/myuu222/myuugo/util"
)

var labelNumber = 0
var program *parse.Program

func getFrameSize(functionName string) int {
	fn := program.FindFunction(functionName)
	if fn == nil {
		panic("関数 \"" + functionName + " は存在しません")
	}
	var size int = 0
	for _, lvar := range fn.LocalVariables {
		size += lang.Sizeof(lvar.Type)
	}
	return size
}

func register(nth int, byteCount int) string {
	var regs64 = []string{"rax", "rdi", "rsi", "rdx", "rcx", "r8", "r9"}
	var regs8 = []string{"al", "dil", "sil", "dl", "cl", "r8b", "r9b"}

	if byteCount == 8 {
		return regs64[nth]
	} else if byteCount == 1 {
		return regs8[nth]
	} else {
		panic(strconv.Itoa(byteCount) + "Bのレジスタは存在しません")
	}
}

func word(byteCount int) string {
	if byteCount == 8 {
		return "QWORD"
	} else if byteCount == 1 {
		return "BYTE"
	} else {
		panic("違法なバイト数の指定です")
	}
}

func assign(lhs *parse.Node, rhs *parse.Node) {
	genLvalue(lhs)
	gen(rhs)

	fmt.Println("  pop rdi")
	fmt.Println("  pop rax")

	if lhs.ExprType.Kind == lang.TypeArray {
		var size = lang.Sizeof(*lhs.ExprType.PtrTo)

		for i := 0; i < lhs.ExprType.ArraySize; i++ {
			fmt.Printf("  mov r10, %s PTR [rdi+%d]\n", word(size), size*i)
			fmt.Printf("  mov %s PTR [rax+%d], r10\n", word(size), size*i)
		}
		return
	}
	fmt.Println("  mov [rax], " + register(1, lang.Sizeof(lhs.ExprType)))
}

// 多値を返す関数の返り値を左辺にある複数の変数に代入する
func assignMultiple(lhss []*parse.Node, rhs *parse.Node) {
	gen(rhs)

	// 分解する
	// 右端の変数から代入されることになる
	for i := len(lhss) - 1; i >= 0; i-- {
		var l = lhss[i]
		genLvalue(l)

		fmt.Println("  pop rax")
		fmt.Println("  pop rdi")
		fmt.Println("  mov [rax], " + register(1, lang.Sizeof(l.ExprType)))
	}
}

func genLvalue(node *parse.Node) {
	if node.Kind == parse.NodeDeref {
		gen(node.Target)
		return
	} else if node.Kind == parse.NodeVariable {
		if node.Variable.Kind == lang.VariableLocal {
			fmt.Println("  mov rax, rbp")
			fmt.Printf("  sub rax, %d\n", node.Variable.Offset)
			fmt.Println("  push rax")
		} else {
			fmt.Printf("  mov rax, OFFSET FLAT:%s\n", node.Variable.Name)
			fmt.Println("  push rax")
		}
		return
	} else if node.Kind == parse.NodeIndex {
		genLvalue(node.Seq)
		gen(node.Index)
		fmt.Println("  pop rdi")
		fmt.Printf("  imul rdi, %d\n", lang.Sizeof(node.ExprType))
		fmt.Println("  pop rax")
		fmt.Println("  add rax, rdi")
		fmt.Println("  push rax")
		return
	}
	util.Alarm("代入の左辺値が変数またはポインタ参照ではありません")
}

func gen(node *parse.Node) {
	if node.Kind == parse.NodePackageStmt {
		// 何もしない
		return
	}
	if node.Kind == parse.NodeNum {
		fmt.Printf("  push %d\n", node.Val)
		return
	}
	if node.Kind == parse.NodeBool {
		fmt.Printf("  push %d\n", node.Val)
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
				fmt.Println("  pop " + register(len(exprs)-i-1, 8))
			}
		} else {
			// void型
			fmt.Println("  mov rax, 0")
		}
		fmt.Println("  mov rsp, rbp")
		fmt.Println("  pop rbp")
		fmt.Println("  ret")
		return
	}
	if node.Kind == parse.NodeVariable {
		genLvalue(node)
		fmt.Println("  pop rax")

		if node.ExprType.Kind == lang.TypeArray {
			fmt.Println("  push rax")
			return
		}

		if lang.Sizeof(node.ExprType) == 1 {
			fmt.Println("  movzx rax, BYTE PTR [rax]")
		} else { // 8
			fmt.Println("  mov rax, [rax]")
		}
		fmt.Println("  push rax")
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
		fmt.Println(elseLabel + ":")

		if node.Else != nil {
			gen(node.Else)
		}
		fmt.Println(endLabel + ":")
		return
	}
	if node.Kind == parse.NodeIf {
		var endLabel = ".Lend" + strconv.Itoa(labelNumber)
		var elseLabel = ".Lelse" + strconv.Itoa(labelNumber)
		labelNumber += 1

		gen(node.Condition)
		fmt.Println("  pop rax")
		fmt.Println("  cmp rax, 0")
		fmt.Println("  je " + elseLabel)
		gen(node.Body)
		fmt.Println("  jmp " + endLabel)
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
		fmt.Println(beginLabel + ":")
		if node.Condition != nil {
			gen(node.Condition) // 条件
			fmt.Println("  pop rax")
			fmt.Println("  cmp rax, 0")
			fmt.Println("  je " + endLabel)
		}
		gen(node.Body)
		if node.Update != nil {
			gen(node.Update)
		}
		fmt.Println("  jmp " + beginLabel)
		fmt.Println(endLabel + ":")
		return
	}
	if node.Kind == parse.NodeFunctionCall {
		// TODO: rune型と配列型の扱いについて考える
		for _, argument := range node.Arguments {
			gen(argument)
		}
		for i := range node.Arguments {
			// 配列や構造体は先頭のアドレスだけ渡しておいてNodeFunctionDef側でうまいこと代入してもらう
			fmt.Println("  pop " + register(len(node.Arguments)-i, 8))
		}
		fmt.Println("  mov al, 0") // 可変長引数の関数を呼び出すためのルール
		fmt.Println("  call " + node.Label)

		// 今見ている関数が多値だった場合は、rax, rdi, rsi, ...から取り出していく
		fn := program.FindFunction(node.Label)
		if fn != nil && fn.ReturnValueType.Kind == lang.TypeMultiple {
			// raxから順にスタックに突っ込んでいく
			for i := range fn.ReturnValueType.Components {
				fmt.Println("  push " + register(i, 8))
			}
			return
		}
		fmt.Println("  push rax")
		return
	}
	if node.Kind == parse.NodeFunctionDef {
		fmt.Println(node.Label + ":")

		// プロローグ
		fmt.Println("  push rbp")
		fmt.Println("  mov rbp, rsp")

		fmt.Printf("  sub rsp, %d\n", getFrameSize(node.Label))

		for i, param := range node.Parameters { // 引数
			genLvalue(param)
			fmt.Println("  pop rax")

			fmt.Println("  mov [rax], " + register(i+1, lang.Sizeof(param.ExprType)))
		}

		gen(node.Body) // 関数本体

		// エピローグ
		// 関数の返り値の型が void 型だと仮定する
		fmt.Println("  mov rax, 0")
		fmt.Println("  mov rsp, rbp")
		fmt.Println("  pop rbp")
		fmt.Println("  ret")

		return
	}
	if node.Kind == parse.NodeNot {
		gen(node.Target)
		fmt.Println("  pop rax")
		fmt.Println("  xor rax, 1")
		fmt.Println("  push rax")
		return
	}
	if node.Kind == parse.NodeAddr {
		genLvalue(node.Target)
		return
	}
	if node.Kind == parse.NodeDeref {
		gen(node.Target)
		fmt.Println("  pop rax")
		fmt.Println("  mov rax, [rax]")
		fmt.Println("  push rax")
		return
	}
	if node.Kind == parse.NodeShortVarDeclStmt {
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
	if node.Kind == parse.NodeLocalVarStmt {
		if len(node.Children) == 2 {
			assign(node.Children[0], node.Children[1])
		}
		return
	}
	if node.Kind == parse.NodeTopLevelVarStmt {
		fmt.Println(".data")
		var tvar = node.Children[0]
		fmt.Println(tvar.Variable.Name + ":")
		fmt.Printf("  .zero %d\n", lang.Sizeof(tvar.Variable.Type))
		fmt.Println(".text")
		return
	}
	if node.Kind == parse.NodeExprStmt {
		gen(node.Children[0])
		if node.Children[0].ExprType.Kind == lang.TypeMultiple {
			for range node.Children[0].ExprType.Components {
				fmt.Println("  pop rax")
			}
			return
		}
		fmt.Println("  pop rax")
		return
	}
	if node.Kind == parse.NodeIndex {
		genLvalue(node)
		fmt.Println("  pop rax")
		if lang.Sizeof(node.ExprType) == 1 {
			fmt.Println("  movzx rax, BYTE PTR [rax]")
		} else {
			fmt.Println("  mov rax, [rax]")
		}
		fmt.Println("  push rax")
		return
	}
	if node.Kind == parse.NodeString {
		fmt.Printf("  mov rax, OFFSET FLAT:%s\n", node.Str.Label)
		fmt.Println("  push rax")
		return
	}
	if node.Kind == parse.NodeLogicalAnd {
		gen(node.Lhs)
		fmt.Println("  pop rax")
		fmt.Println("  push 0")
		fmt.Println("  cmp rax, 0")

		var label = ".Land" + strconv.Itoa(labelNumber)
		labelNumber++

		// 短絡評価する
		fmt.Println("  je " + label)

		fmt.Println("  pop rax") // スタックから0を削除する
		gen(node.Rhs)

		fmt.Println("  pop rax")
		fmt.Println("  cmp rax, 1")
		fmt.Println("  sete al")
		fmt.Println("  movzb rax, al")
		fmt.Println("  push rax")

		fmt.Println(label + ":")

		return
	}
	if node.Kind == parse.NodeLogicalOr {
		gen(node.Lhs)
		fmt.Println("  pop rax")
		fmt.Println("  push 1")
		fmt.Println("  cmp rax, 1")

		var label = ".Lor" + strconv.Itoa(labelNumber)
		labelNumber++

		// 短絡評価する
		fmt.Println("  je " + label)

		fmt.Println("  pop rax") // スタックから1を削除する
		gen(node.Rhs)

		fmt.Println("  pop rax")
		fmt.Println("  cmp rax, 1")
		fmt.Println("  sete al")
		fmt.Println("  movzb rax, al")
		fmt.Println("  push rax")

		fmt.Println(label + ":")

		return
	}

	gen(node.Lhs)
	gen(node.Rhs)

	fmt.Println("  pop rdi")
	fmt.Println("  pop rax")

	switch node.Kind {
	case parse.NodeAdd:
		fmt.Println("  add rax, rdi")
	case parse.NodeSub:
		fmt.Println("  sub rax, rdi")
	case parse.NodeMul:
		fmt.Println("  imul rax, rdi")
	case parse.NodeDiv:
		fmt.Println("  cqo")
		fmt.Println("  idiv rdi")
	case parse.NodeEql:
		fmt.Println("  cmp rax, rdi")
		fmt.Println("  sete al")
		fmt.Println("  movzb rax, al")
	case parse.NodeNotEql:
		fmt.Println("  cmp rax, rdi")
		fmt.Println("  setne al")
		fmt.Println("  movzb rax, al")
	case parse.NodeLess:
		fmt.Println("  cmp rax, rdi")
		fmt.Println("  setl al")
		fmt.Println("  movzb rax, al")
	case parse.NodeLessEql:
		fmt.Println("  cmp rax, rdi")
		fmt.Println("  setle al")
		fmt.Println("  movzb rax, al")
	case parse.NodeGreater:
		fmt.Println("  cmp rdi, rax")
		fmt.Println("  setl al")
		fmt.Println("  movzb rax, al")
	case parse.NodeGreaterEql:
		fmt.Println("  cmp rdi, rax")
		fmt.Println("  setle al")
		fmt.Println("  movzb rax, al")
	}
	fmt.Println("  push rax")
}

func GenX86_64(p *parse.Program) {
	program = p
	// アセンブリの前半部分
	fmt.Println(".intel_syntax noprefix")
	fmt.Println(".globl main")

	fmt.Println(".data")
	for _, str := range p.StringLiterals {
		fmt.Println(str.Label + ":")
		fmt.Println("  .string " + str.Value)
	}
	fmt.Println(".text")

	for _, c := range p.Code {
		// 抽象構文木を下りながらコード生成
		gen(c)
	}
}
