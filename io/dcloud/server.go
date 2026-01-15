package dcloud

import (
	"context"

	"github.com/microsoft/typescript-go/internal/ls"
	"github.com/microsoft/typescript-go/internal/lsp/lsproto"
)

type Server struct {
	ls.CrossProjectOrchestrator
	// 缓存所有项目
	projects map[string]*Project
}

func NewServer(c ls.CrossProjectOrchestrator) *Server {	
	return &Server{
		CrossProjectOrchestrator: c,
		projects: make(map[string]*Project),
	}
}

func (s *Server) GetProject(ctx context.Context,uri lsproto.DocumentUri) *Project {
	// project, _, _, err := s.session.GetLanguageServiceAndProjectsForFile(ctx, uri)
	// if err != nil {
	// 	return nil
	// }

	projects, err := s.GetProjectsForFile(ctx, uri)

	if err != nil || len(projects) == 0 {
		return nil
	}

	p := projects[0]
	fsPath := p.GetProgram().GetCurrentDirectory()

	if  s.projects[fsPath] == nil {
		s.projects[fsPath] = NewProject(fsPath, s)
		return s.projects[fsPath]
	}
	return s.projects[fsPath]
}
