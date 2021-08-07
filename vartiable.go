package main

type VariableKind string

const (
	VariableLocal    VariableKind = "VARIABLE LOCAL"
	VariableTopLevel VariableKind = "VARIABLE TOP LEVEL"
)

type Variable struct {
	kind    VariableKind // トップレベル変数か、ローカル変数か
	name    string       // 変数の名前
	varType Type         // 変数の型
	offset  int          // RBPからのオフセット。
}
