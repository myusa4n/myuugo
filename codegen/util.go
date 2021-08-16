package codegen

import (
	"fmt"
	"strconv"

	"github.com/myuu222/myuugo/lang"
	"github.com/myuu222/myuugo/parse"
)

func getFrameSize(program *parse.Program, functionName string) int {
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

func entitySizeOf(ty lang.Type) int {
	if ty.Kind == lang.TypeArray {
		return ty.ArraySize * entitySizeOf(*ty.PtrTo)
	}
	return lang.Sizeof(ty)
}

func emit(format string, args ...interface{}) {
	fmt.Printf("  ")
	fmt.Printf(format, args...)
	fmt.Printf("\n")
}

func p(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	fmt.Printf("\n")
}
