package plugins

import (
	"context"
	"strings"

	"github.com/microsoft/typescript-go/internal/compiler"
	"github.com/microsoft/typescript-go/internal/core"
	"github.com/microsoft/typescript-go/internal/ls"
	"github.com/microsoft/typescript-go/internal/lsp/lsproto"
	"github.com/microsoft/typescript-go/internal/tsoptions"
	"github.com/microsoft/typescript-go/internal/tspath"
	"github.com/microsoft/typescript-go/io/dcloud"
)

type TestPlugin struct {
	project *dcloud.Project
	TestLs dcloud.PluginLanguageService
}

var _ dcloud.Plugin = (*TestPlugin)(nil)

func NewTestPlugin(project* dcloud.Project) (dcloud.Plugin ,error) {
	return &TestPlugin{
		project: project,
	}, nil
}

func (p *TestPlugin) GetLanguageService(defaultLs *ls.LanguageService) dcloud.PluginLanguageService {
	program := defaultLs.GetProgram()
	files:=append(program.CommandLine().ParsedConfig.FileNames, "/Users/doyoung/OtherProject/typescript-go/io/dcloud/test/app.d.ts", "/Users/doyoung/OtherProject/typescript-go/internal/io/dcloud/app_virtual.d.ts")
	opts := compiler.ProgramOptions{
		Host: dcloud.NewCompilerHost(p.project.FsPath(),dcloud.NewVirtualFileSystem(&CVFS{}),"",nil,nil),
		Config: tsoptions.NewParsedCommandLine(&core.CompilerOptions{},files,tspath.ComparePathsOptions{
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
	// if p.TestLs == nil{
		p.TestLs = &TestPluginLanguageService{
			LanguageService: ls.NewLanguageService(newProgram, p.project.Server().GetDefaultHost()),
		}
	// }
	return  p.TestLs
}

type TestPluginLanguageService struct {
	*ls.LanguageService
}
var _ dcloud.PluginLanguageService = (*TestPluginLanguageService)(nil)

func (l *TestPluginLanguageService)	IsEnable(fileName lsproto.DocumentUri)bool{
	return true
}

func (l *TestPluginLanguageService)	GetProvideCompletion(ls *ls.LanguageService)(func(ctx context.Context,documentURI lsproto.DocumentUri,LSPPosition lsproto.Position,context *lsproto.CompletionContext) (lsproto.CompletionResponse, error)){
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