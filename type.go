package main

type TypeKind string

const (
	TypeInt       TypeKind = "[TYPE] INT"
	TypePtr       TypeKind = "[TYPE] PTR"
	TypeVoid      TypeKind = "[TYPE] VOID"
	TypeUndefined TypeKind = "[TYPE] UNDEFINED" // まだ型を決めることができていない
)

type Type struct {
	kind  TypeKind
	ptrTo *Type
}

func NewType(kind TypeKind) Type {
	return Type{kind: kind}
}

func sizeof(kind TypeKind) int {
	if kind == TypeInt || kind == TypePtr {
		return 8
	}
	// 変数の型推論をまだ実装していないので、一旦8を返すようにする
	// panic("未定義の型の変数です")
	return 8
}

func typeEquals(t1 Type, t2 Type) bool {
	if t1.kind != t2.kind {
		return false
	}
	if t1.kind == TypePtr && t2.kind == TypePtr {
		return typeEquals(*t1.ptrTo, *t2.ptrTo)
	}
	return true
}
