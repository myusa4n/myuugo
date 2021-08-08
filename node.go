package main

type NodeKind string

const (
	NodeAdd             NodeKind = "ADD"               // +
	NodeSub             NodeKind = "SUB"               // -
	NodeMul             NodeKind = "MUL"               // *
	NodeDiv             NodeKind = "DIV"               // /
	NodeEql             NodeKind = "EQL"               // ==
	NodeNotEql          NodeKind = "NOT EQL"           // !=
	NodeLess            NodeKind = "LESS"              // <
	NodeLessEql         NodeKind = "LESS EQL"          // <=
	NodeGreater         NodeKind = "GREATER"           // >
	NodeGreaterEql      NodeKind = "GREATER EQL"       // >=
	NodeAssign          NodeKind = "ASSIGN"            // =
	NodeReturn          NodeKind = "RETURN"            // return
	NodeVariable        NodeKind = "VARIABLE"          // 変数参照
	NodeNum             NodeKind = "NUM"               // 整数
	NodeMetaIf          NodeKind = "META IF"           // if ... else ...
	NodeIf              NodeKind = "IF"                // if
	NodeElse            NodeKind = "ELSE"              // else
	NodeStmtList        NodeKind = "STMT LIST"         // stmt*
	NodeFor             NodeKind = "FOR"               // for
	NodeFunctionCall    NodeKind = "FUNCTION CALL"     // fn()
	NodeFunctionDef     NodeKind = "FUNCTION DEF"      // func fn() { ... }
	NodeAddr            NodeKind = "ADDR"              // &
	NodeDeref           NodeKind = "DEREF"             // *addr
	NodeLocalVarStmt    NodeKind = "LOCAL VAR STMT"    // (local) var ...
	NodeTopLevelVarStmt NodeKind = "TOPLEVEL VAR STMT" // (toplevel) var ...
	NodePackageStmt     NodeKind = "PACKAGE STMT"      // package ...
	NodeExprStmt        NodeKind = "EXPR STMT"         // 式文
	NodeIndex           NodeKind = "INDEX"             // 添字アクセス
)

type Node struct {
	kind     NodeKind  // ノードの型
	val      int       // kindがNodeNumの場合にのみ使う
	variable *Variable // kindがNodeLocalVarの場合にのみ使う
	label    string    // kindがNodeFunctionCallまたはNodePackageの場合にのみ使う
	exprType Type      // ノードが表す式の型
	children []*Node   // 子。lhs, rhsの順でchildrenに格納される
}

func NewNode(kind NodeKind, children []*Node) *Node {
	return &Node{kind: kind, children: children}
}

func NewBinaryNode(kind NodeKind, lhs *Node, rhs *Node) *Node {
	return &Node{kind: kind, children: []*Node{lhs, rhs}}
}

func NewLeafNode(kind NodeKind) *Node {
	return &Node{kind: kind}
}

func NewNodeNum(val int) *Node {
	return &Node{kind: NodeNum, val: val}
}
