package lang

type StringLiteral struct {
	Label string
	Value string
}

func NewStringLiteral(label string, value string) *StringLiteral {
	return &StringLiteral{Label: label, Value: value}
}
