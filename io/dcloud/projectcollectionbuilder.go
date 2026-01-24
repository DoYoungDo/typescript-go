package dcloud

import (
	"github.com/microsoft/typescript-go/internal/lsp/lsproto"
	"github.com/microsoft/typescript-go/internal/tspath"
	dis "github.com/microsoft/typescript-go/io/dcloud/disposable"
)

type ProjectCollectionBuilder struct{
	projectCollection *ProjectCollection

	openWorkspaceFolders map[lsproto.URI]string
	closeWorkspaceFolders map[lsproto.URI]string
	openFiles map[lsproto.DocumentUri]bool
	closeFiles map[lsproto.DocumentUri]bool
}

func newProjectCollectionBuilder(projectCollection *ProjectCollection) *ProjectCollectionBuilder{
	return &ProjectCollectionBuilder{
		projectCollection: projectCollection,
		openWorkspaceFolders: make(map[lsproto.URI]string),
		closeWorkspaceFolders: make(map[lsproto.URI]string),
		openFiles: make(map[lsproto.DocumentUri]bool),
		closeFiles: make(map[lsproto.DocumentUri]bool),
	}
}

func (p *ProjectCollectionBuilder) Build(){
	p.projectCollection.locker.Lock()
	defer p.projectCollection.locker.Unlock()

	compareOption := &tspath.ComparePathsOptions{
		UseCaseSensitiveFileNames: p.projectCollection.fs.UseCaseSensitiveFileNames(),
		CurrentDirectory: p.projectCollection.cwd,
	}

	for uri, name := range p.openWorkspaceFolders {
		if uri == ""{
			continue
		}

		path := lsproto.DocumentUri(uri).Path(compareOption.UseCaseSensitiveFileNames)
		if _, ok := p.projectCollection.preparedProjectPath[path]; !ok{
			p.projectCollection.preparedProjectPath[path] = name
		}
	}

	done:
	for uri := range p.openFiles{
		for path, projectRef := range p.projectCollection.projects {
			if tspath.ContainsPath((string)(path), uri.FileName(), *compareOption){
				pro := projectRef.Value()
				if pro == nil{
					pro = NewProject(path, p.projectCollection.fs)
					pro.rootFiles = append(pro.rootFiles, uri.FileName())

					p.projectCollection.projects[path] = dis.NewBox(pro)
				}
				// 如果已经存在项目，将当前文件与该项目关联
				p.projectCollection.openFileDefaultProject[uri.Path(compareOption.UseCaseSensitiveFileNames)] = path
				continue done
			}
		}

		for path := range p.projectCollection.preparedProjectPath {
			if tspath.ContainsPath((string)(path), uri.FileName(), *compareOption){
				pro := NewProject(path, p.projectCollection.fs)
				pro.rootFiles = append(pro.rootFiles, uri.FileName())

				// 如果项目还未打开，创建一个项目
				p.projectCollection.projects[path] = dis.NewBox(pro)
				// 再将当前文件与该项目关联
				p.projectCollection.openFileDefaultProject[uri.Path(compareOption.UseCaseSensitiveFileNames)] = path
				// 从准备容器中移除待创建项目
				delete(p.projectCollection.preparedProjectPath, path)
				continue done
			}
		}
	}

	for uri := range p.closeFiles {
		path := uri.Path(compareOption.UseCaseSensitiveFileNames)
		projectPath := p.projectCollection.openFileDefaultProject[path]
		projectRef := p.projectCollection.projects[projectPath]
		if pro := projectRef.Value(); pro != nil{
			for i, rootFile := range pro.rootFiles{
				if rootFile == uri.FileName(){
					pro.rootFiles = append(pro.rootFiles[:i], pro.rootFiles[i + 1:]...)
					break
				}
			}
		}

		delete(p.projectCollection.openFileDefaultProject, path)
	}

	// TODO clean project if project has no root files
}

func (p *ProjectCollectionBuilder) OpenWorkspaceFolders(folders []*lsproto.WorkspaceFolder){
	for _, folder := range folders{
		if _, ok := p.closeWorkspaceFolders[folder.Uri]; ok {
			delete(p.closeWorkspaceFolders, folder.Uri)
			continue
		}
		p.openWorkspaceFolders[folder.Uri] = folder.Name
	}
}

func (p *ProjectCollectionBuilder) CloseWorkspaceFolders(folders []*lsproto.WorkspaceFolder){
	for _, folder := range folders{
		if _, ok := p.openWorkspaceFolders[folder.Uri]; ok {
			delete(p.openWorkspaceFolders, folder.Uri)
			continue
		}
		p.closeWorkspaceFolders[folder.Uri] = folder.Name
	}
}

func (p *ProjectCollectionBuilder) OpenFile(uri lsproto.DocumentUri){
	// 如果待关闭的文件中有打开的文件，则将其从待关闭的文件中删除,也不需要再重新打开了
	if _, ok := p.closeFiles[uri]; ok {
		delete(p.closeFiles, uri)
		return
	}
	p.openFiles[uri] = true
}

func (p *ProjectCollectionBuilder) CloseFile(uri lsproto.DocumentUri){
	if _, ok := p.openFiles[uri]; ok {
		delete(p.openFiles, uri)
		return
	}
	p.closeFiles[uri] = true
}
