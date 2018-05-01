package fsraft

import "labrpc"

type FSClerk struct {
	// TODO
}

func MakeFsClerk(servers []*labrpc.ClientEnd) *FSClerk {
	panic("TODO")
}

// Open a file.
// Path is a relative path name beginning from the top-level synchronized directory.
// mode is exactly one of {R, W, RW}, and flags can be any number of OpenFlags OR'd together.
// Returns an integer file descriptor, which is guaranteed to be the lowest file descriptor available.
// If the file exists, the file is opened and the Create flag is ignored.
// If the file does not exist and the Create flag is included, creates it.
// If the file does not exist and the Create flag is not included, returns NotFound error.
// If the Truncate flag is set, truncates the file size to 0 (if opening succeeds).
// Possible errors are InvalidPath, IsDirectory, MaxFDsOpen, NotFound, and TryAgain. fileDescriptor is -1 iff err is non-nil.
func (ck *FSClerk) Open(path string, mode OpenMode, flags OpenFlags) (fileDescriptor int, err error) {
	panic("TODO")
}

// Close a file.
// Possible errors are InvalidFD and TryAgain.
func (ck *FSClerk) Close(fileDescriptor int) (success bool, err error) {
	panic("TODO")
}

// Adjusts the file offset for this file and returns the new offset.
// If base is FromBeginning, sets the offset to offset bytes.
// If base is FromCurrent, sets the offset to its current location plus offset.
// If base is FromEnd, sets the offset to the size of the file plus offset.
// The seek() function shall allow the file offset to be set beyond the end of the existing data in the file.
// If data is later written at this point, subsequent reads of data in the gap shall
// return bytes with the value 0 until data is actually written into the gap.
// The seek() function shall not, by itself, extend the size of a file.
// Specification adapted from http://man7.org/linux/man-pages/man2/lseek.2.html.
// Possible errors are InvalidFD and TryAgain.
func (ck *FSClerk) Seek(fileDescriptor int, offset int, base SeekMode) (newPosition int, err error) {
	panic("TODO")
}

// Attempts to read up to numBytes bytes from a file descriptor.
// On files that support seeking, the read operation commences at the
// file offset, and the file offset is incremented by the number of
// bytes read.  If the file offset is at or past the end of file, no
// bytes are read, and read() returns zero.
// Possible errors are IsDirectory, IOError, InvalidFD and TryAgain.
func (ck *FSClerk) Read(fileDescriptor int, numBytes int) (bytesRead int, data []byte, err error) {
	panic("TODO")
}

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
func (ck *FSClerk) Write(fileDescriptor int, numBytes int, data []byte) (bytesWritten int, err error) {
	panic("TODO")
}

// Creates a copy of the file descriptor, using the lowest-numbered unused file descriptor.
// This function is not yet supported, so the spec is incomplete.
func (ck *FSClerk) Duplicate(fileDescriptor int) (newFileDescriptor int, err error) {
	panic("Not supported.")
}
