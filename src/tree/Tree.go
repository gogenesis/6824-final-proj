package tree

// A tree of files representing a directory structure.
// This data structure is NOT threadsafe.
type Tree struct{}

// Constructor ================================================================

// Create an empty Tree rooted at rootPath.
// Panics if rootPath is not formatted as an absolute file path.
func createEmptyTree(rootPath string) Tree {
	panic("TODO")
}

// Observer methods ==========================================================

// Return the root directory of the filesystem.
func (tr *Tree) GetRootDir() Directory {
	panic("TODO")
}

// Get the absolute path name of the root directory.
// Returns a string starting with "/" and ending with the name of the root's parent.
// For example, if the sync directory is "/home/kigaltan/sync", returns "/home/kigaltan".
// The ending "/" is not included.
func (tr *Tree) RootPath() string {
	panic("TODO")
}

// Package-private methods ====================================================

// Get all the nodes in the tree, opening them all for reading.
// Should be used only for testing purposes.
// Order not specified.
func (tr *Tree) readAllNodes() []Node {
	panic("TODO")
}

// Close all nodes, panicking if they're not all open.
// Should be used only for testing purposes.
func (tr *Tree) closeAllNodes() {
	panic("TODO")
}
