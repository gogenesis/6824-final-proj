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
	ad.Debug(ad.TRACE, "Beginning Write(%d, len(data)=%d)", numBytes, len(data))
	if numBytes < 0 {
		ad.Debug(ad.TRACE, "Returning IllegalArgument because numBytes is negative.")
		return -1, filesystem.IllegalArgument
	}
	if file.openMode == filesystem.ReadOnly {
		ad.Debug(ad.TRACE, "Returning WrongMode because this file is open for ReadOnly.")
		return -1, filesystem.WrongMode
	}

	// Two stages: first, make sure the file is long enough by adding zero bytes if necessary.
	bytesWritten = numBytes
	if len(data) < bytesWritten {
		bytesWritten = len(data)
	}
	if file.offset+bytesWritten > len(file.contents) {
		padZeroes := make([]byte, file.offset+bytesWritten-len(file.contents))
		ad.Debug(ad.TRACE, "File offset is at %d and would write %d, but file is only %d bytes long. "+
			"Padding space with %d+%d-%d=%d zero bytes.",
			file.offset, bytesWritten, len(file.contents),
			file.offset, bytesWritten, len(file.contents), len(padZeroes))
		file.contents = append(file.contents, padZeroes...)
	}

	// Then, overwrite bytes without having to worry about length.
	copy(file.contents[file.offset:], data[:bytesWritten])
	file.offset += bytesWritten
	ad.Assert(file.offset <= len(file.contents))
	ad.Debug(ad.TRACE, "Done writing %d bytes, offset now at %d", bytesWritten, file.offset)
	return bytesWritten, nil
}

// See FileSystem::Delete.
func (file *File) Delete() (success bool, err error) {
	return file.inode.Delete()
}

func (file *File) Parent() *Directory {
	return file.inode.Parent()
}
