package main

type LocalVar struct {
	name   string // 変数の名前
	offset int    // RBPからのオフセット
}

var locals []LocalVar

func findLocalVar(token Token) (LocalVar, bool) {
	for _, lvar := range locals {
		if lvar.name == token.str {
			return lvar, true
		}
	}
	return LocalVar{}, false
}
