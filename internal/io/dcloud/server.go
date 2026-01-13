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

func NewServer() *Server {	
	return &Server{}
}

func (s *Server) GetProject(ctx context.Context,fileName string) *Project {
	_, project, err , _:= s.session.GetLanguageServiceAndProjectsForFile(ctx, lsproto.DocumentUri(fileName))
	if err != nil {
		return nil
	}

	fsPath := project.GetProgram().GetCurrentDirectory()
	if  s.projects[fsPath] == nil {
		s.projects[fsPath] = NewProject(fsPath)
		return s.projects[fsPath]
	}
	return s.projects[fsPath]
}
