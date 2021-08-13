package main

import "strconv"

type Program struct {
	topLevelVariables []*Variable
	functions         []*Function

	StringLiterals []*StringLiteral
}

func NewProgram() *Program {
	return &Program{topLevelVariables: []*Variable{}, functions: []*Function{}, StringLiterals: []*StringLiteral{}}
}

func (p *Program) AddTopLevelVariable(ty Type, name string) *Variable {
	if p.FindTopLevelVariable(name) != nil {
		return p.FindTopLevelVariable(name)
	}
	var newVar = NewTopLevelVariable(ty, name)
	p.topLevelVariables = append(p.topLevelVariables, newVar)
	return newVar
}

func (p *Program) FindTopLevelVariable(name string) *Variable {
	for _, v := range p.topLevelVariables {
		if v.name == name {
			return v
		}
	}
	return nil
}

func (p *Program) RegisterFunction(fn *Function) {
	if p.FindFunction(fn.Label) != nil {
		panic("関数" + fn.Label + "は既に存在しています")
	}
	p.functions = append(p.functions, fn)
}

func (p *Program) FindFunction(name string) *Function {
	for _, f := range p.functions {
		if f.Label == name {
			return f
		}
	}
	return nil
}

func (p *Program) AddStringLiteral(value string) *StringLiteral {
	var label = ".LStr" + strconv.Itoa(len(p.StringLiterals))
	var str = NewStringLiteral(label, value)
	p.StringLiterals = append(p.StringLiterals, str)
	return str
}
