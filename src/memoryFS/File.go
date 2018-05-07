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
	inode    Inode
	isOpen   bool
	openMode fsraft.OpenMode
	contents []byte
	offset   int // Invariant: offset >= 0
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
	// currently supporting FromBeginning
	file.offset = offset
	return file.offset, nil
}

// See FileSystem::Read.
func (file *File) Read(numBytes int) (bytesRead int, data []byte, err error) {
	readBytes := make([]byte, numBytes)
	if numBytes < 0 {
		return -1, make([]byte, 0), fsraft.IllegalArgument
	}
	//  DEBUG CODE stashed ... please remind me how we do leveled logs
	print("offset")
	print(file.offset)
	print("\n")
	print("numBytes")
	print(numBytes)
	print("\n")
	print("lenfile")
	print(len(file.contents))
	print("\n")
	//TODO if directory, return IsDirectory
	if file.offset+bytesRead > len(file.contents) { //if offset goes past end of file
		return 0, make([]byte, 0), nil //no bytes are read
	}
	copy(readBytes, file.contents[file.offset:file.offset+numBytes])
	file.offset += numBytes
	print("offset now")
	print(file.offset)
	print("\n")
	return numBytes, readBytes, nil
}

// See FileSystem::Write.
func (file *File) Write(numBytes int, data []byte) (bytesWritten int, err error) {
	// when we have assert, assert numBytes == len(data) to catch tester bugs
	// grow file as needed, leaving holes >EOF written
	if file.offset+numBytes > len(file.contents) {
		// DEBUG CODE stashed ... please remind me how we do leveled logs
		print("offset")
		print(file.offset)
		print("\n")
		print("numBytes")
		print(numBytes)
		print("\n")
		print("lenfile")
		print(len(file.contents))
		print("\nGROWING\n")
		realloc := make([]byte, file.offset+numBytes)
		copy(realloc[0:len(file.contents)], file.contents)
		file.contents = realloc //garbage collect old contents but need to confirm
		// DEBUG CODE
		print("lenfile now")
		print(len(file.contents))
		print("\n")
	}
	copy(file.contents[file.offset:file.offset+numBytes], data)
	// we currently assume all bytes are written correctly
	// more strict checks would check datastore space first and write up to limit
	file.offset += numBytes
	print("offset now")
	print(file.offset)
	print("\n")
	return numBytes, nil
}

// See FileSystem::Delete.
func (file *File) Delete() (success bool, err error) {
	return file.inode.Delete()
}
