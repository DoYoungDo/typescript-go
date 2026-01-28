package plugins

import (
	"context"

	"github.com/microsoft/typescript-go/internal/ast"
	"github.com/microsoft/typescript-go/internal/compiler"
	"github.com/microsoft/typescript-go/internal/ls"
	"github.com/microsoft/typescript-go/internal/ls/autoimport"
	"github.com/microsoft/typescript-go/internal/ls/lsconv"
	"github.com/microsoft/typescript-go/internal/ls/lsutil"
	"github.com/microsoft/typescript-go/internal/lsp/lsproto"
	"github.com/microsoft/typescript-go/internal/packagejson"
	"github.com/microsoft/typescript-go/internal/sourcemap"
	"github.com/microsoft/typescript-go/internal/tspath"
	"github.com/microsoft/typescript-go/internal/vfs"
	"github.com/microsoft/typescript-go/io/dcloud"
)

type LanguageServiceHost struct{
	project *dcloud.Project
	program *compiler.Program
	converters *lsconv.Converters
	registry *autoimport.Registry
	allUserPreferences *lsutil.UserConfig
}
var (
	_ ls.Host = (*LanguageServiceHost)(nil)
	_ dcloud.LanguageServiceHost = (dcloud.LanguageServiceHost)(nil)
)

func NewLanguageServiceHost(project *dcloud.Project, program *compiler.Program) *LanguageServiceHost{
	host := &LanguageServiceHost{
		project: project,
		program: program,
		allUserPreferences: lsutil.NewUserConfig(nil),
	}
	host.converters = lsconv.NewConverters(lsproto.PositionEncodingKindUTF8, host.LSPLineMap)
	host.registry = autoimport.NewRegistry(func(fileName string) tspath.Path{
		return tspath.ToPath(fileName, program.GetCurrentDirectory(), host.UseCaseSensitiveFileNames())
	}, host.allUserPreferences.TS())

	return host
}

func (l *LanguageServiceHost) UseCaseSensitiveFileNames() bool{
	return  l.program.Host().FS().UseCaseSensitiveFileNames()
}

func (l *LanguageServiceHost) ReadFile(path string) (contents string, ok bool){
	file := l.program.GetSourceFile(path)
	if file != nil {
		return file.Text(), true
	}
	return l.program.Host().FS().ReadFile(path)
}

func (l *LanguageServiceHost) Converters() *lsconv.Converters{
	return l.converters
}

func (l *LanguageServiceHost) GetPreferences(activeFile string) *lsutil.UserPreferences{
	return l.allUserPreferences.GetPreferences(activeFile)
}

func (l *LanguageServiceHost) GetECMALineInfo(fileName string) *sourcemap.ECMALineInfo{
	return nil
}

func (l *LanguageServiceHost) AutoImportRegistry() *autoimport.Registry{
	return l.registry
}

func (l *LanguageServiceHost) LSPLineMap(fileName string) *lsconv.LSPLineMap {
	file := l.program.GetSourceFile(fileName)
	return lsconv.ComputeLSPLineStarts(file.Text())
}

func (l *LanguageServiceHost) UpdateAutoImport(ctx context.Context, uri lsproto.DocumentUri){
	path := uri.Path(l.UseCaseSensitiveFileNames())

	openFiles := make(map[tspath.Path]string)
	openFiles[path] = uri.FileName()

	registry, err := l.registry.Clone(ctx, autoimport.RegistryChange{
		RequestedFile: path,
		OpenFiles: openFiles,
	}, &registryCloneHost{
		project: l.project,
		program: l.program,
	}, nil)

	if err == nil{
		l.registry = registry
	}
}

type registryCloneHost struct{
	project *dcloud.Project
	program *compiler.Program
}
var _ autoimport.RegistryCloneHost = (*registryCloneHost)(nil)

func (r *registryCloneHost) FS() vfs.FS{
	return r.program.Host().FS()
}

func (r *registryCloneHost) GetCurrentDirectory() string{
	return r.program.GetCurrentDirectory()
}

func (r *registryCloneHost) GetDefaultProject(_ tspath.Path) (tspath.Path, *compiler.Program){
	return r.project.ToPath(r.project.FsPath()), r.program
}

func (r *registryCloneHost) GetProgramForProject(projectPath tspath.Path) *compiler.Program{
	return r.program
}

func (r *registryCloneHost) GetPackageJson(fileName string) *packagejson.InfoCacheEntry{
	fs := r.FS()
	content, _ := fs.ReadFile(fileName)
	
	fields, err := packagejson.Parse([]byte(content))
	if err != nil {
		return &packagejson.InfoCacheEntry{
			DirectoryExists:  true,
			PackageDirectory: tspath.GetDirectoryPath(fileName),
			Contents: &packagejson.PackageJson{
				Parseable: false,
			},
		}
	}
	return &packagejson.InfoCacheEntry{
		DirectoryExists:  true,
		PackageDirectory: tspath.GetDirectoryPath(fileName),
		Contents: &packagejson.PackageJson{
			Fields:    fields,
			Parseable: true,
		},
	}
}

func (r *registryCloneHost) GetSourceFile(fileName string, path tspath.Path) *ast.SourceFile{
	return r.program.GetSourceFileByPath(path)
}

func (r *registryCloneHost) Dispose(){

}

