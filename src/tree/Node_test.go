package tree

import (
	"container/list"
	"testing"
	"time"
)

// Returns a small memoryFS that will be useful for testing.
// The memoryFS is mounted at "/root" and contains one empty file
// in the top-level directory named "foo".
func getSmallTree() MemoryFS {
	tree := createEmptyTree("/root")
	root := tree.GetRootDir()
	root.Open(WriteOnly)
	foo := root.CreateFile("foo")
	foo.Close()
	root.Close()
	return tree
}

// Returns a memoryFS that will be useful for testing.
// The memoryFS is mounted at "/root" and looks like this:
// /root/
//   |- fileA
//   |- child/
//   |    |- grandchild/
//   |    |    |- fileB
//   |    |    |- fileC
// All files are empty.
func getLargeTree() MemoryFS {
	tree := createEmptyTree("/root")

	root := tree.GetRootDir()
	root.Open(WriteOnly)
	fileA := root.CreateFile("fileA")
	fileA.Close()
	child := root.CreateDir("child")
	grandChild := child.CreateDir("grandchild")
	fileB := grandChild.CreateFile("fileB")
	fileB.Close()
	fileC := grandChild.CreateFile("fileC")
	fileC.Close()

	grandChild.Close()
	child.Close()
	root.Close()
	return tree
}

// Create a small memoryFS, open the root dir and the file, and return them all.
func getSmallTreeParts() (MemoryFS, Directory, File) {
	tree := getSmallTree()
	root := tree.GetRootDir()
	root.Open(ReadWrite)
	child := root.GetChildNamed("foo").(File)
	child.Open(ReadWrite)
	return tree, root, child
}

func TestNodeName(t *testing.T) {
	_, root, child := getSmallTreeParts()
	defer root.Close()
	defer child.Close()

	assertEquals("root", root.Name())
	assertEquals("foo", child.Name())
}

func TestRelativePath(t *testing.T) {
	_, root, child := getSmallTreeParts()
	defer root.Close()
	defer child.Close()

	assertEquals("/root", root.RelativePath())
	assertEquals("/root/child", child.Path())
}

func TestCreateTime(t *testing.T) {
	startTime := time.Now()
	_, root, child := getSmallTreeParts()
	endTime := time.Now()
	defer root.Close()
	defer child.Close()
	assert(startTime.Before(root.CreateTime()))
	assert(startTime.Before(child.CreateTime()))
	assert(endTime.After(root.CreateTime()))
	assert(endTime.After(child.CreateTime()))

	startTime = time.Now()
	bigTree := getLargeTree()
	endTime = time.Now()
	for _, node := range bigTree.readAllNodes() {
		assert(startTime.Before(node.CreateTime()))
		assert(endTime.After(node.CreateTime()))
	}
	bigTree.closeAllNodes()
}

func TestGetPointerToTree(t *testing.T) {
	tree, root, child := getSmallTreeParts()
	defer root.Close()
	defer child.Close()
	assertEquals(tree, root.Tree())
	assertEquals(tree, child.Tree())

	bigTree := getLargeTree()
	for _, node := range bigTree.readAllNodes() {
		assertEquals(bigTree, node.Tree())
	}
	bigTree.closeAllNodes()
}

func TestGetParent(t *testing.T) {
	_, smallRoot, child := getSmallTreeParts()
	defer smallRoot.Close()
	defer child.Close()
	assertEquals(smallRoot, child.Parent())

	bigTree := getLargeTree()
	bigTree.readAllNodes()
	defer bigTree.closeAllNodes()
	directoriesToCheck := list.List{} // a doubly-linked list
	directoriesToCheck.PushFront(bigTree.GetRootDir())
	var next *list.Element
	for i := directoriesToCheck.Front(); i != nil; i = next {
		next = i.Next()
		parentDir := i.Value.(Directory)
		for _, childNode := range parentDir.Children() {
			assertEquals(parentDir, childNode.Parent())
			// If child is a directory, add it to the list
			childDir, isDirectory := childNode.(Directory)
			if isDirectory {
				directoriesToCheck.PushBack(childDir)
			}
		}
	}
}
