package dcloud

import (
	"github.com/microsoft/typescript-go/internal/ast"
	"github.com/microsoft/typescript-go/internal/compiler"
	"github.com/microsoft/typescript-go/internal/core"
	"github.com/microsoft/typescript-go/internal/diagnostics"
	"github.com/microsoft/typescript-go/internal/parser"
	"github.com/microsoft/typescript-go/internal/tsoptions"
	"github.com/microsoft/typescript-go/internal/tspath"
	"github.com/microsoft/typescript-go/internal/vfs"
)
type CompilerHost struct {
	currentDirectory    string
	fs                  vfs.FS
	defaultLibraryPath  string
	extendedConfigCache tsoptions.ExtendedConfigCache
	trace               func(msg *diagnostics.Message, args ...any)

	getCacheSourceFile  func(path tspath.Path)*ast.SourceFile
}
var _ compiler.CompilerHost = (*CompilerHost)(nil)

func NewCompilerHost(
	currentDirectory string,
	fs vfs.FS,
	defaultLibraryPath string,
	extendedConfigCache tsoptions.ExtendedConfigCache,
	trace func(msg *diagnostics.Message, args ...any),
	getCacheSourceFile  func(path tspath.Path)*ast.SourceFile,
) *CompilerHost {
	if trace == nil {
		trace = func(msg *diagnostics.Message, args ...any) {}
	}
	return &CompilerHost{
		currentDirectory:    currentDirectory,
		fs:                  fs,
		defaultLibraryPath:  defaultLibraryPath,
		extendedConfigCache: extendedConfigCache,
		trace:               trace,
		getCacheSourceFile: getCacheSourceFile,
	}
}

func (h *CompilerHost) FS() vfs.FS {
	return h.fs
}

func (h *CompilerHost) DefaultLibraryPath() string {
	return h.defaultLibraryPath
}

func (h *CompilerHost) GetCurrentDirectory() string {
	return h.currentDirectory
}

func (h *CompilerHost) Trace(msg *diagnostics.Message, args ...any) {
	h.trace(msg, args...)
}

func (h *CompilerHost) GetSourceFile(opts ast.SourceFileParseOptions) *ast.SourceFile {
	if h.getCacheSourceFile != nil{
		if ast := h.getCacheSourceFile(opts.Path); ast != nil{
			return  ast
		}
	}

	text, ok := h.FS().ReadFile(opts.FileName)
	if !ok {
		return nil
	}
	return parser.ParseSourceFile(opts, text, core.GetScriptKindFromFileName(opts.FileName))
}

func (h *CompilerHost) GetResolvedProjectReference(fileName string, path tspath.Path) *tsoptions.ParsedCommandLine {
	commandLine, _ := tsoptions.GetParsedCommandLineOfConfigFilePath(fileName, path, nil, nil /*optionsRaw*/, h, h.extendedConfigCache)
	return commandLine
}
