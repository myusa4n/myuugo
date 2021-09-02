package parse

import "strings"

type Source struct {
	FileName string
	Packages []string
	Code     []*Node
}

func NewSource(fileName string) *Source {
	return &Source{FileName: fileName, Packages: []string{}, Code: []*Node{}}
}

func (s *Source) AddPackage(name string) string {
	pkg, ok := s.FindPackage(name)
	if ok {
		return pkg
	}
	s.Packages = append(s.Packages, name)
	return name
}

func (s *Source) FindPackage(name string) (string, bool) {
	for _, pkg := range s.Packages {
		sections := strings.Split(strings.Trim(pkg, "\""), "/")
		if sections[len(sections)-1] == name {
			return pkg, true
		}
	}
	return "", false
}
