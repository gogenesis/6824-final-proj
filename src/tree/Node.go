package tree

import (
	"fsraft"
	"path"
	"project"
	"time"
)

// A single file or directory in the filesystem.
// This data structure is THREADSAFE.
type Node interface {
	// Observers ==============================================================

	// Get the name of this Node, not including the path.
	Name() string

	// Get the path from the root of the Tree to this Node.
	RelativePath() string

	//// Get the time this Node was last modified.
	// Removed for simplicity (for now).
	//ModifyTime() time.Time

	// Get the time this Node was created.
	CreateTime() time.Time

	// Get a pointer to the Tree containing this Node.
	Tree() Tree

	// Get the Directory containing this node.
	// Returns nil if this folder is the root directory of the Tree.
	Parent() Directory

	// Mutators ===============================================================

	// Delete this Node and, if it is a directory, all Nodes within it.
	// Requires the write lock on this node AND on its parent directory.
	// Panics if this Node is the root directory of its Tree.
	Delete()

	// Changes the name of this Node to name NewName.
	// Requires the lock on this node.
	Rename(newName string)

	// Copy this Node into a new Node named newName.
	// Requires the lock on this node.
	// Panics if another Node in the same directory is also named NewName.
	Copy(newName string)

	// Locking methods ========================================================

	// A Node's locks can be in one of four states:
	// - no locks held by any threads
	// - read lock held by one thread
	// - read lock held by multiple threads
	// - write lock held by one thread
	// No other states are valid.

	// Opens a file for reading, writing, or reading and writing as determined by mode..
	// Unlike the standard system call, cannot create a file.
	// If the file is opened by another goroutine in an incompatible mode, blocks until the file can be opened.
	Open(mode fsraft.OpenMode)

	// Close  an open file. Panics if the file is not open.
	// Note that a file is opened for RW when created, so you must close a file you create.
	Close()

	// Lock this node for reading, preventing others from mutating, but not observing, it.
	// If another goroutine holds a write lock, blocks until they release it.
	// Any number of threads can hold read locks simultaneously.
	readLock()

	// Unlock this node for reading and, if applicable, assert relevant invariants.
	// Panics if you do not hold a read lock.
	readUnlock()

	// Lock this node for writing, preventing others from mutating and observing it.
	// If another gorouting holds a write or read lock, blocks until they release it.
	// Only one thread can hold a write lock at a time.
	writeLock()

	// Unlock this node for writing and, if applicable, assert relevant invariants.
	// Panics if you do not hold the write lock.
	writeUnlock()
}

// WriteLock a node and all of its ancestors, starting by locking the root directory and ending by locking the Node itself.
func WriteLockNodeAndAncestors(node Node) {
	panic("TODO")
}

// WriteUnlock a node and all of its ancestors, starting by unlocking the node and ending by unlocking the root directory.
func WriteUnlockNodeAndAncestors(node Node) {
	panic("TODO")
}

// Get the absolute path name of this Node, starting with the root of the entire filesystem.
func AbsolutePath(node Node) string {
	panic("TODO")
	//node.readLock()
	//defer node.readUnlock()
	//
	//return path.Join(node.Tree().rootPath(), node.RelativePath())
}
