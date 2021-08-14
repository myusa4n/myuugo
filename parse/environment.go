package parse

import (
	"github.com/myuu222/myuugo/lang"
	"github.com/myuu222/myuugo/util"
)

type Environment struct {
	program        *Program
	parent         *Environment
	localVariables []*lang.Variable

	FunctionName string
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
	newE.FunctionName = e.FunctionName
	return newE
}

func (e *Environment) RegisterFunc(label string) *lang.Function {
	if e.program.FindFunction(label) != nil {
		util.Alarm("関数%sは既に存在しています", label)
	}
	var fn = lang.NewFunction(label, []lang.Type{}, lang.NewType(lang.TypeUndefined))
	e.program.RegisterFunction(fn)
	return fn
}

func (e *Environment) AddLocalVar(fnLabel string, token Token) *lang.Variable {
	lvar := e.FindLocalVar(fnLabel, token)
	if lvar != nil {
		return lvar
	}
	lvar = lang.NewLocalVariable(lang.NewType(lang.TypeUndefined), token.str)
	fn := e.program.FindFunction(fnLabel)

	if fn == nil {
		panic("存在しない関数" + fnLabel + "の中でローカル変数を宣言しようとしています")
	}

	fn.LocalVariables = append(fn.LocalVariables, lvar)
	e.localVariables = append(e.localVariables, lvar)
	return lvar
}

func (e *Environment) FindLocalVar(fnLabel string, token Token) *lang.Variable {
	for _, lvar := range e.localVariables {
		if lvar.Name == token.str {
			return lvar
		}
	}
	return nil
}

func (e *Environment) FindTopLevelVar(token Token) *lang.Variable {
	return e.program.FindTopLevelVariable(token.str)
}

func (e *Environment) AddTopLevelVar(token Token) *lang.Variable {
	return e.program.AddTopLevelVariable(lang.NewType(lang.TypeUndefined), token.str)
}

func (e *Environment) FindVar(fnLabel string, token Token) *lang.Variable {
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
		panic("関数 \"" + fnLabel + " は存在しません")
	}
	var size int = 0
	for _, lvar := range fn.LocalVariables {
		size += lang.Sizeof(lvar.Type)
	}
	return size
}

func (e *Environment) AlignLocalVars(fnLabel string) {
	fn := e.program.FindFunction(fnLabel)
	if fn == nil {
		panic("関数 \"" + fnLabel + " は存在しません")
	}
	var offset = 0
	for _, lvar := range fn.LocalVariables {
		offset += lang.Sizeof(lvar.Type)
		lvar.Offset = offset
	}
}

func (e *Environment) AddStringLiteral(token Token) *lang.StringLiteral {
	return e.program.AddStringLiteral(token.str)
}

func (e *Environment) FindFunction(name string) *lang.Function {
	return e.program.FindFunction(name)
}
