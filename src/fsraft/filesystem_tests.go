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
   TestOpenCloseBasic,
   TestOpenROClose,
   TestOpenROClose4,
   TestOpenRWClose,
   TestOpenRWClose4, //XXX generation marker
   TestReadWriteBasic,
   TestReadWriteBasic4,
}

func TestOpenCloseBasic(t *testing.T, fs FileSystem) {
   fileName := "/foo.txt" // arbitrarily

   //@dedup pending
   fd, err := fs.Open(fileName, ReadWrite, Create)
   assertNoErrorFail(t, err)
   assertValidFD(t, fd)

   success, err := fs.Close(fd)
   assertNoErrorFail(t, err)
   assertFail(t, success)
}

func TestOpenROClose(t *testing.T, fs FileSystem) {
   fileName := "/fooRO.txt"

   //@dedup pending
   fd, err := fs.Open(fileName, ReadOnly, Create)
   assertNoErrorFail(t, err)
   assertValidFD(t, fd)

   success, err := fs.Close(fd)
   assertNoErrorFail(t, err)
   assertFail(t, success)
}

func TestOpenROClose4 (t *testing.T, fs FileSystem) {
   path1 := "/fooRO1.txt"
   path2 := "/fooRO2.txt"
   path3 := "/fooRO3.txt"
   path4 := "/fooRO4.txt"

   //@dedup pending
   fd1, err1 := fs.Open(path1, ReadOnly, Create)
   assertNoErrorFail(t, err1)
   assertValidFD(t, fd1)

   fd2, err2 := fs.Open(path2, ReadOnly, Create)
   assertNoErrorFail(t, err2)
   assertValidFD(t, fd2)

   fd3, err3 := fs.Open(path3, ReadOnly, Create)
   assertNoErrorFail(t, err3)
   assertValidFD(t, fd3)

   fd4, err4 := fs.Open(path4, ReadOnly, Create)
   assertNoErrorFail(t, err4)
   assertValidFD(t, fd4)

   //@dedup pending
   success, err := fs.Close(fd1)
   assertNoErrorFail(t, err)
   assertFail(t, success)

   success, err = fs.Close(fd2)
   assertNoErrorFail(t, err)
   assertFail(t, success)

   success, err = fs.Close(fd3)
   assertNoErrorFail(t, err)
   assertFail(t, success)

   success, err = fs.Close(fd4)
   assertNoErrorFail(t, err)
   assertFail(t, success)
}

func TestOpenRWClose(t *testing.T, fs FileSystem) {
   path := "/fooRW.txt"

   //@dedup pending
   fd, err := fs.Open(path, ReadWrite, Create)
   assertNoErrorFail(t, err)
   assertValidFD(t, fd)

   success, err := fs.Close(fd)
   assertNoErrorFail(t, err)
   assertFail(t, success)
}

func TestOpenRWClose4 (t *testing.T, fs FileSystem) {
   path1 := "/fooRW1.txt"
   path2 := "/fooRW2.txt"
   path3 := "/fooRW3.txt"
   path4 := "/fooRW4.txt"

   //@dedup pending
   fd1, err1 := fs.Open(path1, ReadWrite, Create)
   assertNoErrorFail(t, err1)
   assertValidFD(t, fd1)

   fd2, err2 := fs.Open(path2, ReadWrite, Create)
   assertNoErrorFail(t, err2)
   assertValidFD(t, fd2)

   fd3, err3 := fs.Open(path3, ReadWrite, Create)
   assertNoErrorFail(t, err3)
   assertValidFD(t, fd3)

   fd4, err4 := fs.Open(path4, ReadWrite, Create)
   assertNoErrorFail(t, err4)
   assertValidFD(t, fd4)

   //@dedup pending
   success, err := fs.Close(fd1)
   assertNoErrorFail(t, err)
   assertFail(t, success)

   success, err = fs.Close(fd2)
   assertNoErrorFail(t, err)
   assertFail(t, success)

   success, err = fs.Close(fd3)
   assertNoErrorFail(t, err)
   assertFail(t, success)

   success, err = fs.Close(fd4)
   assertNoErrorFail(t, err)
   assertFail(t, success)
}

func HelpReadWrite(t *testing.T, fs FileSystem, path string, contents string) {
   bytes := []byte(contents)
   numBytes := len(bytes)

   fd, err := fs.Open(path, ReadWrite, Create)
   assertNoErrorFail(t, err)
   assertValidFD(t, fd)

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

func TestReadWriteBasic(t *testing.T, fs FileSystem) {
   HelpReadWrite(t, fs, "/foo.txt", "bar") //TODO randomize contents
}

func TestReadWriteBasic4(t *testing.T, fs FileSystem) {
   HelpReadWrite(t, fs, "/foo1.txt", "bar1") //TODO randomize contents
   HelpReadWrite(t, fs, "/foo2.txt", "bar2") //TODO randomize contents
   HelpReadWrite(t, fs, "/foo3.txt", "bar3") //TODO randomize contents
   HelpReadWrite(t, fs, "/foo4.txt", "bar4") //TODO randomize contents
}


// TODO more unit tests.
