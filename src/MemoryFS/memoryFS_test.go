package MemoryFS

import (
	"fsraft"
	"testing"
)

func TestMemoryFS_OpenClose(t *testing.T) {
	mfs := CreateEmptyMemoryFS()
	fsraft.TestFSOpenClose(&mfs)
}

func TestMemoryFS_BasicReadWrite(t *testing.T) {
	mfs := CreateEmptyMemoryFS()
	fsraft.TestFSBasicReadWrite(&mfs)
}

// TODO when more tests are written in filesystem_tests.go, call them here.
