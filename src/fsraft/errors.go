package fsraft

// This file holds definitions of various file-related errors.
// These are loosely based off the standard POSIX/C error codes available at
// http://www.virtsync.com/c-error-codes-include-errno but with names changed for readability.
// The error numbers here *DO NOT* correspond to the numbers of POSIX error codes!
// For example, InvalidPath (ENOTDIR) is 1 here, but 20 in the official standard.
// This was done so that we would not need to define unused error codes or worry about
// which order errors appear in in the standard.

type Error int

const (
	InvalidPath  Error = iota // A component of the path prefix is not a directory (ENOTDIR).
	NotFound                  // No such file or directory (ENOENT)
	IsDirectory               // The named file is a directory, and the operation is only valid on files. (EISDIR).
	MaxFDsOpen                // The process has already reached its limit for open file descriptors (EMFILE).
	InvalidFD                 // The specified file descriptor is not an active file descriptor (EBADF) or is negative (EINVAL).
	TryAgain                  // Try the operation again, perhaps (though not necessarily) because the Raft leader lost leadership (EAGAIN).
	IOError                   // Something went wrong when trying to read from the hardware (EIO).
	FileTooLarge              // An attempt was made to write a file that exceeds the file size limit (EFBIG).
	NoMoreSpace               // There is no remaining space on the file system containing the file (ENOSPC).
)
