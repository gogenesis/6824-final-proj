package tree

import (
	"fsraft"
	"time"
)

// A file (not a directory) in a filesystem.
// This class extends Node.
// This data structure is THREADSAFE.
// All methods require a file to be Open()ed before calling (except Open(), obviously)
// and panic when called on a non-open file.
type File struct {
	Inode Inode
}

// Constructors ===============================================================

// Construct a file by calling createFile(fileName string) in the desired parent directory.

// Observers ==================================================================

// Return the size of this file, in bytes.
func (file *File) size() int {
	panic("TODO")
}

// Reads up to count bytes into the target byte[] and returns the number of bytes read.
// The bytes read is the greatest of {count, len(target), the number of bytes between offset and the end of the file}.
// Also increases the offset by the number of bytes read.
// Reads commence at the file's current offset, which can be adjusted with lseek().
// Spec adapted from http://man7.org/linux/man-pages/man2/read.2.html.
func (file *File) Read(target []byte, count int) int {
	panic("TODO")
}

// Return the offset into the file.
func (file *File) Offset() int {
	panic("TODO")
}

// Mutators ===================================================================
// Adjusts the file offset for this file and returns the new offset.
// If base is AfterBeginning, sets the offset to offset bytes.
// If base is AfterCurrent, sets the offset to its current location plus offset.
// If base is AfterEnd, sets the offset to the size of the file plus offset.
// The seek() function shall allow the file offset to be set beyond the end of the existing data in the file.
// If data is later written at this point, subsequent reads of data in the gap shall
// return bytes with the value 0 until data is actually written into the gap.
// The seek() function shall not, by itself, extend the size of a file.
// Specification adapted from http://man7.org/linux/man-pages/man2/lseek.2.html.
func (file *File) Lseek(offset int, base fsraft.SeekMode) int {
	panic("TODO")
}

// Writes up to count bytes from the source byte[] and returns the number of bytes written.
// The bytes written is usually the greatest of {count, len(source)}, though it may be less if, for example, the
// underlying hardware runs out of space. Increases the file offset by the number of bytes written.
// Writes commence at the file's current offset, which can be adjusted with lseek().
// Spec adapted from http://man7.org/linux/man-pages/man2/write.2.html.
func (file *File) Write(source []byte, count int) int {
	panic("TODO")
}

// Copied from Node ===========================================================

func (file *File) Name() string {
	return file.Inode.Name()
}

func (file *File) RelativePath() string {
	return file.Inode.RelativePath()
}

func (file *File) CreateTime() time.Time {
	return file.Inode.CreateTime()
}

func (file *File) Tree() Tree {
	return file.Inode.Tree()
}

func (file *File) Parent() Directory {
	return file.Inode.Parent()
}

func (file *File) Delete() {
	file.Inode.Delete()
}

func (file *File) Rename(newName string) {
	file.Inode.Rename(newName)
}

func (file *File) Copy(newName string) {
	file.Inode.Copy(newName)
}

func (file *File) Open(mode OpenMode) {
	file.Inode.Open(mode)
}

func (file *File) Close() {
	file.Inode.Close()
}

func (file *File) readLock() {
	file.Inode.readLock()
}

func (file *File) readUnlock() {
	file.Inode.readUnlock()
}

func (file *File) writeLock() {
	file.Inode.writeLock()
}

func (file *File) writeUnlock() {
	file.Inode.writeUnlock()
}
