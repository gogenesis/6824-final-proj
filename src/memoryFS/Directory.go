package memoryFS

import (
	"ad"
	"fmt"
)

// A directory in a filesystem.
// This class extends Node.
// This data structure is NOT THREADSAFE.
type Directory struct {
	inode    Inode
	children map[string]Node
}

// Directory has no public constructor.
// instead, create a root Directory by creating a MemoryFS and then create children of that root Directory.

// See Node::Name.
func (dir *Directory) Name() string {
	return dir.inode.Name()
}

// Creates a Directory within this one named childName and returns it.
// Panics if there is already a Node named childName in this directory.
func (dir *Directory) CreateDir(childName string) *Directory {
	return dir.createChild(childName, true).(*Directory)
}

// Creates a File within this Directory named childName and returns it.
// Panics if there is already a Node named childName in this directory.
func (dir *Directory) CreateFile(childName string) *File {
	return dir.createChild(childName, false).(*File)
}

// Creates a child, either a File if isDirectory is false or a Directory otherwise.
// Panics if there is already a Node named childName in this directory.
func (dir *Directory) createChild(childName string, isDirectory bool) Node {
	if dir.HasChildNamed(childName) {
		panic(fmt.Sprintf("Already has child named %v", childName))
	}
	ad.Debug(ad.TRACE, "Creating child named %v", childName)

	var node Node
	var inode Inode
	// Initialize fields
	inode.name = childName
	inode.parent = dir
	if isDirectory {
		// Create it with type Directory so we can access private field inode temporarily
		dir := &Directory{}
		dir.inode = inode
		dir.children = make(map[string]Node, 0)
		node = dir
	} else {
		file := &File{}
		file.inode = inode
		file.isOpen = false
		node = file
	}

	// Finish up
	dir.children[childName] = node
	return node

}

// Get the contents of this directory.
// Keys are names and values are the Nodes with those names.
// Modifications to the returned map will not change this Directory's actual children.
func (dir *Directory) Children() map[string]Node {
	// Defensive copying needed so that mutations to the returned map don't affect state.
	childrenCopy := make(map[string]Node)
	for name, child := range dir.children {
		childrenCopy[name] = child
	}
	return childrenCopy
}

// Checks whether a child of this directory is named childName.
func (dir *Directory) HasChildNamed(childName string) bool {
	_, hasChild := dir.children[childName]
	return hasChild
}

// Get the child Node named ChildNode.
// Panics if there is no such child.
func (dir *Directory) GetChildNamed(childName string) Node {
	if dir.HasChildNamed(childName) {
		return dir.children[childName]
	} else {
		panic(fmt.Sprintf("I have no child named %v!", childName))
	}
}

// See FileSystem::Delete.
func (dir *Directory) Delete() (success bool, err error) {
	ad.AssertExplain(len(dir.children) == 0, "Cannot delete a non-empty directory!")
	return dir.inode.Delete()
}

func (dir *Directory) Parent() *Directory {
	return dir.inode.Parent()
}
