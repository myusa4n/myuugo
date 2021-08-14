package main

import (
	"fmt"
	"os"

	. "github.com/myuu222/myuugo/codegen"
	. "github.com/myuu222/myuugo/parse"
	. "github.com/myuu222/myuugo/passes"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "引数の個数が正しくありません")
		os.Exit(1)
	}

	var path = os.Args[1]
	var tokenizer = NewTokenizer()
	tokenizer.Tokenize(path)

	var program = Parse(tokenizer)

	Semantic(program.Code)

	GenX86_64(program)
}
