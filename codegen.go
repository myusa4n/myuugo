package main

import "fmt"

func gen(node *Node) {
	if node.kind == NodeNum {
		fmt.Printf("  push %d\n", node.val)
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
