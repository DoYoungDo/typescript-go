package dcloud

import (
	"context"

	"github.com/microsoft/typescript-go/internal/lsp/lsproto"
	"github.com/microsoft/typescript-go/internal/project"
)

type Server struct {
	session *project.Session

	// 缓存所有项目
	projects map[string]*Project
}

func NewServer(session *project.Session) *Server {	
	return &Server{
		session: session,
		projects: make(map[string]*Project),
	}
}

func (s *Server) GetProject(ctx context.Context,uri lsproto.DocumentUri) *Project {
	project, _, _, err := s.session.GetLanguageServiceAndProjectsForFile(ctx, uri)
	if err != nil {
		return nil
	}

	fsPath := project.GetProgram().GetCurrentDirectory()
	if  s.projects[fsPath] == nil {
		s.projects[fsPath] = NewProject(fsPath, s.session)
		return s.projects[fsPath]
	}
	return s.projects[fsPath]
}
