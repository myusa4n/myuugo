package main

import (
	"fmt"
	"strconv"
)

var labelNumber = 0

func genLvalue(node *Node) {
	if node.kind != NodeLocalVar {
		madden("代入の左辺値が変数ではありません")
	}
	fmt.Println("  mov rax, rbp")
	fmt.Printf("  sub rax, %d\n", node.offset)
	fmt.Println("  push rax")
}

func gen(node *Node) {
	if node.kind == NodeNum {
		fmt.Printf("  push %d\n", node.val)
		return
	}
	if node.kind == NodeStmtList {
		for i, stmt := range node.children {
			if i > 0 {
				fmt.Println("  pop rax")
			}
			gen(stmt)
		}
		return
	}
	if node.kind == NodeReturn {
		gen(node.children[0])
		fmt.Println("  pop rax")
		fmt.Println("  mov rsp, rbp")
		fmt.Println("  pop rbp")
		fmt.Println("  ret")
		return
	}
	if node.kind == NodeLocalVar {
		genLvalue(node)
		fmt.Println("  pop rax")
		fmt.Println("  mov rax, [rax]")
		fmt.Println("  push rax")
		return
	}
	if node.kind == NodeAssign {
		genLvalue(node.children[0]) // lhs
		gen(node.children[1])       // rhs

		fmt.Println("  pop rdi")
		fmt.Println("  pop rax")
		fmt.Println("  mov [rax], rdi")
		fmt.Println("  push rdi")
		return
	}
	if node.kind == NodeMetaIf {
		var endLabel = ".Lend" + strconv.Itoa(labelNumber)
		var elseLabel = ".Lelse" + strconv.Itoa(labelNumber)

		gen(node.children[0]) // if
		fmt.Println(elseLabel + ":")
		if node.children[1] != nil {
			gen(node.children[1]) // else
		}
		fmt.Println(endLabel + ":")
		labelNumber += 1
		return
	}
	if node.kind == NodeIf {
		var endLabel = ".Lend" + strconv.Itoa(labelNumber)
		var elseLabel = ".Lelse" + strconv.Itoa(labelNumber)

		gen(node.children[0]) // lhs
		fmt.Println("  pop rax")
		fmt.Println("  cmp rax, 0")
		fmt.Println("  je " + elseLabel)
		gen(node.children[1]) // rhs
		fmt.Println("  jmp " + endLabel)
		return
	}
	if node.kind == NodeElse {
		gen(node.children[0])
		return
	}
	if node.kind == NodeFor {
		var beginLabel = ".Lbegin" + strconv.Itoa(labelNumber)
		var endLabel = ".Lend" + strconv.Itoa(labelNumber)
		labelNumber += 1

		if node.children[0] != nil {
			gen(node.children[0])
		}
		fmt.Println(beginLabel + ":")
		if node.children[1] != nil {
			gen(node.children[1]) // 条件
			fmt.Println("  pop rax")
			fmt.Println("  cmp rax, 0")
			fmt.Println("  je " + endLabel)
		}
		gen(node.children[3])
		if node.children[2] != nil {
			gen(node.children[2])
		}
		fmt.Println("  jmp " + beginLabel)
		fmt.Println(endLabel + ":")
		return
	}

	gen(node.children[0]) // lhs
	gen(node.children[1]) // rhs

	fmt.Println("  pop rdi")
	fmt.Println("  pop rax")

	switch node.kind {
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
