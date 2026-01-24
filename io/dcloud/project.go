package dcloud

import (
	"context"
	"strconv"
	"sync"

	"github.com/microsoft/typescript-go/internal/compiler"
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

type FileHandle struct{
	path tspath.Path
	content string
}
type FileChangedSummary struct{
	closedFiles []FileHandle
	openedFiles []FileHandle
	changedFiles []FileHandle
}

type Project struct {
	kind ProjectKind
	fsPath string
	rootFiles []string
	version uint64

	plugins map[string]*dis.Box[Plugin]

	fs vfs.FS

	programWatchedChannels []chan FileChangedSummary 
	programWatchedChannelsGroup sync.WaitGroup
	programWatchedChannelsMu sync.Mutex
}
var _ dis.Disposable = (*Project)(nil)

func NewProject(fsPath tspath.Path, fs vfs.FS) *Project{
	project := &Project{
		fsPath: string(fsPath),
		version: 0,
		plugins: make(map[string]*dis.Box[Plugin]),
		fs: fs,
	}
	project.Init()
	return project
}

func (p *Project) Dispose() {
	// 关闭所有 channels，让 goroutines 退出
	p.programWatchedChannelsMu.Lock()
	defer p.programWatchedChannelsMu.Unlock()
    for _, ch := range p.programWatchedChannels {
        close(ch)
    }
    p.programWatchedChannels = nil

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

func (p *Project) Version() string{
	return strconv.FormatUint(p.version, 10)
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

type ProgramOptions struct{
	GetFiles func() []string
	Plugin compiler.ProgramPlugin
	DefaultProgram *compiler.Program
}
func (p *Project) CreateListenedProgram(opt compiler.ProgramOptions, update func(program *compiler.Program, file FileHandle)) *compiler.Program{
	program := compiler.NewProgram(opt)

	ch := make(chan FileChangedSummary, 5)

	p.programWatchedChannelsMu.Lock()
	defer p.programWatchedChannelsMu.Unlock()
	p.programWatchedChannels = append(p.programWatchedChannels, ch)
	
	go func(){
		for changed := range ch{
			defer func ()  {
				close(ch)
				p.programWatchedChannelsGroup.Done()
			}()
			if len(changed.closedFiles) > 0{

			}
			// if host, ok := opt.Host.(CompilerHost); ok{
			// 	host.SetForeceParseSourceFile(true)
			// 	defer host.SetForeceParseSourceFile(false)
			// }

			// newProgram, _ := program.UpdateProgram(file.path, opt.Host)
			// if update != nil{
			// 	update(newProgram, file)
			// }
		}
	}()

	return program
}