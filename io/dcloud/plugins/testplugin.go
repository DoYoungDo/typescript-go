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
	"github.com/microsoft/typescript-go/internal/tsoptions"
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
}

var _ dcloud.Plugin = (*TestPlugin)(nil)

func NewTestPlugin(project* dcloud.Project) (dcloud.Plugin ,error) {
	return &TestPlugin{
		project: project,
	}, nil
}

func (p *TestPlugin) Dispose() {
}

func (p *TestPlugin) GetLanguageService(ctx context.Context, defaultLs *ls.LanguageService, documentURI lsproto.DocumentUri) dcloud.PluginLanguageService{
	if !strings.HasSuffix(documentURI.FileName(), ".uts"){
		return nil
	}

	if p.lastProgram != nil && p.lastLs != nil{
		return p.lastLs
	}

	program := core.IfElse(p.lastProgram != nil, p.lastProgram, defaultLs.GetProgram()) 
	// if p.customLanguageServices[program] == nil{
	files := append(p.project.GetRootFiles(), "/Users/doyoung/OtherProject/typescript-go/io/dcloud/test/app.d.ts", "/Users/doyoung/OtherProject/typescript-go/internal/io/dcloud/app_virtual.d.ts")
	// files:=program.CommandLine().ParsedConfig.FileNames
	programPlugin := &testProgramPlugin{
		resolverPlugins: []module.ResolverPlugin{newTestResolverPlugin(p.project)},
		checkerPlugins: []checker.CheckerPlugin{&testCheckerPlugin{enable: true}},
	}
	// 创建共用的虚拟文件系缚
	// vfs := NewVirtualFileSystem(&CVFS{}, program)
	opts := compiler.ProgramOptions{
		Host: NewCompilerHost(p.project.FsPath(),p.project.FS() ,"",nil,nil, func(path tspath.Path)*ast.SourceFile{
			if ast := defaultLs.GetProgram().GetSourceFileByPath(path); ast != nil{
				return ast;
			}
			if p.lastProgram != nil{
				if ast := p.lastProgram.GetSourceFileByPath(path); ast != nil{
					return  ast
				}
			}
			return nil
		}),
		Config: tsoptions.NewParsedCommandLine(program.CommandLine().CompilerOptions(),files,tspath.ComparePathsOptions{
			UseCaseSensitiveFileNames :p.project.FS().UseCaseSensitiveFileNames(),
			CurrentDirectory:p.project.FsPath(),
		}),
		// UseSourceOfProjectReference :program.UseCaseSensitiveFileNames(),
		// SingleThreaded:program.Options().SingleThreaded,
		// CreateCheckerPool:program.CreateCheckerPool,
		// TypingsLocation:program.GetGlobalTypingsCacheLocation(),
		// ProjectName:program.GetProjectName(),
		Plugins: []compiler.ProgramPlugin{programPlugin},
	}
	newProgram := p.project.CreateListenedProgram(opts, func(program *compiler.Program, _ tspath.Path) {
		p.lastProgram = program
	})

	lsHost := NewLanguageServiceHost(p.project, newProgram)

	LS := &TestPluginLanguageService{
		LanguageService: ls.NewLanguageService(tspath.Path(p.project.FsPath()), newProgram, lsHost),
		project: p.project,
		host: lsHost,
	}

	p.lastProgram = newProgram
	p.lastLs = LS

	return LS
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

type TreeInfo map[string]TreeInfo
type MapInfo map[string]struct{
	File string `json:"file"`
	Name string `json:"name"`
	Tp string `json:"type"`
}
type AndroidLibMap struct{
	Tree TreeInfo `json:"tree"`
	Map MapInfo `json:"map"`
}

type testResolverPlugin struct{
	fs vfs.FS
	enable bool
	androidLibMap AndroidLibMap
}
var _ module.ResolverPlugin = (*testResolverPlugin)(nil)

func newTestResolverPlugin(project *dcloud.Project) *testResolverPlugin{
	plugin := &testResolverPlugin{
		fs: project.FS(),
		enable: true,
	}

	data, _ := project.FS().ReadFile("/Users/doyoung/Project/uts-development-android/uts-types/app-android/typeMap.json")
	json.Unmarshal([]byte(data), &plugin.androidLibMap)
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
