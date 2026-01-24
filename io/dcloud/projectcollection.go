package dcloud

import (
	"sync"

	"github.com/microsoft/typescript-go/internal/tspath"
	"github.com/microsoft/typescript-go/internal/vfs"
	dis "github.com/microsoft/typescript-go/io/dcloud/disposable"
)

type ProjectCollection struct{
	toPath func(string) tspath.Path
	fs vfs.FS
	cwd string

	// 打开的文件默认项目路径
	openFileDefaultProject map[tspath.Path /* openFilePath */]tspath.Path /* project path which is contain this file */
	// 所有已经创建的project路径映射map
	projects map[tspath.Path]*dis.Box[*Project]
	// 准备创建的项目
	// 如果打开或创建项目时，应将项目路径添加到此列表，待项目下文件打开时，创建项目
	// 如果关闭项目触发时，此列表中有相应的项目，需要将其移除
	// 如果项目下的文件打开后，需要使用此项目路径创建出项目，并将其从此列表中移除
	// 如果项目下打开的文件被全部关闭后，需要将项目移除，并将其放入到此列表中
	// 操作此列表时需要加锁
	preparedProjectPath map[tspath.Path /* prepared project path */]string /* prepared project name */
	
	// 操作上述数据时的锁
	locker sync.Mutex
}

func newProjectCollection(toPath func(string) tspath.Path, fs vfs.FS, cwd string) *ProjectCollection{
	return &ProjectCollection{
		toPath: toPath,
		fs: fs,
		cwd: cwd,
		openFileDefaultProject: make(map[tspath.Path]tspath.Path),
		projects: make(map[tspath.Path]*dis.Box[*Project]),
		preparedProjectPath :make(map[tspath.Path]string),
	}
}

func (p *ProjectCollection) Projects() []*Project{
	var r []*Project = make([]*Project, len(p.projects))
	for _, project := range p.projects{
		if project.Value() != nil{
			r = append(r, project.Value())
		}
	}
	return r
}

func (p *ProjectCollection) GetProjectByPath(projectPath tspath.Path) *Project{
	return p.projects[projectPath].Value()
}

func (p *ProjectCollection) GetProjectByFileName(fileName tspath.Path) *Project{
	if projectPath, ok := p.openFileDefaultProject[fileName]; ok{
		return p.GetProjectByPath(projectPath)
	}
	return nil
}