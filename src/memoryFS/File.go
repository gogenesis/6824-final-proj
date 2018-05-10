package memoryFS

import (
	"ad"
	"filesystem"
)

// A file (not a directory) in a filesystem.
// This class extends Node.
// This data structure is NOT THREADSAFE.
// All methods require a file to be Open()ed before calling (except Open(), obviously)
// and panic when called on a non-open file.
type File struct {
	inode    Inode
	isOpen   bool
	openMode filesystem.OpenMode
	contents []byte
	offset   int // Invariant: offset >= 0
}

// Construct a file by calling createFile(fileName string) in the desired parent directory.

// See Node::Name.
func (file *File) Name() string {
	return file.inode.Name()
}

// See FileSystem::Open.
func (file *File) Open(mode filesystem.OpenMode, flags filesystem.OpenFlags) (err error) {
	if file.isOpen {
		return filesystem.AlreadyOpen
	}
	file.isOpen = true
	file.openMode = mode
	if filesystem.FlagIsSet(flags, filesystem.Truncate) {
		file.contents = make([]byte, 0)
		file.offset = 0
	}
	if filesystem.FlagIsSet(flags, filesystem.Append) {
		file.offset = len(file.contents)
	}
	return nil
}

// See FileSystem::Close.
func (file *File) Close() (success bool, err error) {
	ad.AssertExplain(file.isOpen, "Attempted to close a closed file! "+
		"This should never happen because you need a fd to close a file.")
	file.isOpen = false
	return true, nil
}

// See FileSystem::Seek.
func (file *File) Seek(offset int, base filesystem.SeekMode) (newPosition int, err error) {
	prevOffset := file.offset // In case we need to rollback the operation.

	switch base {
	case filesystem.FromBeginning:
		file.offset = offset
	case filesystem.FromCurrent:
		file.offset += offset
	case filesystem.FromEnd:
		file.offset = len(file.contents) + offset
	}

	if file.offset < 0 {
		// Cannot have offset before the beginning of the file, so revert
		file.offset = prevOffset
		return -1, filesystem.IllegalArgument
	}

	return file.offset, nil
}

// See FileSystem::Read.
func (file *File) Read(numBytes int) (bytesRead int, data []byte, err error) {
	if numBytes < 0 {
		return -1, nil, filesystem.IllegalArgument
	}
	if file.openMode == filesystem.WriteOnly {
		return -1, nil, filesystem.WrongMode
	}
	if numBytes == 0 || file.offset >= len(file.contents) {
		// This is specified to be a no-op.
		return 0, make([]byte, 0), nil
	}

	if file.offset+numBytes <= len(file.contents) {
		// We can read numBytes without hitting the end of the file.
		bytesRead = numBytes
		data = make([]byte, bytesRead)
		copy(data, file.contents[file.offset:file.offset+numBytes])
		file.offset += bytesRead
	} else {
		// We can only read up to the end of the file.
		bytesRead = len(file.contents) - file.offset
		data = make([]byte, bytesRead)
		copy(data, file.contents[file.offset:])
		file.offset = len(file.contents)
	}

	return bytesRead, data, nil
}

// See FileSystem::Write.
func (file *File) Write(numBytes int, data []byte) (bytesWritten int, err error) {
	if file.openMode == filesystem.ReadOnly {
		return -1, filesystem.WrongMode
	}
	ad.AssertExplain(numBytes == len(data), "bad numBytes %d vs len(data) %d",
		numBytes, len(data))
	// grow file as needed, leaving holes >EOF written
	if file.offset+numBytes > len(file.contents) {
		ad.Debug(ad.RPC, "growing file - offset %d numBytes %d len(contents) %d",
			file.offset, numBytes, len(file.contents))
		realloc := make([]byte, file.offset+numBytes)
		copy(realloc[0:len(file.contents)], file.contents)
		file.contents = realloc //hopefully garbage collect old contents, needs confirm
	}
	copy(file.contents[file.offset:file.offset+numBytes], data)
	// we currently assume all bytes are written correctly
	// more strict checks would check datastore space first and write up to limit
	file.offset += numBytes
	ad.Debug(ad.RPC, "done, seek offset %d, file now %d bytes", file.offset, len(file.contents))
	return numBytes, nil
}

// See FileSystem::Delete.
func (file *File) Delete() (success bool, err error) {
	return file.inode.Delete()
}
