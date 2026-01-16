package plugins

import (
	"context"
	"strings"

	"github.com/microsoft/typescript-go/internal/compiler"
	"github.com/microsoft/typescript-go/internal/ls"
	"github.com/microsoft/typescript-go/internal/lsp/lsproto"
	"github.com/microsoft/typescript-go/internal/tsoptions"
	"github.com/microsoft/typescript-go/internal/tspath"
	"github.com/microsoft/typescript-go/io/dcloud"
	dis "github.com/microsoft/typescript-go/io/dcloud/disposable"
)

type TestPlugin struct {
	project *dcloud.Project
	TestLs *dis.Box[*TestPluginLanguageService]
}

var _ dcloud.Plugin = (*TestPlugin)(nil)

func NewTestPlugin(project* dcloud.Project) (dcloud.Plugin ,error) {
	return &TestPlugin{
		project: project,
	}, nil
}

func (p *TestPlugin) Dispose() {
	if p.TestLs != nil {
		p.TestLs.Delete()
		p.TestLs = nil
	}
}

func (p *TestPlugin) GetLanguageService(defaultLs *ls.LanguageService) dcloud.PluginLanguageService {
	if p.TestLs == nil{
		program := defaultLs.GetProgram()
		files:=append(program.CommandLine().ParsedConfig.FileNames, "/Users/doyoung/OtherProject/typescript-go/io/dcloud/test/app.d.ts", "/Users/doyoung/OtherProject/typescript-go/internal/io/dcloud/app_virtual.d.ts")
		// files:=program.CommandLine().ParsedConfig.FileNames
		opts := compiler.ProgramOptions{
			Host: dcloud.NewCompilerHost(p.project.FsPath(),dcloud.NewVirtualFileSystem(&CVFS{}, program),"",nil,nil),
			Config: tsoptions.NewParsedCommandLine(program.CommandLine().CompilerOptions(),files,tspath.ComparePathsOptions{
				UseCaseSensitiveFileNames :false,
				CurrentDirectory:p.project.FsPath(),
			}),
			// UseSourceOfProjectReference :program.UseCaseSensitiveFileNames(),
			// SingleThreaded:program.Options().SingleThreaded,
			// CreateCheckerPool:program.CreateCheckerPool,
			// TypingsLocation:program.GetGlobalTypingsCacheLocation(),
			// ProjectName:program.GetProjectName(),
		}
		newProgram := compiler.NewProgram(opts)
		newProgram.BindSourceFiles()
		// newLs := ls.NewLanguageService(newProgram, l.project.Server().GetDefaultHost())
		// res , _:= defaultLs.ProvideCompletion(ctx, documentURI, LSPPosition, context)
		// len(res.List.Items)
		// res.List.Items = append(res.List.Items)
		// newRes, err := newLs.ProvideCompletion(ctx, documentURI, LSPPosition, context)
		// return newRes, err
		p.TestLs = dis.NewBox(&TestPluginLanguageService{
			LanguageService: ls.NewLanguageService(newProgram, p.project.Server().GetDefaultHost()),
			// LanguageService: defaultLs,
			project: p.project,
		})
	} else {
		// program := p.TestLs.Value().GetProgram()
		// program.UpdateProgram()
	}
	return  p.TestLs.Value()
}

type TestPluginLanguageService struct {
	*ls.LanguageService
	project *dcloud.Project
}
var _ dcloud.PluginLanguageService = (*TestPluginLanguageService)(nil)
var _ dis.Disposable = (*TestPluginLanguageService)(nil)

func (l *TestPluginLanguageService) Dispose() {}

func (l *TestPluginLanguageService)	IsEnable(fileName lsproto.DocumentUri)bool{
	return true
}

func (l *TestPluginLanguageService)	GetProvideCompletion(defaultLs *ls.LanguageService)(func(ctx context.Context,documentURI lsproto.DocumentUri,LSPPosition lsproto.Position,context *lsproto.CompletionContext) (lsproto.CompletionResponse, error)){
	return l.LanguageService.ProvideCompletion;
}

type CVFS struct{}
var _ dcloud.VirtualFS = (*CVFS)(nil)
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

// func (l *LanguageService) ProvideCompletion(
// 	ctx context.Context,
// 	documentURI lsproto.DocumentUri,
// 	LSPPosition lsproto.Position,
// 	context *lsproto.CompletionContext,
// ) (lsproto.CompletionResponse, error) {

// 	snapShot ,_ := l.session.Snapshot()
// 	newLs := ls.NewLanguageService(l.GetProgram(), snapShot)
// 	res, err :=newLs.ProvideCompletion(ctx, documentURI, LSPPosition, context)



// 	program := l.GetProgram()
// 	files:=append(program.CommandLine().ParsedConfig.FileNames, "/Users/doyoung/OtherProject/typescript-go/internal/io/dcloud/app.d.ts", "/Users/doyoung/OtherProject/typescript-go/internal/io/dcloud/app_virtual.d.ts")
// 	opts := compiler.ProgramOptions{
// 		Host: newCompilerHost(l.project.fsPath,NewVirtualFileSystem(&CVFS{}),"",nil,nil),
// 		Config: tsoptions.NewParsedCommandLine(&core.CompilerOptions{},files,tspath.ComparePathsOptions{
// 			UseCaseSensitiveFileNames :false,
// 			CurrentDirectory:l.project.fsPath,
// 		}),
// 		// UseSourceOfProjectReference :program.UseCaseSensitiveFileNames(),
// 		// SingleThreaded:program.Options().SingleThreaded,
// 		// CreateCheckerPool:program.CreateCheckerPool,
// 		// TypingsLocation:program.GetGlobalTypingsCacheLocation(),
// 		// ProjectName:program.GetProjectName(),
// 	}
// 	// opts.Config.ParsedConfig.FileNames = append(opts.Config.ParsedConfig.FileNames, "/Users/doyoung/OtherProject/typescript-go/internal/io/dcloud/app.d.ts")

// 	newProgram := compiler.NewProgram(opts)
// 	newLs1 := ls.NewLanguageService(newProgram, snapShot)
// 	res1, _ :=newLs1.ProvideCompletion(ctx, documentURI, lsproto.Position{Line: LSPPosition.Line, Character: LSPPosition.Character - 1}, context)
	

// 	// res, err := l.LanguageService.ProvideCompletion(ctx, documentURI, LSPPosition, context)

// 	if res.List != nil {
// 		kind := lsproto.CompletionItemKindSnippet
// 		res.List.Items = append(res.List.Items, &lsproto.CompletionItem{
// 			Label: "DCloud_Snippet",
// 			Kind:  &kind,
// 		})
// 		res.List.Items = append(res.List.Items, res1.List.Items...)
// 	}
// 	return res, err
// }