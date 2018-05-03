package fsraft

// This file holds definitions of various file-related errors.
// These are loosely based off the standard POSIX/C error codes available at
// http://www.virtsync.com/c-error-codes-include-errno but with names changed for readability.
// The error numbers here *DO NOT* correspond to the numbers of POSIX error codes!
// For example, InvalidPath (ENOTDIR) is 1 here, but 20 in the official standard.
// This was done so that we would not need to define unused error codes or worry about
// which order errors appear in in the standard.
// When you add a new ErrorCode, be sure to add it to the ValuesToStrings map!

type ErrorCode int

const (
	InvalidPath       ErrorCode = iota // A component of the path prefix is not a directory (ENOTDIR).
	NotFound                           // No such file or directory (ENOENT)
	IsDirectory                        // The named file is a directory, and the operation is only valid on files. (EISDIR).
	TooManyFDsOpen                     // The process has already reached its limit for open file descriptors (EMFILE).
	InactiveFD                         // The specified file descriptor is not an active file descriptor (EBADF).
	IllegalArgument                    // Some illegal argument, such as an attempt to seek before the beginning of the file. (EINVAL).
	TryAgain                           // Try the operation again, perhaps (though not necessarily) because the Raft leader lost leadership (EAGAIN).
	IOError                            // Something went wrong when trying to interact with the hardware (EIO).
	FileTooLarge                       // An attempt was made to write a file that exceeds the file size limit (EFBIG).
	NoMoreSpace                        // There is no remaining space on the file system containing the file (ENOSPC).
	DirectoryNotEmpty                  // There was an attempt to delete a non-empty directory (ENOTEMPTY).
	AlreadyExists                      // The specified pathname already exists (EEXIST).
	AlreadyOpen                        // An attempt was made to open a file that is already open. This is an error that does not exist in POSIX because a file can only be opened once here.
)

var valuesToStrings = map[ErrorCode]string{
	InvalidPath:       "InvalidPath",
	NotFound:          "NotFound",
	IsDirectory:       "IsDirectory",
	TooManyFDsOpen:    "TooManyFDsOpen",
	InactiveFD:        "InactiveFD",
	IllegalArgument:   "IllegalArgument",
	TryAgain:          "TryAgain",
	IOError:           "IOError",
	FileTooLarge:      "FileTooLarge",
	NoMoreSpace:       "NoMoreSpace",
	DirectoryNotEmpty: "DirectoryNotEmpty",
	AlreadyExists:     "AlreadyExists",
	AlreadyOpen:       "AlreadyOpen",
}

// Needed for ErrorCode to conform to the builtin interface "error",
// Note that ErrorCode uses value receivers, not pointer receivers.
// see https://golang.org/ref/spec#Errors
func (e ErrorCode) Error() string {
	return valuesToStrings[e]
}

// Used for traditionally turning an object into a string.
func (e ErrorCode) String() string {
	return e.Error()
}
