package plugins

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/microsoft/typescript-go/internal/ast"
	"github.com/microsoft/typescript-go/internal/checker"
	"github.com/microsoft/typescript-go/internal/compiler"
	"github.com/microsoft/typescript-go/internal/core"
	"github.com/microsoft/typescript-go/internal/ls"
	"github.com/microsoft/typescript-go/internal/lsp/lsproto"
	"github.com/microsoft/typescript-go/internal/module"
	"github.com/microsoft/typescript-go/internal/vfs"
	"github.com/microsoft/typescript-go/io/dcloud"
	dis "github.com/microsoft/typescript-go/io/dcloud/disposable"
)

type TestPlugin struct {
	project *dcloud.Project
	// lastProgram *compiler.Program
	// lastLs *TestPluginLanguageService

	androidLibMap AndroidLibMap
	iosLibMap IOSLibMap

	cls *ConcurrentLanguageService
}

var _ dcloud.Plugin = (*TestPlugin)(nil)

func NewTestPlugin(project* dcloud.Project) (dcloud.Plugin ,error) {
	plugin := &TestPlugin{
		project: project,
	}
	
	data, _ := project.FS().ReadFile("/Users/doyoung/Project/uts-development-android/uts-types/app-android/typeMap.json")
	json.Unmarshal([]byte(data), &plugin.androidLibMap)

	data, _ = project.FS().ReadFile("/Users/doyoung/Project/uts-development-ios/uts-types/app-ios/typeMap.json")
	json.Unmarshal([]byte(data), &plugin.iosLibMap)

	return plugin, nil
}

func (p *TestPlugin) Dispose() {
}

func (p *TestPlugin) GetLanguageService(ctx context.Context, defaultLs *ls.LanguageService, documentURI lsproto.DocumentUri) dcloud.PluginLanguageService{
	if !strings.HasSuffix(documentURI.FileName(), ".uts"){
		return nil
	}

	if p.cls == nil{
		p.cls = NewConcurrentLS(p.project, defaultLs, &p.androidLibMap, &p.iosLibMap)
	}
	p.cls.SyncLs()
	return p.cls 

	// code, ok := p.project.FS().ReadFile("/Users/doyoung/Project/hbuilderx-language-services-tsgo-tests/uniapp-x-default/pages/index/index.uvue")
	// if !ok{

	// }

	// parser := tree_sitter.NewParser()
	// defer parser.Close()
	// parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_html.Language()))

	// tree := parser.Parse([]byte(code), nil)
	// defer tree.Close()

	// cursor := tree.Walk()
	// root := tree.RootNode()
	// if root != nil{
	// 	s := root.ToSexp()
	// 	if s != ""{}
	// 	cc := root.ChildCount()
	// 	if cc > 0{}
	// 	var walk func(node *tree_sitter.Node)
	// 	walk = func(node *tree_sitter.Node) {
	// 		if node == nil{
	// 			return
	// 		}

	// 		text := node.Utf8Text([]byte(code))
	// 		kd := node.Kind()
	// 		gn := node.GrammarName()
	// 		if text != "" || gn != "" || kd != ""{}

	// 		if node.ChildCount() > 0{
	// 			for _, child := range node.Children(cursor){
	// 				walk(&child)
	// 			}
	// 		}
	// 	}
	// 	walk(root)
	// }

	

	// if p.lastProgram != nil{
	// 	if p.lastLs != nil && p.lastLs.GetProgram() == p.lastProgram{
	// 		return p.lastLs
	// 	}

	// 	return p.updateLs()
	// }

	// program := p.project.CreateListenedProgram(dcloud.ProgramOptions{
	// 	GetFiles:func() []string {
	// 		files := append(p.project.GetRootFiles(), "/Users/doyoung/OtherProject/typescript-go/io/dcloud/test/app.d.ts", "/Users/doyoung/OtherProject/typescript-go/internal/io/dcloud/app_virtual.d.ts")
	// 		fs := p.project.FS()
	// 		for _, name := range p.androidLibMap.DefaultImport{
	// 			dirs := strings.Split(name, ".")
	// 			name := dirs[len(dirs) -1]
	// 			dirs = dirs[:len(dirs) -1]
	// 			dir := strings.Join(dirs, "/")
	// 			filePath := "/Users/doyoung/Project/uts-development-android/uts-types/app-android/" + dir + "/" + name + ".d.ts"
	// 			if fs.FileExists(filePath){
	// 				files = append(files, filePath)
	// 			}
	// 		}
	// 		return files
	// 	},
	// 	Plugin: &testProgramPlugin{
	// 	resolverPlugins: []module.ResolverPlugin{newTestResolverPlugin(p.project, &p.androidLibMap)},
	// 	checkerPlugins: []checker.CheckerPlugin{&testCheckerPlugin{enable: true}},
	// },
	// 	DefaultProgram: defaultLs.GetProgram(),
	// }, func(program *compiler.Program, _ dcloud.FileChangedSummary) {
	// 	p.lastProgram = program
	// })


	// p.lastProgram = program
	// return p.updateLs()
}

// func (p *TestPlugin) updateLs() *TestPluginLanguageService{
// 	if p.lastProgram != nil {
// 		lsHost := NewLanguageServiceHost(p.project, p.lastProgram)
// 		LS := &TestPluginLanguageService{
// 			LanguageService: ls.NewLanguageService(tspath.Path(p.project.FsPath()), p.lastProgram, lsHost),
// 			project: p.project,
// 			host: lsHost,
// 		}
// 		p.lastLs = LS
// 	}

// 	return p.lastLs
// }


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

	libMap *LibMap
}
var _ module.ResolverPlugin = (*testResolverPlugin)(nil)

func newTestResolverPlugin(project *dcloud.Project,libMap *LibMap) *testResolverPlugin{
	plugin := &testResolverPlugin{
		fs: project.FS(),
		enable: true,
		libMap: libMap,
	}

	return plugin
}

func (t *testResolverPlugin) GetResolveModuleName()(func(moduleName string, containingFile string, resolutionMode core.ResolutionMode, redirectedReference module.ResolvedProjectReference) (*module.ResolvedModule, []module.DiagAndArgs)){
	// mod := ast.ModuleName
	return func(moduleName, containingFile string, resolutionMode core.ResolutionMode, redirectedReference module.ResolvedProjectReference) (*module.ResolvedModule, []module.DiagAndArgs) {
		if info, ok := t.libMap.Map[moduleName]; ok{
			switch info.Tp {
			case "class":
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
			case "module":
				filePath := "/Users/doyoung/Project/uts-development-ios/uts-types/app-ios/" + info.Name + "/" + info.File
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

type ConcurrentLanguageServiceHost struct{
	cls *ConcurrentLanguageService
}
type ConcurrentLanguageService struct{
	project *dcloud.Project
	host *ConcurrentLanguageServiceHost
	multLs map[string]struct{
		lastProgram *compiler.Program
		lastLs *TestPluginLanguageService
	}
}
var (
	_ dcloud.LanguageService = (*ConcurrentLanguageService)(nil)
	_ dcloud.LanguageServiceHost = (*ConcurrentLanguageServiceHost)(nil)
)
func NewConcurrentLS(pro *dcloud.Project, defaultLs *ls.LanguageService, androidLid *AndroidLibMap, iosLib *IOSLibMap)* ConcurrentLanguageService{
	host := &ConcurrentLanguageServiceHost{}
	cls := &ConcurrentLanguageService{
		project: pro,
		host: host,
		multLs: make(map[string]struct{lastProgram *compiler.Program; lastLs *TestPluginLanguageService}),
	}

	createProgram := func (name string)  {
		program := pro.CreateListenedProgram(dcloud.ProgramOptions{
			GetFiles:func() []string {
				files := append(pro.GetRootFiles(), "/Users/doyoung/OtherProject/typescript-go/io/dcloud/test/app.d.ts", "/Users/doyoung/OtherProject/typescript-go/internal/io/dcloud/app_virtual.d.ts")
				// fs := pro.FS()
				// for _, name := range androidLid.DefaultImport{
				// 	dirs := strings.Split(name, ".")
				// 	name := dirs[len(dirs) -1]
				// 	dirs = dirs[:len(dirs) -1]
				// 	dir := strings.Join(dirs, "/")
				// 	filePath := "/Users/doyoung/Project/uts-development-android/uts-types/app-android/" + dir + "/" + name + ".d.ts"
				// 	if fs.FileExists(filePath){
				// 		files = append(files, filePath)
				// 	}
				// }
				return files
			},
			Plugin: &testProgramPlugin{
			resolverPlugins: []module.ResolverPlugin{/* newTestResolverPlugin(pro, &androidLid.LibMap),newTestResolverPlugin(pro, &iosLib.LibMap) */},
			checkerPlugins: []checker.CheckerPlugin{&testCheckerPlugin{enable: true}},
		},
			DefaultProgram: defaultLs.GetProgram(),
		}, func(program *compiler.Program, _ dcloud.FileChangedSummary) {
			cls.multLs[name] = struct{lastProgram *compiler.Program; lastLs *TestPluginLanguageService}{
				lastProgram: program,
				lastLs: nil,
			}
		})	

		cls.multLs[name] = struct{
			lastProgram *compiler.Program 
			lastLs *TestPluginLanguageService
		}{lastProgram: program, lastLs: nil}
	}

	createProgram("android")
	// createProgram("ios")
	// createProgram("harmony")
	// createProgram("web")
	// createProgram("mp-weixin")

	host.cls = cls
	return cls
}

func (c *ConcurrentLanguageServiceHost) UpdateAutoImport(ctx context.Context, uri lsproto.DocumentUri){
	for _, msl := range c.cls.multLs{
		if msl.lastLs != nil{
			msl.lastLs.GetHost().UpdateAutoImport(ctx, uri)
		}
	}
}
func (*ConcurrentLanguageService) Dispose(){}
func (c *ConcurrentLanguageService) GetHost() dcloud.LanguageServiceHost{
	return c.host
}
func (c *ConcurrentLanguageService) SyncLs(){
	start := time.Now()
	defer func ()  {
		println("create ls spend:", time.Since(start).Milliseconds())
	}()

	for name, msl := range c.multLs{
		lastProgram := msl.lastProgram
		lastLs := msl.lastLs

		if lastProgram != nil {
			if lastLs != nil && lastLs.GetProgram() == lastProgram {

			}else{
				lsHost := NewLanguageServiceHost(c.project, lastProgram)
				lastLs = &TestPluginLanguageService{
					LanguageService: ls.NewLanguageService(c.project.ToPath(c.project.FsPath()), lastProgram, lsHost),
					project: c.project,
					host: lsHost,
				}
				// 重新赋值
				c.multLs[name] = struct{
					lastProgram *compiler.Program
					lastLs *TestPluginLanguageService
				}{lastProgram: lastProgram, lastLs: lastLs}
			}
		}
	}
}

func (c *ConcurrentLanguageService) GetProvideCompletion()(func(ctx context.Context,documentURI lsproto.DocumentUri,LSPPosition lsproto.Position,context *lsproto.CompletionContext) (lsproto.CompletionResponse, error)){
	return func(ctx context.Context, documentURI lsproto.DocumentUri, LSPPosition lsproto.Position, context *lsproto.CompletionContext) (lsproto.CompletionResponse, error) {
		start := time.Now()
		defer func()  {
			println(len(c.multLs)," ls call all spend", time.Since(start).Milliseconds(), `\n`)
		}()

		ch := make(chan struct {
			res lsproto.CompletionResponse
			err error
			}, len(c.multLs))

		for _, msl := range c.multLs{
			go func ()  {
				start := time.Now()
				lastLs := msl.lastLs
				if lastLs != nil {
					res, err := lastLs.GetProvideCompletion()(ctx, documentURI, LSPPosition, context)
					ch <- struct{res lsproto.CompletionResponse; err error}{res:res, err:err}
				}else{
					ch <- struct{res lsproto.CompletionResponse; err error}{res:lsproto.CompletionResponse{}, err: nil}
				}
				println(len(c.multLs)," ls call one spend", time.Since(start).Milliseconds(), `\n`)
			}()
		}

		mergeResult := lsproto.CompletionResponse{
			List: &lsproto.CompletionList{
				IsIncomplete: true,
				ItemDefaults: nil,
				ApplyKind: nil,
				Items: []*lsproto.CompletionItem{},
			},
		}


		startMerge := time.Now()
		for i := 0; i < len(c.multLs); i++ {
			res := <-ch
			if res.res.List != nil {
				mergeResult.List.Items = append(mergeResult.List.Items, res.res.List.Items...)
			}
		}
		itemLen := len(mergeResult.List.Items)
		println("merge spend", time.Since(startMerge).Milliseconds(), " item len", itemLen, `\n`)
		max := 4000
		if itemLen > max{
			mergeResult.List.Items = mergeResult.List.Items[:max]
		}

		close(ch)
		return mergeResult, nil
	}
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
type LibMap struct{
	Tree TreeInfo `json:"tree"`
	Map MapInfo `json:"map"`
}
type AndroidLibMap struct{
	DefaultImport []string `json:"defaultImport"`
	LibMap
}
type IOSLibMap struct{
	LibMap
}