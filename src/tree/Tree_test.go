package tree

import (
	"fmt"
	"os"
	"testing"
)

func assert(cond bool) {
	if !cond {
		panic("Assertion failed!")
	}
}

func assertEquals(expected, actual interface{}) {
	if !(expected == actual) {
		panic(fmt.Sprintf("AssertionError: expected %v, got %v\n", expected, actual))
	}
}

func TestCreateEmptyTree(t *testing.T) {
	createEmptyTree("/root")
	// If that function call succeeds, that test passed
}

func TestTreeRootPath(t *testing.T) {
	tree := createEmptyTree("/root")
	assertEquals("/root", tree.RootPath())
}

func TestTreeRootDirNameEqualsRootPath(t *testing.T) {
	tree := createEmptyTree("/root")
	assertEquals("/root", tree.RootPath())
	rootDir := tree.GetRootDir()
	rootDir.Open(ReadOnly)
	defer rootDir.Close()
	assertEquals("root", rootDir.Name())
}

func TestTreeRootDirNameEqualsLastDirInRootPath(t *testing.T) {
	tree := createEmptyTree("/foo/bar/baz")
	assertEquals("/foo/bar/baz", tree.RootPath())
	rootDir := tree.GetRootDir()
	rootDir.Open(ReadOnly)
	defer rootDir.Close()
	assertEquals("baz", rootDir.Name())
}

func TestTreeRootDirNameEndsWithSlash(t *testing.T) {
	tree := createEmptyTree("/foo/bar/baz/")
	// no ending slash
	assertEquals("/foo/bar/baz", tree.RootPath())
	rootDir := tree.GetRootDir()
	rootDir.Open(ReadOnly)
	defer rootDir.Close()
	// no ending slash
	assertEquals("baz", rootDir.Name())
}

func TestTreePanicsOnInvalidRootPath(t *testing.T) {
	invalidPathsToMessages := map[string]string{
		"foo": "is not absolute",
	}
	for invalidPath, errorMessage := range invalidPathsToMessages {
		func(invalidPath, errorMessage string) {
			// error expected, so defer recover
			defer func() { recover() }()
			// Should panic, otherwise, fail with the message
			createEmptyTree(invalidPath)
			t.Fatalf(fmt.Sprintf("Able to create a Tree with rootPath=\"%v\", which is problematic because it %v",
				invalidPath, errorMessage))
		}(invalidPath, errorMessage)
	}
}
