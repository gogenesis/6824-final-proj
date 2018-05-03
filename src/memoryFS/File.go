package memoryFS

import (
	"fsraft"
)

// A file (not a directory) in a filesystem.
// This class extends Node.
// This data structure is NOT THREADSAFE.
// All methods require a file to be Open()ed before calling (except Open(), obviously)
// and panic when called on a non-open file.
type File struct {
	inode Inode
	isOpen bool
	openMode fsraft.OpenMode
	contents []byte
	offset int // Invariant: offset >= 0
}

// Construct a file by calling createFile(fileName string) in the desired parent directory.

// See Node::Name.
func (file *File) Name() string {
	return file.inode.Name()
}

// See FileSystem::Open.
func (file *File) Open(mode fsraft.OpenMode, flags fsraft.OpenFlags) (err error) {
	if file.isOpen {
		return fsraft.AlreadyOpen
	}
	file.isOpen = true
	file.openMode = mode
	if fsraft.FlagIsSet(flags, fsraft.Truncate) {
		file.contents = make([]byte, 0)
		file.offset = 0
	}
	if fsraft.FlagIsSet(flags, fsraft.Append) {
		file.offset = len(file.contents)
	}
	return nil
}

// See FileSystem::Close.
func (file *File) Close() (success bool, err error) {
	if !file.isOpen {
		panic("Attempted to close a closed file! This should never happen because you need a fd to close a file.")
	}
	file.isOpen = false
	return true, nil
}

// See FileSystem::Seek.
func (file *File) Seek(offset int, base fsraft.SeekMode) (newPosition int, err error) {
	panic("TODO")
}

// See FileSystem::Read.
func (file *File) Read(numBytes int) (bytesRead int, data []byte, err error) {
	panic("TODO")
}

// See FileSystem::Write.
func (file *File) Write(numBytes int, data []byte) (bytesWritten int, err error) {
	panic("TODO")
}

// See FileSystem::Delete.
func (file *File) Delete() (success bool, err error) {
	panic("TODO")
	// Don't forget to call inode Delete (but other code will be necessary)
}
