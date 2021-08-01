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
	localVarTable = make(map[string][]LocalVar)
	tokens = tokenize(userInput)
	program()

	// アセンブリの前半部分
	fmt.Println(".intel_syntax noprefix")
	fmt.Println(".globl main")

	for _, c := range code {
		// 抽象構文木を下りながらコード生成
		gen(c)
	}
}
