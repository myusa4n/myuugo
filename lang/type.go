package lang

type TypeKind string

const (
	TypeInt         TypeKind = "[TYPE] INT"
	TypeRune        TypeKind = "[TYPE] RUNE"
	TypePtr         TypeKind = "[TYPE] PTR"
	TypeVoid        TypeKind = "[TYPE] VOID"
	TypeArray       TypeKind = "[TYPE] ARRAY"
	TypeSlice       TypeKind = "[TYPE] SLICE"
	TypeBool        TypeKind = "[TYPE] BOOL"
	TypeStmt        TypeKind = "[TYPE] STMT"         // 簡便のため存在させている
	TypeMultiple    TypeKind = "[TYPE] MULTIPLE"     // 関数の返り値が複数だった場合に使う
	TypeUndefined   TypeKind = "[TYPE] UNDEFINED"    // まだ型を決めることができていない
	TypeUserDefined TypeKind = "[TYPE] USER DEFINED" // typeによりユーザが定義した型
	TypeStruct      TypeKind = "[TYPE] USER STRUCT"
)

type Type struct {
	Kind        TypeKind
	PtrTo       *Type
	ArraySize   int
	Components  []Type
	DefinedName string

	MemberNames   []string
	MemberTypes   []Type
	MemberOffsets []int
}

func NewType(kind TypeKind) Type {
	return Type{Kind: kind}
}

func NewMultipleType(components []Type) Type {
	return Type{Kind: TypeMultiple, Components: components}
}

func NewArrayType(elemType Type, size int) Type {
	return Type{Kind: TypeArray, PtrTo: &elemType, ArraySize: size}
}

func NewUserDefinedType(name string, entity Type) Type {
	return Type{Kind: TypeUserDefined, DefinedName: name, PtrTo: &entity}
}

func NewPointerType(to *Type) Type {
	return Type{Kind: TypePtr, PtrTo: to}
}

func NewUndefinedType() Type {
	return NewType(TypeUndefined)
}

func NewSliceType(elemType Type) Type {
	return Type{Kind: TypeSlice, PtrTo: &elemType}
}

func NewStructType(names []string, types []Type) Type {
	ty := Type{Kind: TypeStruct, MemberNames: names, MemberTypes: types}
	ty.MemberOffsets = []int{}
	for range types {
		ty.MemberOffsets = append(ty.MemberOffsets, 0)
	}
	return ty
}

func Sizeof(ty Type) int {
	if ty.Kind == TypeUserDefined {
		return Sizeof(*ty.PtrTo)
	}
	if ty.Kind == TypeInt || ty.Kind == TypePtr || ty.Kind == TypeArray || ty.Kind == TypeSlice || ty.Kind == TypeStruct {
		return 8
	}
	if ty.Kind == TypeRune || ty.Kind == TypeBool {
		return 1
	}
	// 未定義
	return 0
}

func TypeEquals(t1 Type, t2 Type) bool {
	if t1.Kind != t2.Kind {
		return false
	}
	if t1.Kind == TypePtr {
		return TypeEquals(*t1.PtrTo, *t2.PtrTo)
	}
	if t1.Kind == TypeArray {
		return t1.ArraySize == t2.ArraySize && TypeEquals(*t1.PtrTo, *t2.PtrTo)
	}
	if t1.Kind == TypeUserDefined {
		return TypeEquals(*t1.PtrTo, *t2.PtrTo)
	}
	if t1.Kind == TypeStruct {
		if len(t1.MemberNames) != len(t2.MemberNames) {
			return false
		}
		for i := 0; i < len(t1.MemberNames); i++ {
			var ok = false
			for j := 0; j < len(t1.MemberNames); j++ {
				ok = ok || (TypeEquals(t1.MemberTypes[i], t2.MemberTypes[j]) && t1.MemberNames[i] == t2.MemberNames[j])
			}
			if !ok {
				return false
			}
		}
		return true
	}
	if t1.Kind == TypeMultiple {
		if len(t1.Components) != len(t2.Components) {
			return false
		}
		for i := range t1.Components {
			if !TypeEquals(t1.Components[i], t2.Components[i]) {
				return false
			}
		}
		return true
	}
	return true
}

func IsKindOfNumber(t Type) bool {
	return t.Kind == TypeInt || t.Kind == TypeRune
}

func TypeCompatable(t1 Type, t2 Type) bool {
	return (IsKindOfNumber(t1) && IsKindOfNumber(t2)) || TypeEquals(t1, t2)
}
