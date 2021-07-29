package main

import "fmt"

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
	if node.kind == NodeReturn {
		gen(node.lhs)
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
		genLvalue(node.lhs)
		gen(node.rhs)

		fmt.Println("  pop rdi")
		fmt.Println("  pop rax")
		fmt.Println("  mov [rax], rdi")
		fmt.Println("  push rdi")
		return
	}
	gen(node.lhs)
	gen(node.rhs)

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
