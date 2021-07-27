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
	program()

	// アセンブリの前半部分
	fmt.Println(".intel_syntax noprefix")
	fmt.Println(".globl main")
	fmt.Println("main:")

	// プロローグ
	// 変数26個分の領域を確保する
	fmt.Println("  push rbp")
	fmt.Println("  mov rbp, rsp")
	fmt.Println("  sub rsp, 208")

	for i := 0; code[i] != nil; i++ {
		// 抽象構文木を下りながらコード生成
		gen(code[i])

		// 式の評価結果としてスタックに一つの値が残っているはずなので、スタックが溢れないようにポップしておく
		fmt.Println("  pop rax")
	}

	// エピローグ
	// 最後の式の結果がRAXに残っているのでそれが返り値になる
	fmt.Println("  mov rsp, rbp")
	fmt.Println("  pop rbp")
	fmt.Println("  ret")
}
