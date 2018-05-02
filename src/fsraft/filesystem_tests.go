package fsraft

// Unit tests for a FileSystem go here.
// Functions in this file are NOT "real tests" that will be run by "go test" because
// this file ends in "_testS", plural and their argument is not of type testing.T.
// Instead, these tests can run against any class that implements the FileSystem interface
// by creating a unit test suite for your implementation class that calls these tests.

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
