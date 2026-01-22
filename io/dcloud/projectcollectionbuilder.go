package dcloud

import "github.com/microsoft/typescript-go/internal/lsp/lsproto"

type ProjectCollectionBuilder struct{
	projectCollection *ProjectCollection

	openWorkspaceFolders []*lsproto.WorkspaceFolder
	closeWorkspaceFolders []*lsproto.WorkspaceFolder
	openFiles []lsproto.DocumentUri
	closeFiles []lsproto.DocumentUri
}

func newProjectCollectionBuilder(projectCollection *ProjectCollection) *ProjectCollectionBuilder{
	return &ProjectCollectionBuilder{
		projectCollection: projectCollection,
	}
}

func (p *ProjectCollectionBuilder) Build(){
}

func (p *ProjectCollectionBuilder) OpenWorkspaceFolders(folders []*lsproto.WorkspaceFolder){
	p.openWorkspaceFolders = append(p.openWorkspaceFolders, folders...)
}

func (p *ProjectCollectionBuilder) CloseWorkspaceFolders(folders []*lsproto.WorkspaceFolder){
	p.closeWorkspaceFolders = append(p.closeWorkspaceFolders, folders...)
}

func (p *ProjectCollectionBuilder) OpenFile(uri lsproto.DocumentUri){
	p.openFiles = append(p.openFiles, uri)
}

func (p *ProjectCollectionBuilder) CloseFile(uri lsproto.DocumentUri){
	p.closeFiles = append(p.closeFiles, uri)
}
