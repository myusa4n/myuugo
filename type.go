package main

type TypeKind string

const (
	TypeInt       TypeKind = "[TYPE] INT"
	TypeRune      TypeKind = "[TYPE] RUNE"
	TypePtr       TypeKind = "[TYPE] PTR"
	TypeVoid      TypeKind = "[TYPE] VOID"
	TypeArray     TypeKind = "[TYPE] ARRAY"
	TypeStmt      TypeKind = "[TYPE] STMT"      // 簡便のため存在させている
	TypeUndefined TypeKind = "[TYPE] UNDEFINED" // まだ型を決めることができていない
)

type Type struct {
	kind      TypeKind
	ptrTo     *Type
	arraySize int
}

func NewType(kind TypeKind) Type {
	return Type{kind: kind}
}

func NewArrayType(elemType Type, size int) Type {
	return Type{kind: TypeArray, ptrTo: &elemType, arraySize: size}
}

func Sizeof(ty Type) int {
	if ty.kind == TypeInt || ty.kind == TypePtr {
		return 8
	}
	if ty.kind == TypeRune {
		return 1
	}
	if ty.kind == TypeArray {
		return ty.arraySize * Sizeof(*ty.ptrTo)
	}
	// 未定義
	return 0
}

func typeEquals(t1 Type, t2 Type) bool {
	if t1.kind != t2.kind {
		return false
	}
	if t1.kind == TypePtr {
		return typeEquals(*t1.ptrTo, *t2.ptrTo)
	}
	if t1.kind == TypeArray {
		return t1.arraySize == t2.arraySize && typeEquals(*t1.ptrTo, *t2.ptrTo)
	}
	return true
}

func IsKindOfNumber(t Type) bool {
	return t.kind == TypeInt || t.kind == TypeRune
}

func TypeCompatable(t1 Type, t2 Type) bool {
	if t1.kind == TypePtr && t2.kind == TypePtr {
		return typeEquals(*t1.ptrTo, *t2.ptrTo)
	}
	if t1.kind == TypeArray && t2.kind == TypeArray {
		return typeEquals(*t1.ptrTo, *t2.ptrTo)
	}
	return IsKindOfNumber(t1) && IsKindOfNumber(t2)
}
