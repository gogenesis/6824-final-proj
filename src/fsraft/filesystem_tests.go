package fsraft

import "testing"
import "fmt"

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
   TestOpenRWClose64,
   TestOpenRWClose,
   TestOpenRWClose4,
   TestOpenRWClose64,
   TestReadWriteBasic,
   TestReadWriteBasic4, //XXX generation marker
}

func HelpOpen(t *testing.T, fs FileSystem,
              path string, mode OpenMode, flags OpenFlags) int {
   fd, err := fs.Open(path, mode, flags)
   assertNoErrorFail(t, err)
   assertValidFD(t, fd)
   return fd
}

func HelpClose(t *testing.T, fs FileSystem, fd int) {
   success, err := fs.Close(fd)
   assertNoErrorFail(t, err)
   assertFail(t, success)
}

func HelpOpenClose(t *testing.T, fs FileSystem,
                      path string, mode OpenMode, flags OpenFlags) {
   HelpClose(t, fs, HelpOpen(t, fs, path, mode, flags))
}

func HelpBatchOpenClose(t *testing.T, fs FileSystem,
                        nFiles int, mode OpenMode, flags OpenFlags) {
   fds := make([]int, nFiles)
   // open N files with same mode and flags
   for ix := 0; ix < nFiles; ix++ {
      fds[ix] = HelpOpen(t, fs, //TODO could randomize name further
                         fmt.Sprintf("/foo%d.txt", ix), mode, flags)
   }
   // then close all N files
   for ix := 0; ix < nFiles; ix++ { HelpClose(t, fs, fds[ix]) }  
}

// ====== END HELPERS ===== BEGIN OPEN CLOSE TESTS ====== 

func TestOpenCloseBasic(t *testing.T, fs FileSystem) {
   HelpOpenClose(t, fs, "/foo.txt", ReadWrite, Create)
}

func TestOpenROClose(t *testing.T, fs FileSystem) {
   HelpOpenClose(t, fs, "/fooRO.txt", ReadOnly, Create)
}

func TestOpenRWClose(t *testing.T, fs FileSystem) {
   HelpOpenClose(t, fs, "/fooRO.txt", ReadOnly, Create)
}

func TestOpenROClose4 (t *testing.T, fs FileSystem) {
   HelpBatchOpenClose(t, fs, 4, ReadOnly, Create)
}

func TestOpenROClose64 (t *testing.T, fs FileSystem) {
   HelpBatchOpenClose(t, fs, 64, ReadOnly, Create)
}

func TestOpenRWClose4 (t *testing.T, fs FileSystem) {
   HelpBatchOpenClose(t, fs, 4, ReadWrite, Create)
}

func TestOpenRWClose64 (t *testing.T, fs FileSystem) {
   HelpBatchOpenClose(t, fs, 64, ReadWrite, Create)
} // holding off on pushing open close more 

// ===== END OPEN CLOSE TESTS ===== BEGIN READ WRITE TESTS =====

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
