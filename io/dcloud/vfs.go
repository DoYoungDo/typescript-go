package dcloud

import (
	"time"

	"github.com/microsoft/typescript-go/internal/ast"
	"github.com/microsoft/typescript-go/internal/vfs"
	"github.com/microsoft/typescript-go/internal/vfs/osvfs"
)

type VirtualFS interface{
	FileExists(path string) bool;
	ReadFile(path string) (contents string, ok bool);
}

type Cache interface{
	GetSourceFile(filename string) *ast.SourceFile;
}


func NewVirtualFileSystem(fs VirtualFS, cache Cache) vfs.FS {
	return &virtualFileSystem{fs: fs, cache: cache}
}

type virtualFileSystem struct{
	fs VirtualFS
	cache Cache
}

var _ vfs.FS = (*virtualFileSystem)(nil)

var osVFS = osvfs.FS()

func (*virtualFileSystem) UseCaseSensitiveFileNames() bool {
	return osVFS.UseCaseSensitiveFileNames()
}

func (v *virtualFileSystem) FileExists(path string) bool {
	if v.fs != nil && v.fs.FileExists(path) {
		return true
	}
	return osVFS.FileExists(path)
}

func (v *virtualFileSystem) ReadFile(path string) (contents string, ok bool) {
	if v.fs != nil {
		if contents, ok = v.fs.ReadFile(path) ; ok {
			return contents, ok
		}
	}
	if v.cache != nil {
		if sf := v.cache.GetSourceFile(path); sf != nil {
			return sf.Text(), true
		}
	}
	return osVFS.ReadFile(path)
}

func (v *virtualFileSystem) WriteFile(path string, data string, writeByteOrderMark bool) error {
	return osVFS.WriteFile(path, data, writeByteOrderMark)
}

func (v *virtualFileSystem) Remove(path string) error {
	return osVFS.Remove(path)
}

func (v *virtualFileSystem) Chtimes(path string, aTime time.Time, mTime time.Time) error {
	return osVFS.Chtimes(path, aTime, mTime)
}

func (v *virtualFileSystem) DirectoryExists(path string) bool {
	return osVFS.DirectoryExists(path)
}

func (v *virtualFileSystem) GetAccessibleEntries(path string) vfs.Entries {
	return osVFS.GetAccessibleEntries(path)
}

func (v *virtualFileSystem) Stat(path string) vfs.FileInfo {
	return osVFS.Stat(path)
}

func (v *virtualFileSystem) WalkDir(root string, walkFn vfs.WalkDirFunc) error {
	return osVFS.WalkDir(root, walkFn)
}

func (v *virtualFileSystem) Realpath(path string) string {
	return osVFS.Realpath(path)
}