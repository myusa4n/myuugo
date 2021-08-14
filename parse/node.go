package parse

import (
	"github.com/myuu222/myuugo/lang"
)

type NodeKind string

const (
	NodeAdd              NodeKind = "ADD"                 // +
	NodeSub              NodeKind = "SUB"                 // -
	NodeMul              NodeKind = "MUL"                 // *
	NodeDiv              NodeKind = "DIV"                 // /
	NodeEql              NodeKind = "EQL"                 // ==
	NodeNotEql           NodeKind = "NOT EQL"             // !=
	NodeLess             NodeKind = "LESS"                // <
	NodeLessEql          NodeKind = "LESS EQL"            // <=
	NodeGreater          NodeKind = "GREATER"             // >
	NodeGreaterEql       NodeKind = "GREATER EQL"         // >=
	NodeAssign           NodeKind = "ASSIGN"              // =
	NodeReturn           NodeKind = "RETURN"              // return
	NodeVariable         NodeKind = "VARIABLE"            // 変数参照
	NodeNum              NodeKind = "NUM"                 // 整数
	NodeMetaIf           NodeKind = "META IF"             // if ... else ...
	NodeIf               NodeKind = "IF"                  // if
	NodeElse             NodeKind = "ELSE"                // else
	NodeStmtList         NodeKind = "STMT LIST"           // stmt*
	NodeFor              NodeKind = "FOR"                 // for
	NodeFunctionCall     NodeKind = "FUNCTION CALL"       // fn()
	NodeFunctionDef      NodeKind = "FUNCTION DEF"        // func fn() { ... }
	NodeAddr             NodeKind = "ADDR"                // &
	NodeDeref            NodeKind = "DEREF"               // *addr
	NodeLocalVarStmt     NodeKind = "LOCAL VAR STMT"      // (local) var ...
	NodeTopLevelVarStmt  NodeKind = "TOPLEVEL VAR STMT"   // (toplevel) var ...
	NodePackageStmt      NodeKind = "PACKAGE STMT"        // package ...
	NodeExprStmt         NodeKind = "EXPR STMT"           // 式文
	NodeIndex            NodeKind = "INDEX"               // 添字アクセス
	NodeString           NodeKind = "STRING"              // 文字列
	NodeShortVarDeclStmt NodeKind = "SHORT VAR DECL STMT" // 短絡変数宣言
	NodeExprList         NodeKind = "EXPR LIST"           // 複数の要素からなる式
	NodeLocalVarList     NodeKind = "LOCAL VAR LIST"      // 複数の変数からなる式
)

type Node struct {
	Kind     NodeKind            // ノードの型
	Val      int                 // kindがNodeNumの場合にのみ使う
	Variable *lang.Variable      // kindがNodeLocalVarの場合にのみ使う
	Str      *lang.StringLiteral // kindがNodeStringの場合にのみ使う
	Label    string              // kindがNodeFunctionCallまたはNodePackageの場合にのみ使う
	ExprType lang.Type           // ノードが表す式の型
	Children []*Node             // 子。
	Env      *Environment        // そのノードで管理している変数などの情報をまとめたもの

	// 二項演算を行うノードの場合にのみ使う
	Lhs *Node
	Rhs *Node

	// kindがNodeIndexの場合にのみ使う
	Seq   *Node
	Index *Node

	// kindがNodeMetaIfの場合にのみ使う
	If   *Node
	Else *Node

	// kindがNodeFunctionDef, NodeIf, NodeElse, NodeForの場合にのみ使う
	Body *Node

	// kindがNodeIf, NodeForの場合にのみ使う
	Condition *Node

	// kindがNodeForの場合にのみ使う
	// for Init; Condition; Update {}
	Init   *Node
	Update *Node

	// kindがNodeFunctionDefの場合にのみ使う
	Parameters []*Node

	// kindがNodeFunctionCallの場合にのみ使う
	Arguments []*Node

	// kindがNodeReturn, NodeAddr, NodeDerefの場合にのみ使う
	Target *Node
}

func NewFunctionDefNode(name string, parameters []*Node, body *Node) *Node {
	return &Node{Kind: NodeFunctionDef, Label: name, Parameters: parameters, Body: body, Env: Env}
}

func NewFunctionCallNode(name string, arguments []*Node) *Node {
	return &Node{Kind: NodeFunctionCall, Label: name, Arguments: arguments, Env: Env}
}

func NewNode(kind NodeKind, children []*Node) *Node {
	return &Node{Kind: kind, Children: children, Env: Env}
}

func NewBinaryNode(kind NodeKind, lhs *Node, rhs *Node) *Node {
	return &Node{Kind: kind, Children: []*Node{lhs, rhs}, Env: Env}
}

func NewBinaryOperationNode(kind NodeKind, lhs *Node, rhs *Node) *Node {
	return &Node{Kind: kind, Lhs: lhs, Rhs: rhs, Env: Env}
}

func NewUnaryOperationNode(kind NodeKind, target *Node) *Node {
	return &Node{Kind: kind, Target: target, Env: Env}
}

func NewIndexNode(seq *Node, index *Node) *Node {
	return &Node{Kind: NodeIndex, Seq: seq, Index: index, Env: Env}
}

func NewMetaIfNode(ifn *Node, elsen *Node) *Node {
	return &Node{Kind: NodeMetaIf, If: ifn, Else: elsen, Env: Env}
}

func NewIfNode(cond *Node, body *Node) *Node {
	return &Node{Kind: NodeIf, Condition: cond, Body: body, Env: Env}
}

func NewElseNode(body *Node) *Node {
	return &Node{Kind: NodeElse, Body: body, Env: Env}
}

func NewLeafNode(kind NodeKind) *Node {
	return &Node{Kind: kind, Env: Env}
}

func NewNodeNum(val int) *Node {
	return &Node{Kind: NodeNum, Val: val, Env: Env}
}

func NewForNode(init *Node, cond *Node, update *Node, body *Node) *Node {
	return &Node{Kind: NodeFor, Init: init, Condition: cond, Update: update, Body: body}
}
