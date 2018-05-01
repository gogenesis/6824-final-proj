package fsraft

// This file stores various enums used in the FileServer API,
// so as to not clutter fs_server.go with many enum definitions.

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
	AfterBeginning SeekMode = iota // Seek from the beginning of the file.
	AfterCurrent                   // Seek relative to the current position.
	AfterEnd                       // Seek after the end of the file.NotOpen
)
