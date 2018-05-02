package tree

import "time"

// A directory in a filesystem.
// This class extends Node.
// This data structure is threadsafe.
type Directory struct {
	Inode Inode
}

// Constructors ================================================================

// Directory has no zero-argument constructor.
// instead, create a root Directory by creating a MemoryFS and then create children of that root Directory.

// Creates a Directory within this one named childName and returns it.
// Panics if there is already a Node named childName in this directory.
// The created Directory is opened for RW so you can put things inside of it.
func (dir *Directory) CreateDir(childName string) Directory {
	panic("TODO")
}

// Creates a File within this Directory named childName and returns it.
// Panics if there is already a Node named childName in this directory.
// The created File is opened for RW so you can put things inside of it -
// if you don't want to have to separately close it, use Touch instead.
func (dir *Directory) CreateFile(childName string) File {
	panic("TODO")
}

// Observers ===================================================================

// Get the contents of this directory.
// Note that the contents are guaranteed not to change only as long as you hold the read lock;
// if you open the directory for reading, store the contents into an array, and close the
// directory, the contents may change and be different from your array.
func (dir *Directory) Children() []Node {
	panic("TODO")
}

// Checks whether a child of this directory is named childName.
func (dir *Directory) HasChildNamed(childName string) bool {
	panic("TODO")
}

// Get the child Node named ChildNode.
// Panics if there is no such child.
func (dir *Directory) GetChildNamed(childName string) (child Node, err error) {
	panic("TODO")
}

// Copied from Node ===========================================================

func (dir *Directory) Name() string {
	return dir.Inode.Name()
}

func (dir *Directory) RelativePath() string {
	return dir.Inode.RelativePath()
}

func (dir *Directory) CreateTime() time.Time {
	return dir.Inode.CreateTime()
}

func (dir *Directory) Tree() MemoryFS {
	return dir.Inode.Tree()
}

func (dir *Directory) Parent() Directory {
	return dir.Inode.Parent()
}

func (dir *Directory) Delete() {
	dir.Inode.Delete()
}

func (dir *Directory) Rename(newName string) {
	dir.Inode.Rename(newName)
}

func (dir *Directory) Copy(newName string) {
	dir.Inode.Copy(newName)
}

func (dir *Directory) Open(mode OpenMode) {
	dir.Inode.Open(mode)
}

func (dir *Directory) Close() {
	dir.Inode.Close()
}

func (dir *Directory) readLock() {
	dir.Inode.readLock()
}

func (dir *Directory) readUnlock() {
	dir.Inode.readUnlock()
}

func (dir *Directory) writeLock() {
	dir.Inode.writeLock()
}

func (dir *Directory) writeUnlock() {
	dir.Inode.writeUnlock()
}
