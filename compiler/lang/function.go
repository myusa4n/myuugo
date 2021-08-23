package lang

type Function struct {
	Label           string
	ParameterTypes  []Type
	ReturnValueType Type
	LocalVariables  []*Variable
}

func NewFunction(label string, parameterTypes []Type, returnValueType Type) *Function {
	return &Function{Label: label, ParameterTypes: parameterTypes, ReturnValueType: returnValueType, LocalVariables: []*Variable{}}
}
