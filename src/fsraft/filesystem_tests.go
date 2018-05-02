package fsraft

import "testing"

// Functionality tests for a FileSystem go here.
// Functions in this file are NOT "unit tests" because they ill not be run by "go test" because
// this file ends in "_testS", plural and they have more than one argument.
// Instead, these functionality tests can run against any class that implements the FileSystem interface
// by creating a unit test suite for your implementation class that calls these tests.

// Whenever you add a new functionality test, be sure to add it to this list.
// This list is used in test_setup.go to run every functionality test on every difficulty.
var FunctionalityTests = []func(t *testing.T, fs FileSystem){
	TestOpenClose,
	TestBasicReadWrite,
}

func TestOpenClose(t *testing.T, fs FileSystem) {
	fileName := "/foo.txt" // arbitrarily

	fd, err := fs.Open(fileName, ReadWrite, Create)
	assertNoErrorFail(t, err)

	success, err := fs.Close(fd)
	assertNoErrorFail(t, err)
	assertFail(t, success)
}

func TestBasicReadWrite(t *testing.T, fs FileSystem) {
	fileName := "/foo.txt" // arbitrarily
	contents := "bar"      // also arbitrarily
	bytes := []byte(contents)
	numBytes := len(bytes)

	fd, err := fs.Open(fileName, ReadWrite, Create)
	assertNoErrorFail(t, err)

	numWritten, err := fs.Write(fd, numBytes, bytes)
	assertNoErrorFail(t, err)
	assertEqualsFail(t, numBytes, numWritten)

	newPosition, err := fs.Seek(fd, 0, FromBeginning)
	assertNoErrorFail(t, err)
	assertEqualsFail(t, 0, newPosition)

	numRead, data, err := fs.Read(fd, numBytes)
	assertNoErrorFail(t, err)
	assertEqualsFail(t, numBytes, numRead)
	assertEqualsFail(t, bytes, data)

	success, err := fs.Close(fd)
	assertNoErrorFail(t, err)
	assertFail(t, success)

}

// TODO more unit tests.
