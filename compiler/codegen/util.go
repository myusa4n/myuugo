package codegen

import (
	"fmt"
	"strconv"

	"github.com/myuu222/myuugo/compiler/lang"
	"github.com/myuu222/myuugo/compiler/parse"
)

func getFrameSize(program *parse.Program, functionName string) int {
	fn := program.FindFunction(functionName)
	if fn == nil {
		panic("関数 \"" + functionName + " は存在しません")
	}
	var size int = len(fn.LocalVariables) * 8
	size = ((size + 16 - 1) / 16) * 16
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
	if ty.Kind == lang.TypeUserDefined {
		return entitySizeOf(*ty.PtrTo)
	}
	if ty.Kind == lang.TypeArray {
		return ty.ArraySize * lang.Sizeof(*ty.PtrTo)
	}
	if ty.Kind == lang.TypeStruct {
		sum := 0
		for _, ty := range ty.MemberTypes {
			sum += lang.Sizeof(ty)
		}
		return sum
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
