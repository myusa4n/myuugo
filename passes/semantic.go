package pipeline

import (
	. "github.com/myuu222/myuugo/lang"
	. "github.com/myuu222/myuugo/parse"
	. "github.com/myuu222/myuugo/util"
)

func Semantic(program *Program) {
	for _, node := range program.Code {
		traverse(node)
	}
}

// 式の型を決定するのに使う
func traverse(node *Node) Type {
	var stmtType = NewType(TypeStmt)
	if node.Kind == NodePackageStmt {
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == NodeLocalVarList {
		for _, c := range node.Children {
			traverse(c)
		}
		// とりあえず
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == NodeExprList {
		var types = []Type{}
		for _, c := range node.Children {
			types = append(types, traverse(c))
		}
		if len(types) > 1 {
			node.ExprType = NewMultipleType(types)
		} else {
			node.ExprType = types[0]
		}
		return node.ExprType
	}
	if node.Kind == NodeStmtList {
		for _, stmt := range node.Children {
			traverse(stmt)
		}
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == NodeReturn {
		fn := node.Env.FindFunction(Env.FunctionName)
		if fn.ReturnValueType.Kind == TypeVoid {
			if len(node.Children) > 0 {
				Alarm("返り値の型がvoid型の関数内でreturnに引数を渡すことはできません")
			}
		} else {
			var ty = traverse(node.Children[0])
			if !TypeCompatable(fn.ReturnValueType, ty) {
				Alarm("返り値の型とreturnの引数の型が一致しません")
			}
		}
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == NodeAssign {
		var lhs = node.Children[0]
		var rhs = node.Children[1]
		var ltype = traverse(lhs)
		var rtype = traverse(rhs)

		if !TypeCompatable(ltype, rtype) {
			Alarm("代入式の左辺と右辺の型が違います ")
		}

		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == NodeShortVarDeclStmt {
		var lhs = node.Children[0]
		var rhs = node.Children[1]
		traverse(lhs)
		var rhsType = traverse(rhs)

		if rhsType.Kind == TypeMultiple {
			// componentの数だけ左辺のパラメータが存在していないといけない
			if len(lhs.Children) != len(rhsType.Components) {
				Alarm(":=の左辺に要求されているパラメータの数は%dです", len(rhsType.Components))
			}
		} else {
			if len(lhs.Children) != len(rhs.Children) {
				Alarm(":=の左辺と右辺のパラメータの数が異なります")
			}
		}
		for i, l := range lhs.Children {
			if rhsType.Kind == TypeMultiple {
				l.Variable.Type = rhsType.Components[i]
				l.ExprType = rhsType.Components[i]
			} else {
				l.Variable.Type = rhsType
				l.ExprType = rhsType
			}
		}

		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == NodeMetaIf {
		traverse(node.Children[0])
		if node.Children[1] != nil {
			traverse(node.Children[1])
		}
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == NodeIf {
		traverse(node.Children[0]) // lhs
		traverse(node.Children[1]) // rhs
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == NodeElse {
		traverse(node.Children[0])
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == NodeFor {
		// children := (初期化, 条件, 更新)
		if node.Children[0] != nil {
			traverse(node.Children[0])
		}
		if node.Children[1] != nil {
			traverse(node.Children[1]) // 条件
		}
		if node.Children[2] != nil {
			traverse(node.Children[2])
		}
		traverse(node.Children[3])
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == NodeFunctionDef {
		for _, param := range node.Children[1:] { // 引数
			traverse(param)
		}
		traverse(node.Children[0]) // 関数本体
		node.Env.AlignLocalVars(Env.FunctionName)
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == NodeAddr {
		var ty = traverse(node.Children[0])
		node.ExprType = NewPointerType(&ty)
		return node.ExprType
	}
	if node.Kind == NodeDeref {
		var ty = traverse(node.Children[0])
		if ty.Kind != TypePtr {
			Alarm("ポインタでないものの参照を外そうとしています")
		}
		node.ExprType = *ty.PtrTo
		return *ty.PtrTo
	}
	if node.Kind == NodeFunctionCall {
		fn := node.Env.FindFunction(node.Label)
		if fn != nil {
			if len(fn.ParameterTypes) != len(node.Children) {
				Alarm("関数%sの引数の数が正しくありません", fn.Label)
			}
			for i, argument := range node.Children {
				if !TypeCompatable(fn.ParameterTypes[i], traverse(argument)) {
					Alarm("関数%sの%d番目の引数の型が一致しません", fn.Label, i)
				}
			}
			node.ExprType = fn.ReturnValueType
			return fn.ReturnValueType
		}
		return node.ExprType // おそらくundefined
	}
	if node.Kind == NodeLocalVarStmt || node.Kind == NodeTopLevelVarStmt {
		if len(node.Children) == 2 {
			var lvarType = traverse(node.Children[0])
			var valueType = traverse(node.Children[1])

			if lvarType.Kind == TypeUndefined {
				node.Children[0].Variable.Type = valueType
				node.Children[0].ExprType = valueType
				lvarType = valueType
			}
			if !TypeCompatable(lvarType, valueType) {
				Alarm("var文における変数の型と初期化式の型が一致しません")
			}
		}
		if node.Kind == NodeLocalVarList {
			node.Env.AlignLocalVars(Env.FunctionName)
		}
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == NodeExprStmt {
		traverse(node.Children[0])
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == NodeNum {
		node.ExprType = NewType(TypeInt)
		return node.ExprType
	}
	if node.Kind == NodeVariable {
		node.ExprType = node.Variable.Type
		return node.Variable.Type
	}
	if node.Kind == NodeString {
		var runeType = NewType(TypeRune)
		node.ExprType = NewPointerType(&runeType)
		return node.ExprType
	}
	if node.Kind == NodeIndex {
		var lhsType = traverse(node.Children[0])
		var rhsType = traverse(node.Children[1])
		if lhsType.Kind != TypeArray {
			Alarm("配列ではないものに添字でアクセスしようとしています")
		}
		if !IsKindOfNumber(rhsType) {
			Alarm("配列の添字は整数でなくてはなりません")
		}
		node.ExprType = *lhsType.PtrTo
		return *lhsType.PtrTo
	}

	var lhsType = traverse(node.Children[0])
	var rhsType = traverse(node.Children[1])

	if !TypeCompatable(lhsType, rhsType) {
		Alarm("[%s] 左辺と右辺の式の型が違います %s %s", node.Kind, lhsType.Kind, rhsType.Kind)
	}

	node.ExprType = NewType(TypeInt)
	switch node.Kind {
	case NodeAdd:
		return NewType(TypeInt)
	case NodeSub:
		return NewType(TypeInt)
	case NodeMul:
		return NewType(TypeInt)
	case NodeDiv:
		return NewType(TypeInt)
	case NodeEql:
		return NewType(TypeInt)
	case NodeNotEql:
		return NewType(TypeInt)
	case NodeLess:
		return NewType(TypeInt)
	case NodeLessEql:
		return NewType(TypeInt)
	case NodeGreater:
		return NewType(TypeInt)
	case NodeGreaterEql:
		return NewType(TypeInt)
	}
	node.ExprType = stmtType
	return stmtType
}
