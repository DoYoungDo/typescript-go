package dcloud

import (
	"context"

	"github.com/microsoft/typescript-go/internal/compiler"
	"github.com/microsoft/typescript-go/internal/ls"
	"github.com/microsoft/typescript-go/internal/lsp/lsproto"
	"github.com/microsoft/typescript-go/internal/project"
	dis "github.com/microsoft/typescript-go/io/dcloud/disposable"
)

type Server struct {
	session *project.Session

	// 缓存所有项目
	projects map[string]*dis.Box[*Project]
}

func NewServer(session *project.Session) *Server {	
	return &Server{
		session: session,
		projects: make(map[string]*dis.Box[*Project]),
	}
}

func (s *Server) GetProject(ctx context.Context,uri lsproto.DocumentUri) (*Project, error) {
	p, _ := s.GetDefaultProjectAndSnapShot(uri)
	fsPath := p.GetProgram().GetCurrentDirectory()
	configFilePath := p.ConfigFilePath()

	// 当tsgo默认获取到了项目，此处恒创建
	if entry := s.projects[fsPath]; entry == nil || entry.Value() == nil {
		s.projects[fsPath] = dis.NewBox(NewProject(s, configFilePath, fsPath, func(program *compiler.Program, host ls.Host)*ls.LanguageService{
			return ls.NewLanguageService(configFilePath, program, host)
		}))

		return s.projects[fsPath].Value(), nil
	}

	return s.projects[fsPath].Value(), nil
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
	// 当文件时闭，需要计算是否需要清理项目
}

func (s *Server) DidChangeFile(ctx context.Context, uri lsproto.DocumentUri, version int32, changes []lsproto.TextDocumentContentChangePartialOrWholeDocument) {

}

func (s *Server) DidSaveFile(ctx context.Context, uri lsproto.DocumentUri) {

}

func (s *Server) GetDefaultProjectAndSnapShot(uri lsproto.DocumentUri)(*project.Project, *project.Snapshot){
	snapShot, _ := s.session.Snapshot()
	project := snapShot.GetDefaultProject(uri)
	return project, snapShot
}
