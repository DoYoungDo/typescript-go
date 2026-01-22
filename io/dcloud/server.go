package dcloud

import (
	"context"

	"github.com/microsoft/typescript-go/internal/compiler"
	"github.com/microsoft/typescript-go/internal/ls"
	"github.com/microsoft/typescript-go/internal/lsp/lsproto"
	"github.com/microsoft/typescript-go/internal/project"
	"github.com/microsoft/typescript-go/internal/tspath"
	dis "github.com/microsoft/typescript-go/io/dcloud/disposable"
)

type Server struct {
	session *project.Session

	// 项目管理器
	projectCollection *ProjectCollection
	// 缓存所有项目
	projects map[string]*dis.Box[*Project]
}

func NewServer(session *project.Session) *Server {	
	return &Server{
		session: session,
		projectCollection: newProjectCollection(),
		projects: make(map[string]*dis.Box[*Project]),
	}
}

func (s *Server) getProject(_ context.Context, uri lsproto.DocumentUri) (*Project, error) {
	p, _ := s.GetDefaultProjectAndSnapShot(uri)
	fsPath := p.GetProgram().GetCurrentDirectory()
	configFilePath := tspath.ToPath(fsPath, fsPath, p.GetProgram().Host().FS().UseCaseSensitiveFileNames())
	if p.Kind == project.KindConfigured{
		configFilePath = p.ConfigFilePath()
	}

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
	project, err := s.getProject(ctx, uri)
	if err != nil {
		return nil, nil, err
	}

	return project, project.GetLanguageService(), nil
}

func (s *Server) HandleInitialize(_ context.Context, params *lsproto.InitializeParams){
	if params.WorkspaceFolders != nil && params.WorkspaceFolders.WorkspaceFolders != nil {
		builder := newProjectCollectionBuilder(s.projectCollection)
		builder.OpenWorkspaceFolders(*params.WorkspaceFolders.WorkspaceFolders)
		builder.Build()
	}
}

func (s *Server) HandleDidChangeWorkspaceFolders(_ context.Context, params *lsproto.DidChangeWorkspaceFoldersParams){
	if params.Event != nil{
		builder := newProjectCollectionBuilder(s.projectCollection)
		builder.OpenWorkspaceFolders(params.Event.Added)
		builder.CloseWorkspaceFolders(params.Event.Removed)
		builder.Build()
	}
}

func (s *Server) DidOpenFile(ctx context.Context, uri lsproto.DocumentUri, version int32, content string, languageKind lsproto.LanguageKind) {
	builder := newProjectCollectionBuilder(s.projectCollection)
	builder.OpenFile(uri)
	builder.Build()
}

func (s *Server) DidCloseFile(ctx context.Context, uri lsproto.DocumentUri) {
	builder := newProjectCollectionBuilder(s.projectCollection)
	builder.CloseFile(uri)
	builder.Build()
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
