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
