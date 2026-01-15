package dcloud

import (
	"github.com/microsoft/typescript-go/internal/ls"
)

type ProjectKind string
const (
	Web ProjectKind = "web"
)

type Project struct {
	server *Server

	kind ProjectKind
	fsPath string
}

func NewProject(fsPath string, server *Server) *Project {
	project := &Project{
		server: server,

		fsPath: fsPath,
	}
	project.init()

	return project
}

func (p *Project) init()  {
	// DO INIT
}

func (p *Project) Server() *Server {
	return p.server
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
