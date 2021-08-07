package main

type Environment struct {
	TopLevelVars  []*Variable
	LocalVarTable map[string][]*Variable
	FunctionTable map[string]*Function
}

func NewEnvironment() *Environment {
	return &Environment{
		TopLevelVars:  []*Variable{},
		LocalVarTable: map[string][]*Variable{},
		FunctionTable: map[string]*Function{},
	}
}

func (e *Environment) RegisterFunc(label string) *Function {
	_, ok := e.FunctionTable[label]
	if ok {
		madden("関数%sは既に存在しています", label)
	}
	var fn = NewFunction(label, []Type{}, NewType(TypeUndefined))
	e.FunctionTable[label] = fn
	e.LocalVarTable[label] = []*Variable{}
	return fn
}

func (e *Environment) AddLocalVar(fnLabel string, token Token) *Variable {
	lvar := e.FindLocalVar(fnLabel, token)
	if lvar != nil {
		return lvar
	}
	var locals = e.LocalVarTable[fnLabel]
	lvar = &Variable{name: token.str, varType: Type{kind: TypeUndefined}, kind: VariableLocal}
	if len(locals) == 0 {
		lvar.offset = 0 + 8
	} else {
		lvar.offset = locals[len(locals)-1].offset + 8
	}
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
	for _, tvar := range e.TopLevelVars {
		if tvar.name == token.str {
			return tvar
		}
	}
	return nil
}

func (e *Environment) AddTopLevelVar(token Token) *Variable {
	tvar := e.FindTopLevelVar(token)
	if tvar != nil {
		return tvar
	}
	tvar = &Variable{name: token.str, kind: VariableTopLevel, varType: Type{kind: TypeUndefined}}
	e.TopLevelVars = append(e.TopLevelVars, tvar)
	return tvar
}

func (e *Environment) FindVar(fnLabel string, token Token) *Variable {
	_, ok := e.FunctionTable[fnLabel]
	if ok {
		lvar := e.FindLocalVar(fnLabel, token)
		if lvar != nil {
			return lvar
		}
	}
	for _, tvar := range e.TopLevelVars {
		if tvar.name == token.str {
			return tvar
		}
	}
	return nil
}

func (e *Environment) GetFrameSize(fnLabel string) int {
	locals, ok := e.LocalVarTable[fnLabel]
	if !ok {
		madden("関数%sは存在しません", fnLabel)
	}
	var size int = 0
	for _, lvar := range locals {
		size += sizeof(lvar.varType.kind)
	}
	return size
}
