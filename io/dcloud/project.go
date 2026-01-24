package dcloud

import (
	"context"

	"github.com/microsoft/typescript-go/internal/ls"
	"github.com/microsoft/typescript-go/internal/lsp/lsproto"
	"github.com/microsoft/typescript-go/internal/tspath"
	"github.com/microsoft/typescript-go/internal/vfs"
	dis "github.com/microsoft/typescript-go/io/dcloud/disposable"
)
type ProjectKind string
const (
	Web ProjectKind = "web"
	UniApp ProjectKind = "uni-app"
)

type Project struct {
	kind ProjectKind
	fsPath string
	rootFiles []string

	plugins map[string]*dis.Box[Plugin]

	fs vfs.FS
}
var _ dis.Disposable = (*Project)(nil)

func NewProject(fsPath tspath.Path, fs vfs.FS) *Project{
	project := &Project{
		fsPath: string(fsPath),
		plugins: make(map[string]*dis.Box[Plugin]),
		fs: fs,
	}
	project.Init()
	return project
}

func (p *Project) Dispose() {
	for _, plugin := range p.plugins {
		plugin.Delete()
	}
	p.plugins = nil
}
func (p *Project) Init()  {
	// init kind 
	// TODO

	// init plugin
	creators := GetPluginCreators()
	for pluginId, creator := range creators {
		plugin, err := creator(p)
		if err != nil || plugin == nil {
			continue
		}
		p.plugins[pluginId] = dis.NewBox(plugin)
	}
}

func (p *Project) Kind() ProjectKind {
	return p.kind
}

func (p *Project) FsPath() string {
	return p.fsPath
}

func (p *Project) GetRootFiles() []string{
	rootFiles := make([]string, len(p.rootFiles))
	copy(rootFiles, p.rootFiles)
	return rootFiles
}

func (p *Project) GetLanguageService(ctx context.Context,defaultLs *ls.LanguageService, documentURI lsproto.DocumentUri) LanguageService {
	for _, pluginRef := range p.plugins{
		if plugin := pluginRef.Value(); plugin != nil {
			if pls := plugin.GetLanguageService(ctx, defaultLs, documentURI); pls != nil{
				// 更新autoimport数据
				host := pls.GetHost()
				host.UpdateAutoImport(ctx, documentURI)
				return pls
			}
		}
	}
	return nil
}

func (p *Project) GetPlugins() []Plugin{
	plugins := make([]Plugin, 0, len(p.plugins))
	for _, plugin := range p.plugins {
		if(plugin == nil){
			continue
		}
		p := plugin.Value()
		if p == nil{
			continue
		}
		plugins = append(plugins, p)
	}
	return plugins
}

func (p *Project) GetPlugin(pluginId string) Plugin{
	return p.plugins[pluginId].Value()
}

func (p *Project) ToPath(path string) tspath.Path{
	return tspath.ToPath(path, p.fsPath, p.fs.UseCaseSensitiveFileNames())
}

func (p *Project) FS() vfs.FS{
	return p.fs
}