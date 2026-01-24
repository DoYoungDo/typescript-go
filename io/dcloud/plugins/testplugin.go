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
	customLanguageServices map[*compiler.Program]* dis.Box[*TestPluginLanguageService]
}

var _ dcloud.Plugin = (*TestPlugin)(nil)

func NewTestPlugin(project* dcloud.Project) (dcloud.Plugin ,error) {
	return &TestPlugin{
		project: project,
		customLanguageServices: make(map[*compiler.Program]*dis.Box[*TestPluginLanguageService]),
	}, nil
}

func (p *TestPlugin) Dispose() {
	for _, ls := range p.customLanguageServices{
		if ls != nil{
			ls.Delete()
		}
	}
	p.customLanguageServices = make(map[*compiler.Program]*dis.Box[*TestPluginLanguageService])
}

func (p *TestPlugin) GetLanguageService(ctx context.Context, defaultLs *ls.LanguageService, documentURI lsproto.DocumentUri) dcloud.PluginLanguageService{
	if !strings.HasSuffix(documentURI.FileName(), ".uts"){
		return nil
	}

	program := defaultLs.GetProgram()
	// if p.customLanguageServices[program] == nil{
		files := append(p.project.GetRootFiles(), "/Users/doyoung/OtherProject/typescript-go/io/dcloud/test/app.d.ts", "/Users/doyoung/OtherProject/typescript-go/internal/io/dcloud/app_virtual.d.ts")
		// files:=program.CommandLine().ParsedConfig.FileNames
		programPlugin := &testProgramPlugin{
			resolverPlugins: []module.ResolverPlugin{newTestResolverPlugin(p.project)},
			checkerPlugins: []checker.CheckerPlugin{&testCheckerPlugin{enable: true}},
		}
		// 创建共用的虚拟文件系缚
		vfs := NewVirtualFileSystem(&CVFS{}, program)
		opts := compiler.ProgramOptions{
			Host: NewCompilerHost(p.project.FsPath(),vfs,"",nil,nil, program),
			Config: tsoptions.NewParsedCommandLine(program.CommandLine().CompilerOptions(),files,tspath.ComparePathsOptions{
				UseCaseSensitiveFileNames :false,
				CurrentDirectory:p.project.FsPath(),
			}),
			// UseSourceOfProjectReference :program.UseCaseSensitiveFileNames(),
			// SingleThreaded:program.Options().SingleThreaded,
			// CreateCheckerPool:program.CreateCheckerPool,
			// TypingsLocation:program.GetGlobalTypingsCacheLocation(),
			// ProjectName:program.GetProjectName(),
			Plugins: []compiler.ProgramPlugin{programPlugin},
		}
		newProgram := compiler.NewProgram(opts)
		newProgram.BindSourceFiles()
		// newLs := ls.NewLanguageService(newProgram, l.project.Server().GetDefaultHost())
		// res , _:= defaultLs.ProvideCompletion(ctx, documentURI, LSPPosition, context)
		// len(res.List.Items)
		// res.List.Items = append(res.List.Items)
		// newRes, err := newLs.ProvideCompletion(ctx, documentURI, LSPPosition, context)
		// return newRes, err

		lsHost := NewLanguageServiceHost(p.project, newProgram)

		// p.customLanguageServices[program] = dis.NewBox(&TestPluginLanguageService{
		return dis.NewBox(&TestPluginLanguageService{
			LanguageService: ls.NewLanguageService(tspath.Path(p.project.FsPath()), newProgram, lsHost),
			project: p.project,
			host: lsHost,
		}).Value()
	// }

	// return p.customLanguageServices[program].Value()
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
