package dcloud

import (
	"time"

	"github.com/microsoft/typescript-go/internal/vfs"
	"github.com/microsoft/typescript-go/internal/vfs/osvfs"
)

type VirtualFS interface{
	FileExists(path string) bool;
	ReadFile(path string) (contents string, ok bool);
}


func NewVirtualFileSystem(fs VirtualFS) vfs.FS {
	return &virtualFileSystem{fs: fs}
}

type virtualFileSystem struct{
	fs VirtualFS
}

var _ vfs.FS = (*virtualFileSystem)(nil)

var osVFS = osvfs.FS()

func (vfs *virtualFileSystem) UseCaseSensitiveFileNames() bool {
	return osVFS.UseCaseSensitiveFileNames()
}

func (vfs *virtualFileSystem) FileExists(path string) bool {
	if vfs.fs != nil && vfs.fs.FileExists(path) {
		return true
	}
	return osVFS.FileExists(path)
}

func (vfs *virtualFileSystem) ReadFile(path string) (contents string, ok bool) {
	if vfs.fs != nil {
		if contents, ok = vfs.fs.ReadFile(path) ; ok {
			return contents, ok
		}
	}
	return osVFS.ReadFile(path)
}

func (vfs *virtualFileSystem) WriteFile(path string, data string, writeByteOrderMark bool) error {
	return osVFS.WriteFile(path, data, writeByteOrderMark)
}

func (vfs *virtualFileSystem) Remove(path string) error {
	return osVFS.Remove(path)
}

func (vfs *virtualFileSystem) Chtimes(path string, aTime time.Time, mTime time.Time) error {
	return osVFS.Chtimes(path, aTime, mTime)
}

func (vfs *virtualFileSystem) DirectoryExists(path string) bool {
	return osVFS.DirectoryExists(path)
}

func (vfs *virtualFileSystem) GetAccessibleEntries(path string) vfs.Entries {
	return osVFS.GetAccessibleEntries(path)
}

func (vfs *virtualFileSystem) Stat(path string) vfs.FileInfo {
	return osVFS.Stat(path)
}

func (vfs *virtualFileSystem) WalkDir(root string, walkFn vfs.WalkDirFunc) error {
	return osVFS.WalkDir(root, walkFn)
}

func (vfs *virtualFileSystem) Realpath(path string) string {
	return osVFS.Realpath(path)
}