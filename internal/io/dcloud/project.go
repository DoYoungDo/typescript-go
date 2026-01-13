package dcloud

import (
	"github.com/microsoft/typescript-go/internal/ls"
	"github.com/microsoft/typescript-go/internal/project"
)

type ProjectKind string
const (
	Web ProjectKind = "web"
)

type Project struct {
	kind ProjectKind
	fsPath string

	session *project.Session
}

func NewProject(fsPath string, session *project.Session) *Project {
	project := &Project{
		fsPath: fsPath,
		session: session,
	}
	project.init()

	return project
}

func (p *Project) init()  {
	// DO INIT
}

func (p *Project) Kind() ProjectKind {
	return p.kind
}

func (s *Project) GetLanguageService(defaultLs *ls.LanguageService, fileName string) *LanguageService {
	return &LanguageService{
		LanguageService: defaultLs,
		session: s.session,
		project: s,
	}
}
