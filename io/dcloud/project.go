package dcloud

import (
	"context"
	"sync"

	"github.com/microsoft/typescript-go/internal/compiler"
	"github.com/microsoft/typescript-go/internal/ls"
	"github.com/microsoft/typescript-go/internal/lsp/lsproto"
	dis "github.com/microsoft/typescript-go/io/dcloud/disposable"
)

type PluginLanguageService interface{
	LanguageService
	IsEnable(fileName lsproto.DocumentUri)bool
}
type Plugin interface{
	dis.Disposable
	GetLanguageService(defaultLs *ls.LanguageService) PluginLanguageService
}

type PluginCreator func(project *Project) (Plugin, error)
var pluginHandle = sync.OnceValues(func()(func(string, PluginCreator), func() map[string]PluginCreator){
	creators := make(map[string]PluginCreator)
	var mu sync.Mutex

	register := func(id string, creator PluginCreator){
		mu.Lock()
		defer mu.Unlock()
		creators[id] = creator
	}

	get := func() map[string]PluginCreator{
		mu.Lock()
		defer mu.Unlock()
		return creators
	}

	return register, get
})
func InstallPluginCreator(pluginId string, creator PluginCreator) {
	register, _ := pluginHandle()
	register(pluginId, creator)
}
func GetPluginCreators() map[string]PluginCreator {
	_, get := pluginHandle()
	return get()
}

type ProjectKind string
const (
	Web ProjectKind = "web"
	UniApp ProjectKind = "uni-app"
)

type NewLS func(program *compiler.Program)*ls.LanguageService
type Project struct {
	server *Server

	kind ProjectKind
	fsPath string

	rootLanguageService *dis.Box[*RoutuerLanguageService]
	plugins map[string]*dis.Box[Plugin]

	newLs NewLS
}
var _ dis.Disposable = (*Project)(nil)

func NewProject(fsPath string, server *Server, newLs NewLS) *Project {
	project := &Project{
		server: server,

		fsPath: fsPath,
		plugins: make(map[string]*dis.Box[Plugin]),

		newLs: newLs,
	}

	project.rootLanguageService = dis.NewBox(&RoutuerLanguageService{
		project: project,
	})

	project.init()

	return project
}

func (p *Project) Dispose() {
	p.rootLanguageService.Delete()

	for _, plugin := range p.plugins {
		plugin.Delete()
	}
}
func (p *Project) init()  {
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

func (p *Project) Server() *Server {
	return p.server
}

func (p *Project) Kind() ProjectKind {
	return p.kind
}

func (p *Project) FsPath() string {
	return p.fsPath
}

func (p *Project) GetLanguageService() LanguageService {
	return p.rootLanguageService.Value()
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

func (p *Project) NewLanguageService(program *compiler.Program)* ls.LanguageService{
	return p.newLs(program)
}

type RoutuerLanguageService struct {
	project *Project
}
var _ LanguageService = (*RoutuerLanguageService)(nil)

func (r*RoutuerLanguageService)Dispose(){}

func (r*RoutuerLanguageService)GetProvideCompletion(defaultLs *ls.LanguageService)(func(ctx context.Context,documentURI lsproto.DocumentUri,LSPPosition lsproto.Position,context *lsproto.CompletionContext) (lsproto.CompletionResponse, error)){
	return func(ctx context.Context,documentURI lsproto.DocumentUri,LSPPosition lsproto.Position,context *lsproto.CompletionContext,) (lsproto.CompletionResponse, error){
		plugins := r.project.GetPlugins()
		for _, plugin := range plugins{
			if(plugin == nil){
				continue
			}
			if pls := plugin.GetLanguageService(defaultLs); pls != nil && pls.IsEnable(documentURI){
				if fn := pls.GetProvideCompletion(defaultLs); fn != nil{
					return fn(ctx, documentURI, LSPPosition, context)
				}
			}
		}
		return defaultLs.ProvideCompletion(ctx, documentURI, LSPPosition, context)
	}
}