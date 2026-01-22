package dcloud

import (
	"github.com/microsoft/typescript-go/internal/tspath"
	dis "github.com/microsoft/typescript-go/io/dcloud/disposable"
)

type ProjectCollection struct{
	// 打开的文件默认项目路径
	openFileDefaultProject map[tspath.Path /* openFilePath */]tspath.Path /* project path which is contain this file */
	// 所有已经创建的project
	projects []*dis.Box[*Project]
	// 所有已经创建的project路径映射map
	projectsByPath map[tspath.Path]*dis.Box[*Project]
	// 准备创建的项目
	// 如果打开或创建项目时，应将项目路径添加到此列表，待项目下文件打开时，创建项目
	// 如果关闭项目触发时，此列表中有相应的项目，需要将其移除
	// 如果项目下的文件打开后，需要使用此项目路径创建出项目，并将其从此列表中移除
	// 如果项目下打开的文件被全部关闭后，需要将项目移除，并将其放入到此列表中
	// 操作此列表时需要加锁
	preparedProjectPath []tspath.Path
}

func newProjectCollection() *ProjectCollection{
	return &ProjectCollection{
		openFileDefaultProject: make(map[tspath.Path]tspath.Path),
		projectsByPath: make(map[tspath.Path]*dis.Box[*Project]),
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
	if project, ok := p.projectsByPath[projectPath]; ok && project.Value() != nil{
		return project.Value()
	}

	return nil
}
