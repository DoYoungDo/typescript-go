package dcloud

import (
	"context"

	"github.com/microsoft/typescript-go/internal/collections"
	"github.com/microsoft/typescript-go/internal/compiler"
	"github.com/microsoft/typescript-go/internal/core"
	"github.com/microsoft/typescript-go/internal/lsp/lsproto"
	"github.com/microsoft/typescript-go/internal/project"
	"github.com/microsoft/typescript-go/internal/tspath"
	dis "github.com/microsoft/typescript-go/io/dcloud/disposable"
)

type ProjectCollectionBuilder struct{
	projectCollection *ProjectCollection

	openWorkspaceFolders map[lsproto.URI]string
	closeWorkspaceFolders map[lsproto.URI]string
	openFiles collections.Set[lsproto.DocumentUri]
	closeFiles collections.Set[lsproto.DocumentUri]
	changeFiles collections.Set[lsproto.DocumentUri]
}

func newProjectCollectionBuilder(projectCollection *ProjectCollection) *ProjectCollectionBuilder{
	return &ProjectCollectionBuilder{
		projectCollection: projectCollection,
		openWorkspaceFolders: make(map[lsproto.URI]string),
		closeWorkspaceFolders: make(map[lsproto.URI]string),
		openFiles: collections.Set[lsproto.DocumentUri]{},
		closeFiles: collections.Set[lsproto.DocumentUri]{},
		changeFiles: collections.Set[lsproto.DocumentUri]{},
	}
}

func (p *ProjectCollectionBuilder) Build(ctx context.Context, session *project.Session){
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

	projectFileChangedSummary := make(map[*Project]FileChangedSummary)
	getBaseProgram := func(uri lsproto.DocumentUri) *compiler.Program {
		// lsProject := snapshot.GetProjectsContainingFile(uri)[0]
		ls, err := session.GetLanguageService(ctx, uri)
		return core.IfElse(ls == nil || err != nil, nil,ls.GetProgram())
	}

	done:
	for uri := range p.openFiles.Keys(){
		for path, projectRef := range p.projectCollection.projects {
			if tspath.ContainsPath((string)(path), uri.FileName(), *compareOption){
				pro := projectRef.Value()
				if pro == nil{
					pro = NewProject(path, p.projectCollection.fs)
					pro.rootFiles = append(pro.rootFiles, uri.FileName())

					summary := FileChangedSummary{baseProgram: getBaseProgram(uri)}
					if sum, ok := projectFileChangedSummary[pro]; ok {
						summary = sum
					}
					summary.openedFiles.Add(uri.Path(compareOption.UseCaseSensitiveFileNames))
					projectFileChangedSummary[pro] = summary

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

				summary := FileChangedSummary{baseProgram: getBaseProgram(uri)}
				if sum, ok := projectFileChangedSummary[pro]; ok {
					summary = sum
				}
				summary.openedFiles.Add(uri.Path(compareOption.UseCaseSensitiveFileNames))
				projectFileChangedSummary[pro] = summary

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

	for uri := range p.closeFiles.Keys() {
		path := uri.Path(compareOption.UseCaseSensitiveFileNames)
		projectPath := p.projectCollection.openFileDefaultProject[path]
		projectRef := p.projectCollection.projects[projectPath]
		if pro := projectRef.Value(); pro != nil{
			for i, rootFile := range pro.rootFiles{
				if rootFile == uri.FileName(){
					pro.rootFiles = append(pro.rootFiles[:i], pro.rootFiles[i + 1:]...)

					summary := FileChangedSummary{baseProgram: getBaseProgram(uri)}
					if sum, ok := projectFileChangedSummary[pro]; ok {
						summary = sum
					}
					summary.closedFiles.Add(uri.Path(compareOption.UseCaseSensitiveFileNames))
					projectFileChangedSummary[pro] = summary

					break
				}
			}
		}

		delete(p.projectCollection.openFileDefaultProject, path)
	}

	for uri := range p.changeFiles.Keys(){
		path := uri.Path(compareOption.UseCaseSensitiveFileNames)
		projectPath := p.projectCollection.openFileDefaultProject[path]
		projectRef := p.projectCollection.projects[projectPath]
		if pro := projectRef.Value(); pro != nil{
			summary := FileChangedSummary{baseProgram: getBaseProgram(uri)}
			if sum, ok := projectFileChangedSummary[pro]; ok {
				summary = sum
			}
			summary.changedFiles.Add(uri.Path(compareOption.UseCaseSensitiveFileNames))
			projectFileChangedSummary[pro] = summary
		}
	}


	// 最后做整体更新
	for project, sum := range projectFileChangedSummary{
		// 需要将变量复制一份，方便go去捕获，否则可能出现循环中的go捕获同一个变量的问题
		projectCopy := project
		sumCopy := sum
		// 更新项目版本
		project.version++

		// 更新项目数据
		for _, ch := range project.programWatchedChannels{
			chCopy := ch
			project.programWatchedChannelsGroup.Add(1)
			go func ()  {
				chCopy <- sumCopy
			}()
		}

		// 更新channel
		go func(){
			projectCopy.programWatchedChannelsGroup.Wait()
			projectCopy.programWatchedChannelsMu.Lock()
			defer projectCopy.programWatchedChannelsMu.Unlock()

			for i := len(projectCopy.programWatchedChannels) - 1; i >= 0; i-- {
				select{
				case _, ok := <-projectCopy.programWatchedChannels[i]:
					if !ok {
						projectCopy.programWatchedChannels = append(projectCopy.programWatchedChannels[:i], projectCopy.programWatchedChannels[i + 1:]...)
					}
				default:
				}
			}
		}()
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
	if p.closeFiles.Has(uri){
		p.closeFiles.Delete(uri)
		return
	}
	p.openFiles.Add(uri)
}

func (p *ProjectCollectionBuilder) CloseFile(uri lsproto.DocumentUri){
	if p.openFiles.Has(uri) {
		p.openFiles.Delete(uri)
		return
	}
	p.closeFiles.Add(uri)
}

func (p *ProjectCollectionBuilder) ChangeFile(uri lsproto.DocumentUri){
	p.changeFiles.Add(uri)
}
