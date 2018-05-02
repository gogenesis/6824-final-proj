package memoryFS

import "fsraft"

// An in-memory file system.
// Files here are stored in memory in Go.
type MemoryFS struct {
	openFDs map[int]interface{} // TODO this should be a map to files
	// TODO more
}

// Create an empty in-memory FileSystem rooted at "/".
func CreateEmptyMemoryFS() MemoryFS {
	panic("TODO")
}

// See the spec for FileSystem::Mkdir.
func (mfs *MemoryFS) Mkdir(path string) (success bool, err error) {
	panic("TODO")
}

// See the spec for FileSystem::Open.
func (mfs *MemoryFS) Open(path string, mode fsraft.OpenMode, flags fsraft.OpenFlags) (fileDescriptor int, err error) {
	panic("TODO")
}

// See the spec for FileSystem::Close.
func (mfs *MemoryFS) Close(fileDescriptor int) (success bool, err error) {
	panic("TODO")
}

// See the spec for FileSystem::Seek.
func (mfs *MemoryFS) Seek(fileDescriptor int, offset int, base fsraft.SeekMode) (newPosition int, err error) {
	panic("TODO")
}

// See the spec for FileSystem::Read.
func (mfs *MemoryFS) Read(fileDescriptor int, numBytes int) (bytesRead int, data []byte, err error) {
	panic("TODO")
}

// See the spec for FileSystem::Write.
func (mfs *MemoryFS) Write(fileDescriptor int, numBytes int, data []byte) (bytesWritten int, err error) {
	panic("TODO")
}

// See the spec for FileSystem::Delete.
func (mfs *MemoryFS) Delete(path string) (success bool, err error) {
	panic("TODO")
}
