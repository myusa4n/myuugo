package main

type LocalVar struct {
	name    string // 変数の名前
	offset  int    // RBPからのオフセット
	varType Type   // 変数の型
}
