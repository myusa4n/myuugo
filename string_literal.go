package main

type StringLiteral struct {
	label string
	value string
}

func NewStringLiteral(label string, value string) *StringLiteral {
	return &StringLiteral{label: label, value: value}
}
