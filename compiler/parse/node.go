package parse

import (
	"github.com/myuu222/myuugo/compiler/lang"
)

type NodeKind string

const (
	NodeAdd                          NodeKind = "ADD"                                   // +
	NodeSub                          NodeKind = "SUB"                                   // -
	NodeMul                          NodeKind = "MUL"                                   // *
	NodeDiv                          NodeKind = "DIV"                                   // /
	NodeMod                          NodeKind = "[NODE] MOD"                            // %
	NodeEql                          NodeKind = "EQL"                                   // ==
	NodeNotEql                       NodeKind = "NOT EQL"                               // !=
	NodeLess                         NodeKind = "LESS"                                  // <
	NodeLessEql                      NodeKind = "LESS EQL"                              // <=
	NodeGreater                      NodeKind = "GREATER"                               // >
	NodeGreaterEql                   NodeKind = "GREATER EQL"                           // >=
	NodeAssign                       NodeKind = "ASSIGN"                                // =
	NodeReturn                       NodeKind = "RETURN"                                // return
	NodeTopLevelVariable             NodeKind = "[NODE] TOP LEVEL VARIABLE"             // トップレベル変数参照
	NodeLocalVariable                NodeKind = "[NODE] LOCAL VARIABLE"                 // ローカル変数参照
	NodeNum                          NodeKind = "NUM"                                   // 整数
	NodeBool                         NodeKind = "BOOL"                                  // 真偽値
	NodeMetaIf                       NodeKind = "META IF"                               // if ... else ...
	NodeIf                           NodeKind = "IF"                                    // if
	NodeElse                         NodeKind = "ELSE"                                  // else
	NodeStmtList                     NodeKind = "STMT LIST"                             // stmt*
	NodeFor                          NodeKind = "FOR"                                   // for
	NodeFunctionCall                 NodeKind = "FUNCTION CALL"                         // fn()
	NodeFunctionDef                  NodeKind = "FUNCTION DEF"                          // func fn() { ... }
	NodeAddr                         NodeKind = "ADDR"                                  // &
	NodeDeref                        NodeKind = "DEREF"                                 // *addr
	NodeLocalVarStmt                 NodeKind = "LOCAL VAR STMT"                        // (local) var ...
	NodeTopLevelVarStmt              NodeKind = "TOPLEVEL VAR STMT"                     // (toplevel) var ...
	NodePackageStmt                  NodeKind = "PACKAGE STMT"                          // package ...
	NodeExprStmt                     NodeKind = "EXPR STMT"                             // 式文
	NodeIndex                        NodeKind = "INDEX"                                 // 添字アクセス
	NodeString                       NodeKind = "STRING"                                // 文字列
	NodeShortVarDeclStmt             NodeKind = "SHORT VAR DECL STMT"                   // 短絡変数宣言
	NodeExprList                     NodeKind = "EXPR LIST"                             // 複数の要素からなる式
	NodeLocalVarList                 NodeKind = "LOCAL VAR LIST"                        // 複数の変数からなる式
	NodeNot                          NodeKind = "[NODE] NOT"                            // 否定
	NodeLogicalAnd                   NodeKind = "[NODE] LOGICAL AND"                    // 論理積
	NodeLogicalOr                    NodeKind = "[NODE] LOGICAL OR"                     // 論理和
	NodeDot                          NodeKind = "[NODE] DOT"                            // A.B
	NodeAppendCall                   NodeKind = "[NODE] APPEND CALL"                    // append(..., ...)
	NodeStringCall                   NodeKind = "[NODE] STRING CALL"                    // string(...)
	NodeRuneCall                     NodeKind = "[NODE] RUNE CALL"                      // rune(...)
	NodeLenCall                      NodeKind = "[NODE] LEN CALL"                       // len(...)
	NodeSliceLiteral                 NodeKind = "[NODE] SLICE LITERAL"                  // []type{...}
	NodeStructLiteral                NodeKind = "[NODE] STRUCT LITERAL"                 // typeName{...}
	NodeTypeStmt                     NodeKind = "[NODE] TYPE STMT"                      // type A struct{}
	NodeImportStmt                   NodeKind = "[NODE] IMPORT STMT"                    // import (
	NodeStatementFunctionDeclaration NodeKind = "[NODE] STATEMENT FUNCTION DECLARATION" // 関数宣言
	NodePackageDot                   NodeKind = "[NODE] PACKAGE DOT"
)

type Node struct {
	Kind     NodeKind            // ノードの型
	Val      int                 // kindがNodeNumの場合にのみ使う
	Variable *lang.Variable      // kindがNodeLocalVarの場合にのみ使う
	Str      *lang.StringLiteral // kindがNodeStringの場合にのみ使う
	Label    string              // kindがNodeFunctionCallまたはNodePackage、NodePackageStmtの場合にのみ使う
	ExprType lang.Type           // ノードが表す式の型
	Children []*Node             // 子。
	Env      *Environment        // そのノードで管理している変数などの情報をまとめたもの
	In       string              // 関数や変数が属している名前

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

	// kindがNodeDotの場合にのみ使う
	Owner      *Node
	MemberName string

	// kindがNodeSliceLiteralまたはNodeStructLiteralの場合にのみ使う
	LiteralType lang.Type

	// kindがNodeStructLiteralの場合にのみ使う
	MemberNames  []string
	MemberValues []*Node

	// kindがNodeImportStmtの場合にのみ使う
	Packages []string
}

func newNodeBase(kind NodeKind) *Node {
	return &Node{Kind: kind, Env: Env, In: Env.program.Name}
}

func NewFunctionDefNode(name string, parameters []*Node, body *Node) *Node {
	node := newNodeBase(NodeFunctionDef)
	node.Label = name
	node.Parameters = parameters
	node.Body = body
	return node
}

func NewFunctionCallNode(name string, arguments []*Node) *Node {
	node := newNodeBase(NodeFunctionCall)
	node.Label = name
	node.Arguments = arguments
	return node
}

func NewNode(kind NodeKind, children []*Node) *Node {
	node := newNodeBase(kind)
	node.Children = children
	return node
}

func NewBinaryNode(kind NodeKind, lhs *Node, rhs *Node) *Node {
	return NewNode(kind, []*Node{lhs, rhs})
}

func NewBinaryOperationNode(kind NodeKind, lhs *Node, rhs *Node) *Node {
	node := newNodeBase(kind)
	node.Lhs = lhs
	node.Rhs = rhs
	return node
}

func NewUnaryOperationNode(kind NodeKind, target *Node) *Node {
	node := newNodeBase(kind)
	node.Target = target
	return node
}

func NewIndexNode(seq *Node, index *Node) *Node {
	node := newNodeBase(NodeIndex)
	node.Seq = seq
	node.Index = index
	return node
}

func NewMetaIfNode(ifn *Node, elsen *Node) *Node {
	node := newNodeBase(NodeMetaIf)
	node.If = ifn
	node.Else = elsen
	return node
}

func NewIfNode(cond *Node, body *Node) *Node {
	node := newNodeBase(NodeIf)
	node.Condition = cond
	node.Body = body
	return node
}

func NewElseNode(body *Node) *Node {
	node := newNodeBase(NodeElse)
	node.Body = body
	return node
}

func NewLeafNode(kind NodeKind) *Node {
	return NewNode(kind, []*Node{})
}

func NewNodeNum(val int) *Node {
	node := newNodeBase(NodeNum)
	node.Val = val
	return node
}

func NewNodeBool(val int) *Node {
	node := newNodeBase(NodeBool)
	node.Val = val
	return node
}

func NewTypeStmtNode() *Node {
	return newNodeBase(NodeTypeStmt)
}

func NewSliceLiteral(ty lang.Type, elements []*Node) *Node {
	n := newNodeBase(NodeSliceLiteral)
	n.LiteralType = ty
	n.Children = elements
	return n
}

func NewStructLiteral(ty lang.Type, memberNames []string, memberValues []*Node) *Node {
	n := newNodeBase(NodeStructLiteral)
	n.LiteralType = ty
	n.MemberNames = memberNames
	n.MemberValues = memberValues
	return n
}

func NewForNode(init *Node, cond *Node, update *Node, body *Node) *Node {
	n := newNodeBase(NodeFor)
	n.Init = init
	n.Condition = cond
	n.Update = update
	n.Body = body
	return n
}

func NewDotNode(owner *Node, memberName string) *Node {
	n := newNodeBase(NodeDot)
	n.Owner = owner
	n.MemberName = memberName
	return n
}

func NewImportStmtNode(packages []string) *Node {
	n := newNodeBase(NodeImportStmt)
	n.Packages = packages
	return n
}

func NewAppendCallNode(arg1 *Node, arg2 *Node) *Node {
	n := newNodeBase(NodeAppendCall)
	n.Arguments = []*Node{arg1, arg2}
	return n
}

func NewLenCallNode(arg *Node) *Node {
	n := newNodeBase(NodeLenCall)
	n.Arguments = []*Node{arg}
	return n
}

func NewStringCallNode(arg *Node) *Node {
	n := newNodeBase(NodeStringCall)
	n.Arguments = []*Node{arg}
	return n
}

func NewRuneCallNode(arg *Node) *Node {
	n := newNodeBase(NodeRuneCall)
	n.Arguments = []*Node{arg}
	return n
}
