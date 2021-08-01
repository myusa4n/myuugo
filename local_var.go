package main

type LocalVar struct {
	name   string // 変数の名前
	offset int    // RBPからのオフセット
}

var localVarTable map[string][]LocalVar

func addLocalVar(fnLabel string, token Token) LocalVar {
	lvar, ok := findLocalVar(fnLabel, token)
	if ok {
		return lvar
	}
	locals, ok := localVarTable[fnLabel]
	if !ok {
		locals = make([]LocalVar, 0)
		localVarTable[fnLabel] = locals
	}

	lvar = LocalVar{name: token.str}
	if len(locals) == 0 {
		lvar.offset = 0 + 8
	} else {
		lvar.offset = locals[len(locals)-1].offset + 8
	}
	localVarTable[fnLabel] = append(localVarTable[fnLabel], lvar)
	return lvar
}

func findLocalVar(fnLabel string, token Token) (LocalVar, bool) {
	locals, ok := localVarTable[fnLabel]

	if ok {
		for _, lvar := range locals {
			if lvar.name == token.str {
				return lvar, true
			}
		}
	}
	return LocalVar{}, false
}
