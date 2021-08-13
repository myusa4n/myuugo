package main

type Environment struct {
	program *Program
	parent  *Environment

	localVariables []*Variable
}

func NewEnvironment() *Environment {
	return &Environment{
		program: NewProgram(),
	}
}

func (e *Environment) Fork() *Environment {
	var newE = NewEnvironment()
	newE.parent = e
	newE.program = e.program
	return newE
}

func (e *Environment) RegisterFunc(label string) *Function {
	if e.program.FindFunction(label) != nil {
		madden("関数%sは既に存在しています", label)
	}
	var fn = NewFunction(label, []Type{}, NewType(TypeUndefined))
	e.program.RegisterFunction(fn)
	return fn
}

func (e *Environment) AddLocalVar(fnLabel string, token Token) *Variable {
	lvar := e.FindLocalVar(fnLabel, token)
	if lvar != nil {
		return lvar
	}
	lvar = NewLocalVariable(NewType(TypeUndefined), token.str)
	fn := e.program.FindFunction(fnLabel)

	if fn == nil {
		panic("存在しない関数" + fnLabel + "の中でローカル変数を宣言しようとしています")
	}

	fn.LocalVariables = append(fn.LocalVariables, lvar)
	e.localVariables = append(e.localVariables, lvar)
	return lvar
}

func (e *Environment) FindLocalVar(fnLabel string, token Token) *Variable {
	for _, lvar := range e.localVariables {
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
	var cur = e
	for cur != nil {
		lvar := cur.FindLocalVar(fnLabel, token)
		if lvar != nil {
			return lvar
		}
		cur = cur.parent
	}
	return e.program.FindTopLevelVariable(token.str)
}

func (e *Environment) GetFrameSize(fnLabel string) int {
	fn := e.program.FindFunction(fnLabel)
	if fn == nil {
		madden("関数%sは存在しません", fnLabel)
	}
	var size int = 0
	for _, lvar := range fn.LocalVariables {
		size += Sizeof(lvar.varType)
	}
	return size
}

func (e *Environment) AlignLocalVars(fnLabel string) {
	fn := e.program.FindFunction(fnLabel)
	if fn == nil {
		madden("関数%sは存在しません", fnLabel)
	}
	var offset = 0
	for _, lvar := range fn.LocalVariables {
		offset += Sizeof(lvar.varType)
		lvar.offset = offset
	}
}

func (e *Environment) AddStringLiteral(token Token) *StringLiteral {
	return e.program.AddStringLiteral(token.str)
}
