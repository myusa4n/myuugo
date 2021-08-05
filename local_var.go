package main

type LocalVar struct {
	name    string // 変数の名前
	offset  int    // RBPからのオフセット
	varType Type   // 変数の型
}

var localVarTable map[string][]*LocalVar

func registerFunc(fnLabel string) {
	_, ok := localVarTable[fnLabel]
	if ok {
		madden("関数%sは既に存在しています", fnLabel)
	}
	localVarTable[fnLabel] = []*LocalVar{}
}

func addLocalVar(fnLabel string, token Token) *LocalVar {
	lvar := findLocalVar(fnLabel, token)
	if lvar != nil {
		return lvar
	}
	var locals = localVarTable[fnLabel]
	lvar = &LocalVar{name: token.str, varType: Type{kind: TypeUndefined}}
	if len(locals) == 0 {
		lvar.offset = 0 + 8
	} else {
		lvar.offset = locals[len(locals)-1].offset + 8
	}
	localVarTable[fnLabel] = append(localVarTable[fnLabel], lvar)
	return lvar
}

func findLocalVar(fnLabel string, token Token) *LocalVar {
	locals, ok := localVarTable[fnLabel]

	if !ok {
		madden("関数%sは存在しません", fnLabel)
	}
	for _, lvar := range locals {
		if lvar.name == token.str {
			return lvar
		}
	}
	return nil
}

func getFrameSize(fnLabel string) int {
	locals, ok := localVarTable[fnLabel]
	if !ok {
		madden("関数%sは存在しません", fnLabel)
	}
	var size int = 0
	for _, lvar := range locals {
		size += sizeof(lvar.varType.kind)
	}
	return size
}
