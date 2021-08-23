package parse

import (
	"github.com/myuu222/myuugo/compiler/lang"
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

func (e *Environment) AddLocalVar(ty lang.Type, name string) *lang.Variable {
	lvar := e.FindLocalVar(name)
	if lvar != nil {
		return lvar
	}
	lvar = lang.NewLocalVariable(ty, name)
	fn := e.program.FindFunction(e.FunctionName)

	if fn == nil {
		panic("存在しない関数" + e.FunctionName + "の中でローカル変数を宣言しようとしています")
	}

	fn.LocalVariables = append(fn.LocalVariables, lvar)
	e.localVariables = append(e.localVariables, lvar)
	return lvar
}

func (e *Environment) FindLocalVar(name string) *lang.Variable {
	for _, lvar := range e.localVariables {
		if lvar.Name == name {
			return lvar
		}
	}
	return nil
}

func (e *Environment) FindVar(name string) *lang.Variable {
	var cur = e
	for cur != nil {
		lvar := cur.FindLocalVar(name)
		if lvar != nil {
			return lvar
		}
		cur = cur.parent
	}
	return e.program.FindTopLevelVariable(name)
}
