package parse

import (
	"strconv"
	"strings"

	"github.com/myuu222/myuugo/compiler/lang"
)

type Program struct {
	TopLevelVariables []*lang.Variable
	Functions         []*lang.Function

	// そのうち削除するかも
	StringLiterals   []*lang.StringLiteral
	Code             []*Node
	UserDefinedTypes []lang.Type
	PackagesToImport []string
}

func NewProgram() *Program {
	return &Program{
		TopLevelVariables: []*lang.Variable{},
		Functions:         []*lang.Function{},
		StringLiterals:    []*lang.StringLiteral{},
		Code:              []*Node{},
		PackagesToImport:  []string{},
	}
}

func (p *Program) AddTopLevelVariable(ty lang.Type, name string) *lang.Variable {
	if p.FindTopLevelVariable(name) != nil {
		return p.FindTopLevelVariable(name)
	}
	var newVar = lang.NewTopLevelVariable(ty, name)
	p.TopLevelVariables = append(p.TopLevelVariables, newVar)
	return newVar
}

func (p *Program) FindTopLevelVariable(name string) *lang.Variable {
	for _, v := range p.TopLevelVariables {
		if v.Name == name {
			return v
		}
	}
	return nil
}

func (p *Program) AddPackageToImport(name string) string {
	pkg, ok := p.FindPackageToImport(name)
	if ok {
		return pkg
	}
	p.PackagesToImport = append(p.PackagesToImport, name)
	return name
}

func (p *Program) FindPackageToImport(name string) (string, bool) {
	for _, pkg := range p.PackagesToImport {
		sections := strings.Split(strings.Trim(pkg, "\""), "/")
		if sections[len(sections)-1] == name {
			return pkg, true
		}
	}
	return "", false
}

func (p *Program) RegisterFunction(fn *lang.Function) {
	if p.FindFunction(fn.Label) != nil {
		panic("関数" + fn.Label + "は既に存在しています")
	}
	p.Functions = append(p.Functions, fn)
}

func (p *Program) FindFunction(name string) *lang.Function {
	for _, f := range p.Functions {
		if f.Label == name {
			return f
		}
	}
	return nil
}

func (p *Program) RegisterType(udt lang.Type) {
	_, ok := p.FindType(udt.DefinedName)
	if ok {
		panic("型" + udt.DefinedName + "は既に定義されています")
	}

	// 元が構造体だった場合はオフセットを修正
	if udt.PtrTo.Kind == lang.TypeStruct {
		entityType := udt.PtrTo
		for i := 1; i < len(entityType.MemberNames); i++ {
			entityType.MemberOffsets[i] = entityType.MemberOffsets[i-1] + lang.Sizeof(entityType.MemberTypes[i-1])
		}
	}
	p.UserDefinedTypes = append(p.UserDefinedTypes, udt)
}

func (p *Program) FindType(name string) (lang.Type, bool) {
	for _, t := range p.UserDefinedTypes {
		if t.DefinedName == name {
			return t, true
		}
	}
	return lang.Type{}, false
}

func (p *Program) AddStringLiteral(value string) *lang.StringLiteral {
	var label = ".LStr" + strconv.Itoa(len(p.StringLiterals))
	var str = lang.NewStringLiteral(label, value)
	p.StringLiterals = append(p.StringLiterals, str)
	return str
}
