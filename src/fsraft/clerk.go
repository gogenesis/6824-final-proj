package fsraft

import "labrpc"

type FSClerk struct {
	// TODO
}

func MakeFsClerk(servers []*labrpc.ClientEnd) *FSClerk {
	panic("TODO")
}

// See the spec for FileSystem::Mkdir.
func (ck *FSClerk) Mkdir(path string) (success bool, err error) {
	panic("TODO")
}

// See the spec for FileSystem::Open.
func (ck *FSClerk) Open(path string, mode OpenMode, flags OpenFlags) (fileDescriptor int, err error) {
	panic("TODO")
}

// See the spec for FileSystem::Close.
func (ck *FSClerk) Close(fileDescriptor int) (success bool, err error) {
	panic("TODO")
}

// See the spec for FileSystem::Seek.
func (ck *FSClerk) Seek(fileDescriptor int, offset int, base SeekMode) (newPosition int, err error) {
	panic("TODO")
}

// See the spec for FileSystem::Read.
func (ck *FSClerk) Read(fileDescriptor int, numBytes int) (bytesRead int, data []byte, err error) {
	panic("TODO")
}

// See the spec for FileSystem::Write.
func (ck *FSClerk) Write(fileDescriptor int, numBytes int, data []byte) (bytesWritten int, err error) {
	panic("TODO")
}

// See the spec for FileSystem::Delete.
func (ck *FSClerk) Delete(path string) (success bool, err error) {
	panic("TODO")
}
