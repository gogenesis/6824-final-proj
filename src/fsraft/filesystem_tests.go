package fsraft

// Functionality tests for a FileSystem go here.
// Functions in this file are NOT "unit tests" because they ill not be run by "go test" because
// this file ends in "_testS", plural and their argument is not of type testing.T.
// Instead, these functionality tests can run against any class that implements the FileSystem interface
// by creating a unit test suite for your implementation class that calls these tests.

// Whenever you add a new functionality test, be sure to add it to this list.
// This list is used in test_setup.go to run every functionality test on every difficulty.
var FunctionalityTests = []func(fs FileSystem){
	TestFSOpenClose,
	TestFSBasicReadWrite,
}

func TestFSOpenClose(fs FileSystem) {
	fileName := "/foo.txt" // arbitrarily

	fd, err := fs.Open(fileName, ReadWrite, Create)
	assertNoError(err)

	success, err := fs.Close(fd)
	assertNoError(err)
	assert(success)
}

func TestFSBasicReadWrite(fs FileSystem) {
	fileName := "/foo.txt" // arbitrarily
	contents := "bar"      // also arbitrarily
	bytes := []byte(contents)
	numBytes := len(bytes)

	fd, err := fs.Open(fileName, ReadWrite, Create)
	assertNoError(err)

	numWritten, err := fs.Write(fd, numBytes, bytes)
	assertNoError(err)
	assertEquals(numBytes, numWritten)

	newPosition, err := fs.Seek(fd, 0, FromBeginning)
	assertNoError(err)
	assertEquals(0, newPosition)

	numRead, data, err := fs.Read(fd, numBytes)
	assertNoError(err)
	assertEquals(numBytes, numRead)
	assertEquals(bytes, data)

	success, err := fs.Close(fd)
	assertNoError(err)
	assert(success)

}

// TODO more unit tests.
