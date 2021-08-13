package main

import "strconv"

type Environment struct {
	program        *Program
	StringLiterals []*StringLiteral
	LocalVarTable  map[string][]*Variable

	stringLabelNumber int
}

func NewEnvironment() *Environment {
	return &Environment{
		program:        NewProgram(),
		StringLiterals: []*StringLiteral{},
		LocalVarTable:  map[string][]*Variable{},
	}
}

func (e *Environment) RegisterFunc(label string) *Function {
	if e.program.FindFunction(label) != nil {
		madden("関数%sは既に存在しています", label)
	}
	var fn = NewFunction(label, []Type{}, NewType(TypeUndefined))
	e.program.RegisterFunction(fn)
	e.LocalVarTable[label] = []*Variable{}
	return fn
}

func (e *Environment) AddLocalVar(fnLabel string, token Token) *Variable {
	lvar := e.FindLocalVar(fnLabel, token)
	if lvar != nil {
		return lvar
	}
	lvar = &Variable{name: token.str, varType: Type{kind: TypeUndefined}, kind: VariableLocal}
	e.LocalVarTable[fnLabel] = append(e.LocalVarTable[fnLabel], lvar)
	return lvar
}

func (e *Environment) FindLocalVar(fnLabel string, token Token) *Variable {
	locals, ok := e.LocalVarTable[fnLabel]

	if !ok {
		madden("関数%sは存在しません", fnLabel)
	}
	for _, lvar := range locals {
		if lvar.name == token.str {
			return lvar
		}
	}
	return nil
}

func (e *Environment) FindTopLevelVar(token Token) *Variable {
	return e.program.FindTopLevelVariable(token.str)
}

func (e *Environment) AddTopLevelVar(token Token) *Variable {
	return e.program.AddTopLevelVariable(NewType(TypeUndefined), token.str)
}

func (e *Environment) FindVar(fnLabel string, token Token) *Variable {
	ok := e.program.FindFunction(fnLabel) != nil
	if ok {
		lvar := e.FindLocalVar(fnLabel, token)
		if lvar != nil {
			return lvar
		}
	}
	return e.program.FindTopLevelVariable(token.str)
}

func (e *Environment) GetFrameSize(fnLabel string) int {
	locals, ok := e.LocalVarTable[fnLabel]
	if !ok {
		madden("関数%sは存在しません", fnLabel)
	}
	var size int = 0
	for _, lvar := range locals {
		size += Sizeof(lvar.varType)
	}
	return size
}

func (e *Environment) AlignLocalVars(fnLabel string) {
	locals, ok := e.LocalVarTable[fnLabel]
	if !ok {
		madden("関数%sは存在しません", fnLabel)
	}
	var offset = 0
	for _, lvar := range locals {
		offset += Sizeof(lvar.varType)
		lvar.offset = offset
	}
}

func (e *Environment) AddStringLiteral(token Token) *StringLiteral {
	var label = ".LStr" + strconv.Itoa(e.stringLabelNumber)
	e.stringLabelNumber += 1
	var str = &StringLiteral{label: label, value: token.str}
	e.StringLiterals = append(e.StringLiterals, str)
	return str
}
