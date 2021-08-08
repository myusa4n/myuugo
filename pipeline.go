package main

func pipeline(code []*Node) {
	for _, node := range code {
		traverse(node)
	}
}

// 式の型を決定するのに使う
func traverse(node *Node) Type {
	var stmtType = Type{kind: TypeStmt}
	if node.kind == NodePackageStmt {
		node.exprType = stmtType
		return stmtType
	}
	if node.kind == NodeStmtList {
		for _, stmt := range node.children {
			traverse(stmt)
		}
		node.exprType = stmtType
		return stmtType
	}
	if node.kind == NodeReturn {
		traverse(node.children[0])
		node.exprType = stmtType
		return stmtType
	}
	if node.kind == NodeAssign {
		var lhsType = traverse(node.children[0]) // lhs
		var rhsType = traverse(node.children[1]) // rhs

		if !TypeCompatable(lhsType, rhsType) {
			madden("代入式の左辺と右辺の型が違います ")
		}
		node.exprType = stmtType
		return stmtType
	}
	if node.kind == NodeMetaIf {
		traverse(node.children[0])
		if node.children[1] != nil {
			traverse(node.children[1])
		}
		node.exprType = stmtType
		return stmtType
	}
	if node.kind == NodeIf {
		traverse(node.children[0]) // lhs
		traverse(node.children[1]) // rhs
		node.exprType = stmtType
		return stmtType
	}
	if node.kind == NodeElse {
		traverse(node.children[0])
		node.exprType = stmtType
		return stmtType
	}
	if node.kind == NodeFor {
		// children := (初期化, 条件, 更新)
		if node.children[0] != nil {
			traverse(node.children[0])
		}
		if node.children[1] != nil {
			traverse(node.children[1]) // 条件
		}
		if node.children[2] != nil {
			traverse(node.children[2])
		}
		traverse(node.children[3])
		node.exprType = stmtType
		return stmtType
	}
	if node.kind == NodeFunctionDef {
		var prevFuncLabel = currentFuncLabel
		currentFuncLabel = node.label
		for _, param := range node.children[1:] { // 引数
			traverse(param)
		}
		Env.AlignLocalVars(currentFuncLabel)
		traverse(node.children[0]) // 関数本体
		currentFuncLabel = prevFuncLabel
		node.exprType = stmtType
		return stmtType
	}
	if node.kind == NodeAddr {
		var ty = traverse(node.children[0])
		node.exprType = Type{kind: TypePtr, ptrTo: &ty}
		return Type{kind: TypePtr, ptrTo: &ty}
	}
	if node.kind == NodeDeref {
		var ty = traverse(node.children[0])
		if ty.kind != TypePtr {
			madden("ポインタでないものの参照を外そうとしています")
		}
		node.exprType = *ty.ptrTo
		return *ty.ptrTo
	}
	if node.kind == NodeFunctionCall {
		fn, ok := Env.FunctionTable[node.label]

		if !ok {
			madden("関数%sは定義されていません", fn.Label)
		}
		if len(fn.ParameterTypes) != len(node.children) {
			madden("関数%sの引数の数が正しくありません", fn.Label)
		}
		for i, argument := range node.children {
			if !TypeCompatable(fn.ParameterTypes[i], traverse(argument)) {
				madden("関数%sの%d番目の引数の型が一致しません", fn.Label, i)
			}
		}
		node.exprType = fn.ReturnValueType
		return fn.ReturnValueType
	}
	if node.kind == NodeLocalVarStmt || node.kind == NodeTopLevelVarStmt {
		if len(node.children) == 2 {
			var lvarType = traverse(node.children[0])
			var valueType = traverse(node.children[1])

			if lvarType.kind == TypeUndefined {
				node.children[0].variable.varType = valueType
				node.children[0].exprType = valueType
				lvarType = valueType
			}
			if !TypeCompatable(lvarType, valueType) {
				madden("var文における変数の型と初期化式の型が一致しません")
			}
		}
		if currentFuncLabel != "" {
			Env.AlignLocalVars(currentFuncLabel)
		}
		node.exprType = stmtType
		return stmtType
	}
	if node.kind == NodeExprStmt {
		traverse(node.children[0])
		node.exprType = stmtType
		return stmtType
	}
	if node.kind == NodeNum {
		node.exprType = NewType(TypeInt)
		return Type{kind: TypeInt}
	}
	if node.kind == NodeVariable {
		node.exprType = node.variable.varType
		return node.variable.varType
	}
	if node.kind == NodeIndex {
		var lhsType = traverse(node.children[0])
		var rhsType = traverse(node.children[1])
		if lhsType.kind != TypeArray {
			madden("配列ではないものに添字でアクセスしようとしています")
		}
		if !IsKindOfNumber(rhsType) {
			madden("配列の添字は整数でなくてはなりません")
		}
		node.exprType = *lhsType.ptrTo
		return *lhsType.ptrTo
	}

	var lhsType = traverse(node.children[0])
	var rhsType = traverse(node.children[1])

	if !TypeCompatable(lhsType, rhsType) {
		madden("[%s] 左辺と右辺の式の型が違います %s %s", node.kind, lhsType.kind, rhsType.kind)
	}

	node.exprType = NewType(TypeInt)
	switch node.kind {
	case NodeAdd:
		return Type{kind: TypeInt}
	case NodeSub:
		return Type{kind: TypeInt}
	case NodeMul:
		return Type{kind: TypeInt}
	case NodeDiv:
		return Type{kind: TypeInt}
	case NodeEql:
		return Type{kind: TypeInt}
	case NodeNotEql:
		return Type{kind: TypeInt}
	case NodeLess:
		return Type{kind: TypeInt}
	case NodeLessEql:
		return Type{kind: TypeInt}
	case NodeGreater:
		return Type{kind: TypeInt}
	case NodeGreaterEql:
		return Type{kind: TypeInt}
	}
	node.exprType = stmtType
	return stmtType
}
