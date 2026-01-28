package dcloud

import (
	"context"
	"time"

	"github.com/microsoft/typescript-go/internal/ls"
	"github.com/microsoft/typescript-go/internal/lsp/lsproto"
	"github.com/microsoft/typescript-go/internal/project"
	"github.com/microsoft/typescript-go/internal/tspath"
	"github.com/microsoft/typescript-go/internal/vfs"
)
type ServerOption struct{
	Cwd string
	Fs vfs.FS
}

type Server struct {
	session *project.Session

	cwd string
	fs vfs.FS

	// 项目管理器
	projectCollection *ProjectCollection
}

func NewServer(opt *ServerOption, session *project.Session) *Server {	
	server := &Server{
		session: session,
		cwd: opt.Cwd,
		fs: opt.Fs,
		// projects: make(map[string]*dis.Box[*Project]),
	}

	useCase := server.fs.UseCaseSensitiveFileNames()
	server.projectCollection = newProjectCollection(func(s string) tspath.Path {
		return  tspath.ToPath(s, opt.Cwd, useCase)
	},server.fs, server.cwd)

	return server
}

func (s *Server) HandleInitialized(ctx context.Context, params *lsproto.InitializeParams){
	if params.WorkspaceFolders != nil && params.WorkspaceFolders.WorkspaceFolders != nil {
		builder := newProjectCollectionBuilder(s.projectCollection)
		builder.OpenWorkspaceFolders(*params.WorkspaceFolders.WorkspaceFolders)
		builder.Build(ctx, s.session)
	}
}

func (s *Server) HandleDidChangeWorkspaceFolders(ctx context.Context, params *lsproto.DidChangeWorkspaceFoldersParams){
	if params.Event != nil{
		builder := newProjectCollectionBuilder(s.projectCollection)
		builder.OpenWorkspaceFolders(params.Event.Added)
		builder.CloseWorkspaceFolders(params.Event.Removed)
		builder.Build(ctx, s.session)
	}
}

func (s *Server) DidOpenFile(ctx context.Context, uri lsproto.DocumentUri, version int32, content string, languageKind lsproto.LanguageKind) {
	builder := newProjectCollectionBuilder(s.projectCollection)
	builder.OpenFile(uri)
	builder.Build(ctx, s.session)
}

func (s *Server) DidCloseFile(ctx context.Context, uri lsproto.DocumentUri) {
	builder := newProjectCollectionBuilder(s.projectCollection)
	builder.CloseFile(uri)
	builder.Build(ctx, s.session)
}

func (s *Server) DidChangeFile(ctx context.Context, uri lsproto.DocumentUri, version int32, changes []lsproto.TextDocumentContentChangePartialOrWholeDocument) {
	builder := newProjectCollectionBuilder(s.projectCollection)
	builder.ChangeFile(uri)
	builder.Build(ctx, s.session)
}

func (s *Server) DidSaveFile(ctx context.Context, uri lsproto.DocumentUri) {

}

func (s *Server) HandleCompletion(ctx context.Context, languageService *ls.LanguageService, params *lsproto.CompletionParams) (lsproto.CompletionResponse, error, bool) {
	start := time.Now()
	defer func ()  {
		println("HandleCompletion spend", time.Since(start).Milliseconds() , `\n`)
	}()

	project := s.projectCollection.GetProjectByFileName(params.TextDocumentURI().Path(s.fs.UseCaseSensitiveFileNames()))
	println("HandleCompletion- get project spend", time.Since(start).Milliseconds() , `\n`)
	tmpStart := time.Now()
	if ls := project.GetLanguageService(ctx, languageService, params.TextDocument.Uri); ls != nil{
		println("HandleCompletion- get ls spend", time.Since(tmpStart).Milliseconds() , `\n`)
		tmpStart = time.Now()
		if fn := ls.GetProvideCompletion(); fn != nil{
			println("HandleCompletion- get ls call spend", time.Since(tmpStart).Milliseconds() , `\n`)
			tmpStart = time.Now()
			res, err := fn(ctx, params.TextDocument.Uri, params.Position, params.Context)
			println("HandleCompletion- ls call spend", time.Since(tmpStart).Milliseconds() , `\n`)
			return res, err, true
		}
	}
	return lsproto.CompletionResponse{}, nil, false
}