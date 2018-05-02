package fsraft

type FileSystem interface {

	// Creates a directory.
	//
	// Path is a relative path name beginning from the top-level synchronized directory and
	// ending in the directory to be created.
	// Possible errors are InvalidPath, TryAgain, IOError, NoMoreSpace, and AlreadyExists.
	// Success is false iff err is non-nil.
	Mkdir(path string) (success bool, err error)

	// Open a file.
	//
	// Path is a relative path name beginning from the top-level synchronized directory.
	// mode is exactly one of {R, W, RW}, and flags can be any number of OpenFlags OR'd together.
	// Returns an integer file descriptor, which is guaranteed to be the lowest file descriptor available.
	// If the file exists, the file is opened and the Create flag is ignored.
	// If the file does not exist and the Create flag is included, creates it.
	// If the file does not exist and the Create flag is not included, returns NotFound error.
	// If the Truncate flag is set, truncates the file size to 0 (if opening succeeds).
	// Possible errors are InvalidPath, IsDirectory, MaxFDsOpen, NotFound, and TryAgain.
	// fileDescriptor == -1 if and only iff err is non-nil.
	Open(path string, mode OpenMode, flags OpenFlags) (fileDescriptor int, err error)

	// Close a file.
	// Possible errors are InvalidFD and TryAgain. Success is false if and only if err is non-nil.
	Close(fileDescriptor int) (success bool, err error)

	// Adjusts the file offset for this file and returns the new offset.
	//
	// If base is FromBeginning, sets the offset to offset bytes.
	// If base is FromCurrent, sets the offset to its current location plus offset.
	// If base is FromEnd, sets the offset to the size of the file plus offset.
	// The seek() function shall allow the file offset to be set beyond the end of the existing data in the file.
	// If data is later written at this point, subsequent reads of data in the gap shall
	// return bytes with the value 0 until data is actually written into the gap.
	// The seek() function shall not, by itself, extend the size of a file.
	// Specification adapted from http://man7.org/linux/man-pages/man2/lseek.2.html.
	// Possible errors are InvalidFD and TryAgain. If err is non-nil, newPosition is unspecified.
	Seek(fileDescriptor int, offset int, base SeekMode) (newPosition int, err error)

	// Attempts to read up to numBytes bytes from a file descriptor.
	//
	// On files that support seeking, the read operation commences at the
	// file offset, and the file offset is incremented by the number of
	// bytes read.  If the file offset is at or past the end of file, no
	// bytes are read, and read() returns zero.
	// Possible errors are IsDirectory, IOError, InvalidFD and TryAgain.
	// If err is non-nil, bytesRead is 0 and data is unspecified.
	Read(fileDescriptor int, numBytes int) (bytesRead int, data []byte, err error)

	// Writes up to numBytes bytes from data to the file referred to by the file descriptor fd.
	//
	// The number of bytes written may be less than numBytes if, for example,
	// there is insufficient space on the underlying physical medium.
	// For a seekable file, writing takes place at the file offset, and
	// the file offset is incremented by the number of bytes actually
	// written.  If the file was opened in Append mode, the file offset is
	// first set to the end of the file before writing.  The adjustment of
	// the file offset and the write operation are performed as an atomic step.
	// Possible errors are IsDirectory, IOError, InvalidFD, TryAgain, FileTooLarge, or NoMoreSpace.
	// If err is non-nil, bytesWritten is -1.
	Write(fileDescriptor int, numBytes int, data []byte) (bytesWritten int, err error)

	// Deletes a name from the filesystem.
	//
	// If the name is a file and no processes have the file open, the file
	// is deleted and the space it was using is made available for reuse.
	// If the name is a file and any processes still have the file open, the file will
	// remain in existence until the last file descriptor referring to it is closed.
	// If the name is a directory, deletes it if it is empty or otherwise returns a DirectoryNotEmpty ErrorCode.
	// Possible errors are InvalidPath, NotFound, DirectoryNotEmpty, TryAgain, or IOError.
	// Success is false if and only if err is non-nil.
	Delete(path string) (success bool, err error)

	// Creates a copy of the file descriptor, using the lowest-numbered unused file descriptor.
	//
	// This function is not yet supported, so the spec is incomplete.
	//func (ck *FSClerk) Duplicate(fileDescriptor int) (newFileDescriptor int, err error) { panic("Not supported.") }
}

type OpenMode int

const (
	ReadOnly OpenMode = iota
	WriteOnly
	ReadWrite
)

type OpenFlags int

const (
	Append OpenFlags = 1 << iota
	Create
	Truncate
)

type SeekMode int

const (
	FromBeginning SeekMode = iota // Seek from the beginning of the file.
	FromCurrent                   // Seek relative to the current position.
	FromEnd                       // Seek after the end of the file.
)