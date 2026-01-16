package dcloud

import (
	"context"

	"github.com/microsoft/typescript-go/internal/ls"
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

func (s *Server) GetProject(ctx context.Context,uri lsproto.DocumentUri) (*Project, error) {
	projects, err := s.GetProjectsForFile(ctx, uri)

	if err != nil || len(projects) == 0 {
		return nil, err
	}

	p := projects[0]
	fsPath := p.GetProgram().GetCurrentDirectory()

	if  s.projects[fsPath] == nil {
		s.projects[fsPath] = NewProject(fsPath, s)
		return s.projects[fsPath], nil
	}
	return s.projects[fsPath], nil
}

func (s *Server) GetProjectAndRootLanguageService(ctx context.Context,uri lsproto.DocumentUri) (*Project, LanguageService, error) {
	project, err := s.GetProject(ctx, uri)
	if err != nil {
		return nil, nil, err
	}

	return project, project.GetLanguageService(), nil
}

func (s *Server) DidOpenFile(ctx context.Context, uri lsproto.DocumentUri, version int32, content string, languageKind lsproto.LanguageKind) {

}

func (s *Server) DidCloseFile(ctx context.Context, uri lsproto.DocumentUri) {

}

func (s *Server) DidChangeFile(ctx context.Context, uri lsproto.DocumentUri, version int32, changes []lsproto.TextDocumentContentChangePartialOrWholeDocument) {

}

func (s *Server) DidSaveFile(ctx context.Context, uri lsproto.DocumentUri) {
	
}

// var _ ls.CrossProjectOrchestrator = (*Server)(nil) 

// func (s *Server) GetDefaultProject() ls.Project{
// 	s.session.GetLanguageServiceAndProjectsForFile()
// }

// func (s *Server)GetAllProjectsForInitialRequest() []ls.Project{

// }

// func (s *Server)GetLanguageServiceForProjectWithFile(ctx context.Context, project ls.Project, uri lsproto.DocumentUri) *ls.LanguageService {
	
// }

func (s *Server)GetProjectsForFile(ctx context.Context, uri lsproto.DocumentUri) ([]ls.Project, error){
	return s.session.GetProjectsForFile(ctx, uri)
}

func (s *Server)GetDefaultHost()ls.Host{
	snapShot, _ := s.session.Snapshot()
	return snapShot
}

// func (s *Server)GetProjectsLoadingProjectTree(ctx context.Context, requestedProjectTrees *collections.Set[tspath.Path]) iter.Seq[ls.Project]{

// }