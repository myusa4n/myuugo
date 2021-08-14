package lang

type VariableKind string

const (
	VariableLocal    VariableKind = "VARIABLE LOCAL"
	VariableTopLevel VariableKind = "VARIABLE TOP LEVEL"
)

type Variable struct {
	Kind   VariableKind // トップレベル変数か、ローカル変数か
	Name   string       // 変数の名前
	Type   Type         // 変数の型
	Offset int          // RBPからのオフセット。
}

func NewTopLevelVariable(ty Type, name string) *Variable {
	return &Variable{Kind: VariableTopLevel, Type: ty, Name: name}
}

func NewLocalVariable(ty Type, name string) *Variable {
	return &Variable{Kind: VariableLocal, Type: ty, Name: name}
}
