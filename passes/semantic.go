package passes

import (
	"github.com/myuu222/myuugo/lang"
	"github.com/myuu222/myuugo/parse"
	"github.com/myuu222/myuugo/util"
)

var program *parse.Program

func Semantic(p *parse.Program) {
	program = p
	for _, node := range p.Code {
		traverse(node)
	}
}

// 式の型を決定するのに使う
func traverse(node *parse.Node) lang.Type {
	var stmtType = lang.NewType(lang.TypeStmt)
	if node.Kind == parse.NodePackageStmt {
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == parse.NodeLocalVarList {
		for _, c := range node.Children {
			traverse(c)
		}
		// とりあえず
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == parse.NodeExprList {
		var types = []lang.Type{}
		for _, c := range node.Children {
			types = append(types, traverse(c))
		}
		if len(types) > 1 {
			node.ExprType = lang.NewMultipleType(types)
		} else {
			node.ExprType = types[0]
		}
		return node.ExprType
	}
	if node.Kind == parse.NodeStmtList {
		for _, stmt := range node.Children {
			traverse(stmt)
		}
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == parse.NodeReturn {
		fn := node.Env.FindFunction(node.Env.FunctionName)
		if fn.ReturnValueType.Kind == lang.TypeVoid {
			if len(node.Children) > 0 {
				util.Alarm("返り値の型がvoid型の関数内でreturnに引数を渡すことはできません")
			}
		} else {
			var ty = traverse(node.Children[0])
			if !lang.TypeCompatable(fn.ReturnValueType, ty) {
				util.Alarm("返り値の型とreturnの引数の型が一致しません")
			}
		}
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == parse.NodeAssign {
		var lhs = node.Children[0]
		var rhs = node.Children[1]
		var ltype = traverse(lhs)
		var rtype = traverse(rhs)

		if !lang.TypeCompatable(ltype, rtype) {
			util.Alarm("代入式の左辺と右辺の型が違います ")
		}

		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == parse.NodeShortVarDeclStmt {
		var lhs = node.Children[0]
		var rhs = node.Children[1]
		traverse(lhs)
		var rhsType = traverse(rhs)

		if rhsType.Kind == lang.TypeMultiple {
			// componentの数だけ左辺のパラメータが存在していないといけない
			if len(lhs.Children) != len(rhsType.Components) {
				util.Alarm(":=の左辺に要求されているパラメータの数は%dです", len(rhsType.Components))
			}
		} else {
			if len(lhs.Children) != len(rhs.Children) {
				util.Alarm(":=の左辺と右辺のパラメータの数が異なります")
			}
		}
		for i, l := range lhs.Children {
			if rhsType.Kind == lang.TypeMultiple {
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
	if node.Kind == parse.NodeMetaIf {
		traverse(node.Children[0])
		if node.Children[1] != nil {
			traverse(node.Children[1])
		}
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == parse.NodeIf {
		traverse(node.Children[0]) // lhs
		traverse(node.Children[1]) // rhs
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == parse.NodeElse {
		traverse(node.Children[0])
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == parse.NodeFor {
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
	if node.Kind == parse.NodeFunctionDef {
		for _, param := range node.Children[1:] { // 引数
			traverse(param)
		}
		traverse(node.Children[0]) // 関数本体
		node.Env.AlignLocalVars(node.Env.FunctionName)
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == parse.NodeAddr {
		var ty = traverse(node.Children[0])
		node.ExprType = lang.NewPointerType(&ty)
		return node.ExprType
	}
	if node.Kind == parse.NodeDeref {
		var ty = traverse(node.Children[0])
		if ty.Kind != lang.TypePtr {
			util.Alarm("ポインタでないものの参照を外そうとしています")
		}
		node.ExprType = *ty.PtrTo
		return *ty.PtrTo
	}
	if node.Kind == parse.NodeFunctionCall {
		fn := node.Env.FindFunction(node.Label)
		if fn != nil {
			if len(fn.ParameterTypes) != len(node.Children) {
				util.Alarm("関数%sの引数の数が正しくありません", fn.Label)
			}
			for i, argument := range node.Children {
				if !lang.TypeCompatable(fn.ParameterTypes[i], traverse(argument)) {
					util.Alarm("関数%sの%d番目の引数の型が一致しません", fn.Label, i)
				}
			}
			node.ExprType = fn.ReturnValueType
			return fn.ReturnValueType
		}
		return node.ExprType // おそらくundefined
	}
	if node.Kind == parse.NodeLocalVarStmt || node.Kind == parse.NodeTopLevelVarStmt {
		if len(node.Children) == 2 {
			var lvarType = traverse(node.Children[0])
			var valueType = traverse(node.Children[1])

			if lvarType.Kind == lang.TypeUndefined {
				node.Children[0].Variable.Type = valueType
				node.Children[0].ExprType = valueType
				lvarType = valueType
			}
			if !lang.TypeCompatable(lvarType, valueType) {
				util.Alarm("var文における変数の型と初期化式の型が一致しません")
			}
		}
		if node.Kind == parse.NodeLocalVarList {
			node.Env.AlignLocalVars(node.Env.FunctionName)
		}
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == parse.NodeExprStmt {
		traverse(node.Children[0])
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == parse.NodeNum {
		node.ExprType = lang.NewType(lang.TypeInt)
		return node.ExprType
	}
	if node.Kind == parse.NodeVariable {
		node.ExprType = node.Variable.Type
		return node.Variable.Type
	}
	if node.Kind == parse.NodeString {
		var runeType = lang.NewType(lang.TypeRune)
		node.ExprType = lang.NewPointerType(&runeType)
		return node.ExprType
	}
	if node.Kind == parse.NodeIndex {
		var lhsType = traverse(node.Children[0])
		var rhsType = traverse(node.Children[1])
		if lhsType.Kind != lang.TypeArray {
			util.Alarm("配列ではないものに添字でアクセスしようとしています")
		}
		if !lang.IsKindOfNumber(rhsType) {
			util.Alarm("配列の添字は整数でなくてはなりません")
		}
		node.ExprType = *lhsType.PtrTo
		return *lhsType.PtrTo
	}

	var lhsType = traverse(node.Children[0])
	var rhsType = traverse(node.Children[1])

	if !lang.TypeCompatable(lhsType, rhsType) {
		util.Alarm("[%s] 左辺と右辺の式の型が違います %s %s", node.Kind, lhsType.Kind, rhsType.Kind)
	}

	node.ExprType = lang.NewType(lang.TypeInt)
	switch node.Kind {
	case parse.NodeAdd:
		return lang.NewType(lang.TypeInt)
	case parse.NodeSub:
		return lang.NewType(lang.TypeInt)
	case parse.NodeMul:
		return lang.NewType(lang.TypeInt)
	case parse.NodeDiv:
		return lang.NewType(lang.TypeInt)
	case parse.NodeEql:
		return lang.NewType(lang.TypeInt)
	case parse.NodeNotEql:
		return lang.NewType(lang.TypeInt)
	case parse.NodeLess:
		return lang.NewType(lang.TypeInt)
	case parse.NodeLessEql:
		return lang.NewType(lang.TypeInt)
	case parse.NodeGreater:
		return lang.NewType(lang.TypeInt)
	case parse.NodeGreaterEql:
		return lang.NewType(lang.TypeInt)
	}
	node.ExprType = stmtType
	return stmtType
}
