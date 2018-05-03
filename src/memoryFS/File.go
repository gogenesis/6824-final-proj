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
	// TODO
}

// Construct a file by calling createFile(fileName string) in the desired parent directory.

func (file *File) Name() string {
	panic("TODO")
}

// See FileSystem::Open.
func (file *File) Open(mode fsraft.OpenMode, flags fsraft.OpenFlags) (err error) {
	panic("TODO")
}

// See FileSystem::Close.
func (file *File) Close() (success bool, err error) {
	panic("TODO")
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

// Return the offset into the file.
func (file *File) Offset() int {
	panic("TODO")
}


//func (file *File) readLock() {
//	file.Inode.readLock()
//}
//
//func (file *File) readUnlock() {
//	file.Inode.readUnlock()
//}
//
//func (file *File) writeLock() {
//	file.Inode.writeLock()
//}
//
//func (file *File) writeUnlock() {
//	file.Inode.writeUnlock()
//}
