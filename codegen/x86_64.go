package codegen

import (
	"fmt"
	"strconv"

	. "github.com/myuu222/myuugo/lang"
	. "github.com/myuu222/myuugo/parse"
	. "github.com/myuu222/myuugo/util"
)

var labelNumber = 0

func genLvalue(node *Node) {
	if node.Kind == NodeDeref {
		gen(node.Children[0])
		return
	} else if node.Kind == NodeVariable {
		if node.Variable.Kind == VariableLocal {
			fmt.Println("  mov rax, rbp")
			fmt.Printf("  sub rax, %d\n", node.Variable.Offset)
			fmt.Println("  push rax")
		} else {
			fmt.Printf("  mov rax, OFFSET FLAT:%s\n", node.Variable.Name)
			fmt.Println("  push rax")
		}
		return
	} else if node.Kind == NodeIndex {
		genLvalue(node.Children[0])
		gen(node.Children[1])
		fmt.Println("  pop rdi")
		fmt.Printf("  imul rdi, %d\n", Sizeof(*node.Children[0].Variable.Type.PtrTo))
		fmt.Println("  pop rax")
		fmt.Println("  add rax, rdi")
		fmt.Println("  push rax")
		return
	}
	Alarm("代入の左辺値が変数またはポインタ参照ではありません")
}

func gen(node *Node) {
	if node.Kind == NodePackageStmt {
		// 何もしない
		return
	}
	if node.Kind == NodeNum {
		fmt.Printf("  push %d\n", node.Val)
		return
	}
	if node.Kind == NodeStmtList {
		for _, stmt := range node.Children {
			gen(stmt)
		}
		return
	}
	if node.Kind == NodeReturn {
		if len(node.Children) > 0 {
			var exprs = node.Children[0].Children
			for _, e := range exprs {
				gen(e)
			}

			var registers = [7]string{"rax", "rdi", "rsi", "rdx", "rcx", "r8", "r9"}
			for i := range exprs {
				fmt.Println("  pop " + registers[len(exprs)-i-1])
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
	if node.Kind == NodeVariable {
		genLvalue(node)
		fmt.Println("  pop rax")
		if Sizeof(node.ExprType) == 1 {
			fmt.Println("  movzx rax, BYTE PTR [rax]")
		} else { // 8
			fmt.Println("  mov rax, [rax]")
		}
		fmt.Println("  push rax")
		return
	}
	if node.Kind == NodeAssign {
		// TODO: 左辺が配列だった場合は丸々コピーさせる必要がある
		var lhs = node.Children[0]
		var rhs = node.Children[1]

		if rhs.ExprType.Kind == TypeMultiple && len(rhs.Children) == 1 {
			gen(rhs.Children[0])

			// 分解する
			// 右端の変数から代入されることになる
			for i := len(lhs.Children) - 1; i >= 0; i-- {
				var l = lhs.Children[i]
				genLvalue(l)
				if Sizeof(l.ExprType) == 1 {
					fmt.Println("  pop rax")
					fmt.Println("  pop rdi")
					fmt.Println("  mov [rax], dil")
				} else { // 8
					fmt.Println("  pop rax")
					fmt.Println("  pop rdi")
					fmt.Println("  mov [rax], rdi")
				}
			}
			return
		}

		for i, l := range lhs.Children {
			r := rhs.Children[i]

			genLvalue(l)
			gen(r)

			if Sizeof(l.ExprType) == 1 {
				fmt.Println("  pop rdi")
				fmt.Println("  pop rax")
				fmt.Println("  mov [rax], dil")
			} else { // 8
				fmt.Println("  pop rdi")
				fmt.Println("  pop rax")
				fmt.Println("  mov [rax], rdi")
			}
		}
		return
	}
	if node.Kind == NodeMetaIf {
		var endLabel = ".Lend" + strconv.Itoa(labelNumber)
		var elseLabel = ".Lelse" + strconv.Itoa(labelNumber)

		gen(node.Children[0]) // if
		fmt.Println(elseLabel + ":")
		if node.Children[1] != nil {
			gen(node.Children[1]) // else
		}
		fmt.Println(endLabel + ":")
		labelNumber += 1
		return
	}
	if node.Kind == NodeIf {
		var endLabel = ".Lend" + strconv.Itoa(labelNumber)
		var elseLabel = ".Lelse" + strconv.Itoa(labelNumber)

		gen(node.Children[0]) // lhs
		fmt.Println("  pop rax")
		fmt.Println("  cmp rax, 0")
		fmt.Println("  je " + elseLabel)
		gen(node.Children[1]) // rhs
		fmt.Println("  jmp " + endLabel)
		return
	}
	if node.Kind == NodeElse {
		gen(node.Children[0])
		return
	}
	if node.Kind == NodeFor {
		var beginLabel = ".Lbegin" + strconv.Itoa(labelNumber)
		var endLabel = ".Lend" + strconv.Itoa(labelNumber)
		labelNumber += 1

		// children := (初期化, 条件, 更新)

		if node.Children[0] != nil {
			gen(node.Children[0])
		}
		fmt.Println(beginLabel + ":")
		if node.Children[1] != nil {
			gen(node.Children[1]) // 条件
			fmt.Println("  pop rax")
			fmt.Println("  cmp rax, 0")
			fmt.Println("  je " + endLabel)
		}
		gen(node.Children[3])
		if node.Children[2] != nil {
			gen(node.Children[2])
		}
		fmt.Println("  jmp " + beginLabel)
		fmt.Println(endLabel + ":")
		return
	}
	if node.Kind == NodeFunctionCall {
		// TODO: rune型と配列型の扱いについて考える
		var registers = [7]string{"rax", "rdi", "rsi", "rdx", "rcx", "r8", "r9"}
		for _, argument := range node.Children {
			gen(argument)
		}
		for i := range node.Children {
			fmt.Println("  pop " + registers[len(node.Children)-i])
		}
		fmt.Println("  mov al, 0") // 可変長引数の関数を呼び出すためのルール
		fmt.Println("  call " + node.Label)

		// 今見ている関数が多値だった場合は、rax, rdi, rsi, ...から取り出していく
		fn := node.Env.FindFunction(node.Label)
		if fn != nil && fn.ReturnValueType.Kind == TypeMultiple {
			// raxから順にスタックに突っ込んでいく
			for i := range fn.ReturnValueType.Components {
				fmt.Println("  push " + registers[i])
			}
			return
		}
		fmt.Println("  push rax")
		return
	}
	if node.Kind == NodeFunctionDef {
		fmt.Println(node.Label + ":")

		// プロローグ
		// 変数26個分の領域を確保する
		fmt.Println("  push rbp")
		fmt.Println("  mov rbp, rsp")

		fmt.Printf("  sub rsp, %d\n", node.Env.GetFrameSize(node.Label))

		var registers [6]string = [6]string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}

		for i, param := range node.Children[1:] { // 引数
			genLvalue(param)
			fmt.Println("  pop rax")
			fmt.Println("  mov [rax], " + registers[i])
		}

		gen(node.Children[0]) // 関数本体

		// エピローグ
		// 関数の返り値の型が void 型だと仮定する
		fmt.Println("  mov rax, 0")
		fmt.Println("  mov rsp, rbp")
		fmt.Println("  pop rbp")
		fmt.Println("  ret")

		return
	}
	if node.Kind == NodeAddr {
		genLvalue(node.Children[0])
		return
	}
	if node.Kind == NodeDeref {
		gen(node.Children[0])
		fmt.Println("  pop rax")
		fmt.Println("  mov rax, [rax]")
		fmt.Println("  push rax")
		return
	}
	if node.Kind == NodeShortVarDeclStmt {
		var lhs = node.Children[0]
		var rhs = node.Children[1]

		if rhs.ExprType.Kind == TypeMultiple && len(rhs.Children) == 1 {
			gen(rhs.Children[0])

			// 分解する
			// 右端の変数から代入されることになる
			for i := len(lhs.Children) - 1; i >= 0; i-- {
				var l = lhs.Children[i]
				genLvalue(l)
				if Sizeof(l.ExprType) == 1 {
					fmt.Println("  pop rax")
					fmt.Println("  pop rdi")
					fmt.Println("  mov [rax], dil")
				} else { // 8
					fmt.Println("  pop rax")
					fmt.Println("  pop rdi")
					fmt.Println("  mov [rax], rdi")
				}
			}
			return
		}

		for i, l := range lhs.Children {
			r := rhs.Children[i]

			genLvalue(l)
			gen(r)

			fmt.Println("  pop rdi")
			fmt.Println("  pop rax")

			if Sizeof(l.ExprType) == 1 {
				fmt.Println("  mov [rax], dil")
			} else { // 8
				fmt.Println("  mov [rax], rdi")
			}
		}
		return
	}
	if node.Kind == NodeLocalVarStmt {
		if len(node.Children) == 2 {
			genLvalue(node.Children[0]) // lhs
			gen(node.Children[1])       // rhs

			fmt.Println("  pop rdi")
			fmt.Println("  pop rax")

			if Sizeof(node.Children[0].ExprType) == 1 {
				fmt.Println("  mov [rax], dil")
			} else { // 8
				fmt.Println("  mov [rax], rdi")
			}
			return
		}
		return
	}
	if node.Kind == NodeTopLevelVarStmt {
		fmt.Println(".data")
		var tvar = node.Children[0]
		fmt.Println(tvar.Variable.Name + ":")
		fmt.Printf("  .zero %d\n", Sizeof(tvar.Variable.Type))
		fmt.Println(".text")
		return
	}
	if node.Kind == NodeExprStmt {
		gen(node.Children[0])
		if node.Children[0].ExprType.Kind == TypeMultiple {
			for range node.Children[0].ExprType.Components {
				fmt.Println("  pop rax")
			}
			return
		}
		fmt.Println("  pop rax")
		return
	}
	if node.Kind == NodeIndex {
		genLvalue(node)
		fmt.Println("  pop rax")
		if Sizeof(node.ExprType) == 1 {
			fmt.Println("  movzx rax, BYTE PTR [rax]")
		} else {
			fmt.Println("  mov rax, [rax]")
		}
		fmt.Println("  push rax")
		return
	}
	if node.Kind == NodeString {
		fmt.Printf("  mov rax, OFFSET FLAT:%s\n", node.Str.Label)
		fmt.Println("  push rax")
		return
	}

	gen(node.Children[0]) // lhs
	gen(node.Children[1]) // rhs

	fmt.Println("  pop rdi")
	fmt.Println("  pop rax")

	switch node.Kind {
	case NodeAdd:
		fmt.Println("  add rax, rdi")
	case NodeSub:
		fmt.Println("  sub rax, rdi")
	case NodeMul:
		fmt.Println("  imul rax, rdi")
	case NodeDiv:
		fmt.Println("  cqo")
		fmt.Println("  idiv rdi")
	case NodeEql:
		fmt.Println("  cmp rax, rdi")
		fmt.Println("  sete al")
		fmt.Println("  movzb rax, al")
	case NodeNotEql:
		fmt.Println("  cmp rax, rdi")
		fmt.Println("  setne al")
		fmt.Println("  movzb rax, al")
	case NodeLess:
		fmt.Println("  cmp rax, rdi")
		fmt.Println("  setl al")
		fmt.Println("  movzb rax, al")
	case NodeLessEql:
		fmt.Println("  cmp rax, rdi")
		fmt.Println("  setle al")
		fmt.Println("  movzb rax, al")
	case NodeGreater:
		fmt.Println("  cmp rdi, rax")
		fmt.Println("  setl al")
		fmt.Println("  movzb rax, al")
	case NodeGreaterEql:
		fmt.Println("  cmp rdi, rax")
		fmt.Println("  setle al")
		fmt.Println("  movzb rax, al")
	}
	fmt.Println("  push rax")
}

func GenX86_64(program *Program) {
	// アセンブリの前半部分
	fmt.Println(".intel_syntax noprefix")
	fmt.Println(".globl main")

	fmt.Println(".data")
	for _, str := range program.StringLiterals {
		fmt.Println(str.Label + ":")
		fmt.Println("  .string " + str.Value)
	}
	fmt.Println(".text")

	for _, c := range program.Code {
		// 抽象構文木を下りながらコード生成
		gen(c)
	}
}
