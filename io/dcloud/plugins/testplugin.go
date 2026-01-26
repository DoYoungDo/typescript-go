package plugins

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/microsoft/typescript-go/internal/ast"
	"github.com/microsoft/typescript-go/internal/checker"
	"github.com/microsoft/typescript-go/internal/compiler"
	"github.com/microsoft/typescript-go/internal/core"
	"github.com/microsoft/typescript-go/internal/ls"
	"github.com/microsoft/typescript-go/internal/lsp/lsproto"
	"github.com/microsoft/typescript-go/internal/module"
	"github.com/microsoft/typescript-go/internal/tspath"
	"github.com/microsoft/typescript-go/internal/vfs"
	"github.com/microsoft/typescript-go/io/dcloud"
	dis "github.com/microsoft/typescript-go/io/dcloud/disposable"
)

type TestPlugin struct {
	project *dcloud.Project
	// customLanguageServices map[*compiler.Program]* dis.Box[*TestPluginLanguageService]
	// lastVersion string
	lastProgram *compiler.Program
	lastLs *TestPluginLanguageService

	androidLibMap AndroidLibMap
}

var _ dcloud.Plugin = (*TestPlugin)(nil)

func NewTestPlugin(project* dcloud.Project) (dcloud.Plugin ,error) {
	plugin := &TestPlugin{
		project: project,
	}
	
	data, _ := project.FS().ReadFile("/Users/doyoung/Project/uts-development-android/uts-types/app-android/typeMap.json")
	json.Unmarshal([]byte(data), &plugin.androidLibMap)
	return plugin, nil
}

func (p *TestPlugin) Dispose() {
}

func (p *TestPlugin) GetLanguageService(ctx context.Context, defaultLs *ls.LanguageService, documentURI lsproto.DocumentUri) dcloud.PluginLanguageService{
	if !strings.HasSuffix(documentURI.FileName(), ".uts"){
		return nil
	}

	if p.lastProgram != nil{
		if p.lastLs != nil && p.lastLs.GetProgram() == p.lastProgram{
			return p.lastLs
		}

		return p.updateLs()
	}

	program := p.project.CreateListenedProgram(dcloud.ProgramOptions{
		GetFiles:func() []string {
			files := append(p.project.GetRootFiles(), "/Users/doyoung/OtherProject/typescript-go/io/dcloud/test/app.d.ts", "/Users/doyoung/OtherProject/typescript-go/internal/io/dcloud/app_virtual.d.ts")
			fs := p.project.FS()
			for _, name := range p.androidLibMap.DefaultImport{
				dirs := strings.Split(name, ".")
				name := dirs[len(dirs) -1]
				dirs = dirs[:len(dirs) -1]
				dir := strings.Join(dirs, "/")
				filePath := "/Users/doyoung/Project/uts-development-android/uts-types/app-android/" + dir + "/" + name + ".d.ts"
				if fs.FileExists(filePath){
					files = append(files, filePath)
				}
			}
			return files
		},
		Plugin: &testProgramPlugin{
		resolverPlugins: []module.ResolverPlugin{newTestResolverPlugin(p.project, &p.androidLibMap)},
		checkerPlugins: []checker.CheckerPlugin{&testCheckerPlugin{enable: true}},
	},
		DefaultProgram: defaultLs.GetProgram(),
	}, func(program *compiler.Program, _ dcloud.FileChangedSummary) {
		p.lastProgram = program
	})


	p.lastProgram = program
	return p.updateLs()
}

func (p *TestPlugin) updateLs() *TestPluginLanguageService{
	if p.lastProgram != nil {
		lsHost := NewLanguageServiceHost(p.project, p.lastProgram)
		LS := &TestPluginLanguageService{
			LanguageService: ls.NewLanguageService(tspath.Path(p.project.FsPath()), p.lastProgram, lsHost),
			project: p.project,
			host: lsHost,
		}
		p.lastLs = LS
	}

	return p.lastLs
}


type testProgramPlugin struct{
	resolverPlugins []module.ResolverPlugin
	checkerPlugins []checker.CheckerPlugin
}
var _ compiler.ProgramPlugin = (*testProgramPlugin)(nil)
func (t *testProgramPlugin) GetResolverPlugins() []module.ResolverPlugin{
	return t.resolverPlugins
}
func (t *testProgramPlugin) GetCheckerPlugins() []checker.CheckerPlugin{
	return t.checkerPlugins
}


type testResolverPlugin struct{
	fs vfs.FS
	enable bool

	androidLibMap *AndroidLibMap
}
var _ module.ResolverPlugin = (*testResolverPlugin)(nil)

func newTestResolverPlugin(project *dcloud.Project,androidLibMap *AndroidLibMap) *testResolverPlugin{
	plugin := &testResolverPlugin{
		fs: project.FS(),
		enable: true,
		androidLibMap: androidLibMap,
	}

	return plugin
}

func (t *testResolverPlugin) GetResolveModuleName()(func(moduleName string, containingFile string, resolutionMode core.ResolutionMode, redirectedReference module.ResolvedProjectReference) (*module.ResolvedModule, []module.DiagAndArgs)){
	// mod := ast.ModuleName
	return func(moduleName, containingFile string, resolutionMode core.ResolutionMode, redirectedReference module.ResolvedProjectReference) (*module.ResolvedModule, []module.DiagAndArgs) {
		if info, ok := t.androidLibMap.Map[moduleName]; ok{
			if info.Tp == "class"{
				dirs := strings.Split(info.Name, ".")
				dirs = dirs[:len(dirs) -1]
				dir := strings.Join(dirs, "/")
				filePath := "/Users/doyoung/Project/uts-development-android/uts-types/app-android/" + dir + "/" + info.File
				if t.fs.FileExists(filePath){
					return &module.ResolvedModule{
						ResolvedFileName: filePath,
						Extension: ".d.ts",
						IsExternalLibraryImport: false,
					}, []module.DiagAndArgs{}
				}
			}

		}
		return nil, nil
	}
}
func (t *testResolverPlugin) IsEnable() bool{
	return t.enable
}
func (t *testResolverPlugin) SetEnable(en bool){
	t.enable = en;
}


type testCheckerPlugin struct{
	enable bool
}
var _ checker.CheckerPlugin = (*testCheckerPlugin)(nil)
func (t *testCheckerPlugin) IsEnable() bool{
	return t.enable
}
func (t *testCheckerPlugin) SetEnable(en bool){
	t.enable = en;
}
func (t *testCheckerPlugin) GetCheckExpressionWorker()(func(node *ast.Node, checkMode checker.CheckMode) *checker.Type){
	return nil
}


type TestPluginLanguageService struct {
	*ls.LanguageService
	project *dcloud.Project
	host dcloud.LanguageServiceHost
}
var(
	_ dcloud.PluginLanguageService = (*TestPluginLanguageService)(nil)
	_ dis.Disposable = (*TestPluginLanguageService)(nil)
)

func (l *TestPluginLanguageService) Dispose() {}

func (l *TestPluginLanguageService)	IsEnable(fileName lsproto.DocumentUri)bool{
	return true
}

func (l *TestPluginLanguageService) GetHost() dcloud.LanguageServiceHost{
	return l.host
}

func (l *TestPluginLanguageService)	GetProvideCompletion()(func(ctx context.Context,documentURI lsproto.DocumentUri,LSPPosition lsproto.Position,context *lsproto.CompletionContext) (lsproto.CompletionResponse, error)){
	return l.LanguageService.ProvideCompletion;
}


type CVFS struct{}
var _ VirtualFS = (*CVFS)(nil)
func (vfs *CVFS) FileExists(path string) bool {
	if strings.HasSuffix(path, "app_virtual.d.ts") {
		return true;
	}
	return  false;
}
func (vfs *CVFS) ReadFile(path string) (contents string, ok bool) {
	if strings.HasSuffix(path, "app_virtual.d.ts") {
		return "declare const DCloud_virtual: any;", true
	}
	return "", false
}

type TreeInfo map[string]TreeInfo
type MapInfo map[string]struct{
	File string `json:"file"`
	Name string `json:"name"`
	Tp string `json:"type"`
}
type AndroidLibMap struct{
	DefaultImport []string `json:defaultImport`
	Tree TreeInfo `json:"tree"`
	Map MapInfo `json:"map"`
}