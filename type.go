package main

type TypeKind string

const (
	TypeInt       TypeKind = "[TYPE] INT"
	TypePtr       TypeKind = "[TYPE] PTR"
	TypeUndefined TypeKind = "[TYPE] UNDEFINED" // まだ型を決めることができていない
)

type Type struct {
	kind  TypeKind
	ptrTo *Type
}

func sizeof(kind TypeKind) int {
	if kind == TypeInt || kind == TypePtr {
		return 8
	}
	// 変数の型推論をまだ実装していないので、一旦8を返すようにする
	// panic("未定義の型の変数です")
	return 8
}
