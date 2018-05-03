package memoryFS

import ()

// A directory in a filesystem.
// This class extends Node.
// This data structure is NOT THREADSAFE.
type Directory struct {
	// TODO
}

// Directory has no zero-argument constructor.
// instead, create a root Directory by creating a MemoryFS and then create children of that root Directory.

// Creates a Directory within this one named childName and returns it.
// Panics if there is already a Node named childName in this directory.
// The created Directory is not locked.
func (dir *Directory) CreateDir(childName string) Directory {
	panic("TODO")
}

// Creates a File within this Directory named childName and returns it.
// Panics if there is already a Node named childName in this directory.
// The created File is not opened or locked.
func (dir *Directory) CreateFile(childName string) File {
	panic("TODO")
}

// Get the contents of this directory.
// All children will be of type File or Directory.
func (dir *Directory) Children() []interface{} {
	panic("TODO")
}

// Checks whether a child of this directory is named childName.
func (dir *Directory) HasChildNamed(childName string) bool {
	panic("TODO")
}

// Get the child Node named ChildNode.
// Panics if there is no such child.
// The return value will be of type File or Directory, as appropriate.
func (dir *Directory) GetChildNamed(childName string) interface{} {
	panic("TODO")
}

// From FileSystem::Delete
func (dir *Directory) Delete() {
	panic("TODO")
}

func (dir *Directory) Name() string {
	panic("TODO")
}
