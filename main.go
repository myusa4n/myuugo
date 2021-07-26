package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "引数の個数が正しくありません")
		os.Exit(1)
	}

	userInput = os.Args[1]
	tokens = tokenize(userInput)
	var node = expr()

	// アセンブリの前半部分
	fmt.Println(".intel_syntax noprefix")
	fmt.Println(".globl main")
	fmt.Println("main:")

	// 抽象構文木を下りながらコード生成
	gen(node)

	// スタックトップに式全体の値が残っているはずなので
	// それをRAXにロードして関数からの返り値とする
	fmt.Println("  pop rax")
	fmt.Println("  ret")
}
