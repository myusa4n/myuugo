package passes

import (
	"github.com/myuu222/myuugo/lang"
	"github.com/myuu222/myuugo/parse"
	"github.com/myuu222/myuugo/util"
)

var program *parse.Program

func alignLocalVars(functionName string) {
	fn := program.FindFunction(functionName)
	if fn == nil {
		panic("関数 \"" + functionName + " は存在しません")
	}
	var offset = 0
	for _, lvar := range fn.LocalVariables {
		offset += lang.Sizeof(lvar.Type)
		lvar.Offset = offset
	}
}

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
		fn := program.FindFunction(node.Env.FunctionName)
		if fn.ReturnValueType.Kind == lang.TypeVoid {
			if node.Target != nil {
				util.Alarm("返り値の型がvoid型の関数内でreturnに引数を渡すことはできません")
			}
		} else {
			var ty = traverse(node.Target)
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
		traverse(node.If)
		if node.Else != nil {
			traverse(node.Else)
		}
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == parse.NodeIf {
		traverse(node.Condition)
		traverse(node.Body)
		if node.Condition.ExprType.Kind != lang.TypeBool {
			util.Alarm("if文の条件として使える式はbool型のものだけです")
		}
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == parse.NodeElse {
		traverse(node.Body)
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == parse.NodeFor {
		if node.Init != nil {
			traverse(node.Init)
		}
		if node.Condition != nil {
			traverse(node.Condition) // 条件
		}
		if node.Update != nil {
			traverse(node.Update)
		}
		traverse(node.Body)
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == parse.NodeFunctionDef {
		for _, param := range node.Parameters { // 引数
			traverse(param)
		}
		traverse(node.Body) // 関数本体
		alignLocalVars(node.Env.FunctionName)
		node.ExprType = stmtType
		return stmtType
	}
	if node.Kind == parse.NodeNot {
		var ty = traverse(node.Target)
		if ty.Kind != lang.TypeBool {
			panic("否定演算子の後に続くのはbool型の値だけです")
		}
		node.ExprType = lang.NewType(lang.TypeBool)
		return node.ExprType
	}
	if node.Kind == parse.NodeAddr {
		var ty = traverse(node.Target)
		node.ExprType = lang.NewPointerType(&ty)
		return node.ExprType
	}
	if node.Kind == parse.NodeDeref {
		var ty = traverse(node.Target)
		if ty.Kind != lang.TypePtr {
			util.Alarm("ポインタでないものの参照を外そうとしています")
		}
		node.ExprType = *ty.PtrTo
		return *ty.PtrTo
	}
	if node.Kind == parse.NodeFunctionCall {
		fn := program.FindFunction(node.Label)
		if fn != nil {
			if len(fn.ParameterTypes) != len(node.Arguments) {
				util.Alarm("関数%sの引数の数が正しくありません", fn.Label)
			}
			for i, argument := range node.Arguments {
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
			alignLocalVars(node.Env.FunctionName)
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
	if node.Kind == parse.NodeBool {
		node.ExprType = lang.NewType(lang.TypeBool)
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
		var seqType = traverse(node.Seq)
		var indexType = traverse(node.Index)
		if seqType.Kind != lang.TypeArray {
			util.Alarm("配列ではないものに添字でアクセスしようとしています")
		}
		if !lang.IsKindOfNumber(indexType) {
			util.Alarm("配列の添字は整数でなくてはなりません")
		}
		node.ExprType = *seqType.PtrTo
		return node.ExprType
	}

	var lhsType = traverse(node.Lhs)
	var rhsType = traverse(node.Rhs)

	if !lang.TypeCompatable(lhsType, rhsType) {
		util.Alarm("[%s] 左辺と右辺の式の型が違います %s %s", node.Kind, lhsType.Kind, rhsType.Kind)
	}

	switch node.Kind {
	case parse.NodeAdd, parse.NodeSub, parse.NodeMul, parse.NodeDiv:
		node.ExprType = lang.NewType(lang.TypeInt)
	case parse.NodeEql, parse.NodeNotEql, parse.NodeLess, parse.NodeLessEql, parse.NodeGreater, parse.NodeGreaterEql:
		node.ExprType = lang.NewType(lang.TypeBool)
	case parse.NodeLogicalAnd, parse.NodeLogicalOr:
		// 両辺がBoolであることを期待
		if lhsType.Kind != lang.TypeBool {
			util.Alarm("&&の両辺の値はbool型の値でなくてはなりません")
		}
		node.ExprType = lang.NewType(lang.TypeBool)
	default:
		node.ExprType = stmtType
	}
	return node.ExprType
}
