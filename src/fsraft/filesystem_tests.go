package fsraft

import "testing"
import "fmt"

// Functionality tests for a FileSystem go here.
// Functions in this file are NOT "unit tests" because they ill not be run by "go test" because
// this file ends in "_testS", plural and they have more than one argument.
// Instead, these functionality tests can run against any class that implements the FileSystem interface
// by creating a unit test suite for your implementation class that calls these tests.

// ===== BEGIN OPEN CLOSE DELETE HELPERS =====

func HelpDelete(t *testing.T, fs FileSystem,
   pathname string) {
   success, err := fs.Delete(pathname)
   assertNoErrorFail(t, err)
   assertEqualsFail(t, success, true)
   assertExplain(t, success, "err deleting file %s", pathname)
}

func HelpOpen(t *testing.T, fs FileSystem,
	path string, mode OpenMode, flags OpenFlags) int {
	fd, err := fs.Open(path, mode, flags)
	assertNoErrorFail(t, err)
	assertValidFD(t, fd)
   //assertExplain(t, err != nil, "err opening fd %d", fd) 
	return fd
}

func HelpClose(t *testing.T, fs FileSystem,
   fd int) {
	success, err := fs.Close(fd)
	assertNoErrorFail(t, err)
	assertFail(t, success)
   assertExplain(t, success, "err closing fd %d", fd)
}

func HelpOpenClose(t *testing.T, fs FileSystem,
	path string, mode OpenMode, flags OpenFlags) {
	HelpClose(t, fs, HelpOpen(t, fs, path, mode, flags))
}

//@sync naming convention with HelpBatchDelete
func HelpBatchOpen(t *testing.T, fs FileSystem,
   nFiles int, fmtStr string, mode OpenMode, flags OpenFlags) []int {
   fds := make([]int, nFiles)
	// open N files with same mode and flags
	for ix := 0; ix < nFiles; ix++ {
		fds[ix] = HelpOpen(t, fs, fmt.Sprintf(fmtStr, ix), mode, flags)
	}
   return fds
}

func HelpBatchCloseFDs(t *testing.T, fs FileSystem,
   fds []int) {
   for ix := 0; ix < len(fds); ix++ {
		HelpClose(t, fs, fds[ix])
	}
}

//@sync naming convention with HelpBatchDelete
func HelpBatchOpenClose(t *testing.T, fs FileSystem,
   nFiles int, fmtStr string, mode OpenMode, flags OpenFlags) {
	fds := HelpBatchOpen(t, fs, nFiles, fmtStr, mode, flags)
   HelpBatchCloseFDs(t, fs, fds)
}

//@sync naming convention with HelpBatchOpenClose
func HelpBatchDelete(t *testing.T, fs FileSystem,
   nFiles int, fmtStr string) {
	// delete all files
	for ix := 0; ix < nFiles; ix++ {
		HelpDelete(t, fs, fmt.Sprintf(fmtStr, ix))
	}
}
// ====== END OPEN CLOSE DELETE HELPERS ===== 

// ===== BEGIN MKDIR HELPERS ===== 

func HelpMkdir(t *testing.T, fs FileSystem,
   path string) {
   success, err := fs.Mkdir(path)
   assertNoErrorFail(t, err)
	assertFail(t, success)
   assertExplain(t, success, "mkdir fail on %s", path)
}

// ===== END MKDIR HELPERS =====

// ===== BEGIN READ WRITE SEEK HELPERS =====

// error checked helper
func HelpSeek(t *testing.T, fs FileSystem,
   fd int, offset int, mode SeekMode) (int) {
   newPosition, err := fs.Seek(fd, offset, mode)
	assertNoErrorFail(t, err)
   if mode == FromBeginning {
      assertEqualsFail(t, offset, newPosition)
   } // can we auto-check more seek behavior...
   return newPosition
}
// error checked helper
func HelpRead(t *testing.T, fs FileSystem,
   fd int, contents string, numBytes int) (int, []byte) {
	numRead, data, err := fs.Read(fd, numBytes)
	assertNoErrorFail(t, err)
	assertEqualsFail(t, numBytes, numRead)
	assertEqualsFail(t, contents, data)
   return numRead, data
}
// error checked helper
func HelpWrite(t *testing.T, fs FileSystem,
   fd int, contents string) int {
   bytes := []byte(contents)
	numBytes := len(bytes)
	numWritten, err := fs.Write(fd, numBytes, bytes)
	assertNoErrorFail(t, err)
	assertEqualsFail(t, numBytes, numWritten)
   return numWritten
}
// error checked helper
func HelpReadWrite(t *testing.T, fs FileSystem,
   path string, contents string) (int) {
   fd := HelpOpen(t, fs, path, ReadWrite, Create)
   HelpSeek(t, fs, fd, 0, FromBeginning)
   nBytes := HelpWrite(t, fs, fd, contents)
   assertExplain(t, nBytes == len(contents),
                 "%d bytes written vs %d", nBytes, len(contents))
   nBytes, data := HelpRead(t, fs, fd, contents, len(contents))
   for bite := 0; bite < len(contents); bite++ {
      assertExplain(t, data[bite] == contents[bite],
                    "read data %s vs %s", data[bite], contents[bite])
   }
   return nBytes
}

// Whenever you add a new functionality test, be sure to add it to this list.
// This list is used in test_setup.go to run every functionality test on every difficulty.
//
// "It will be a big beautiful list..."
// "... the huuugest list... and the students are gonna list it."
//
var FunctionalityTests = []func(t *testing.T, fs FileSystem){
	TestOpenCloseBasic,
	TestOpenROClose,
	TestOpenRWClose,
	TestOpenROClose4,
	TestOpenRWClose64,
	TestOpenRWClose4,
	TestOpenRWClose64,
   TestOpenCloseLeastFD,
   TestOpenCloseDeleteFD128,
   TestOpenCloseDeleteAcrossDirectories,
	TestReadWriteBasic,
	TestReadWriteBasic4,
   TestSeekErrorBadFD,
   TestSeekErrorBadOffsetOperation,
   TestSeekErrorBadOffset1,
   TestMkdir,
   TestMkdirTree,
   //many more to come after milestone 1
   //and many more if we still aim to support a real linux driver
}

// ===== BEGIN OPEN CLOSE TESTS ======

func TestOpenCloseBasic(t *testing.T, fs FileSystem) {
	HelpOpenClose(t, fs, "/foo.txt", ReadWrite, Create)
}

// Do we deal with RW / RO writing issues or are those the perimissions we are
// ignoring... bc this could collapse 
func TestOpenROClose(t *testing.T, fs FileSystem) {
	HelpOpenClose(t, fs, "/fooRO.txt", ReadOnly, Create)
}

func TestOpenRWClose(t *testing.T, fs FileSystem) {
	HelpOpenClose(t, fs, "/fooRO.txt", ReadOnly, Create)
}

func TestOpenROClose4(t *testing.T, fs FileSystem) {
	HelpBatchOpenClose(t, fs, 4, "/a-str-with-a-%d", ReadOnly, Create)
}

func TestOpenROClose64(t *testing.T, fs FileSystem) {
	HelpBatchOpenClose(t, fs, 64, "/str-2-with-a-%d", ReadOnly, Create)
}

func TestOpenRWClose4(t *testing.T, fs FileSystem) {
	HelpBatchOpenClose(t, fs, 4, "/str-3-with-a-%d", ReadWrite, Create)
}

func TestOpenRWClose64(t *testing.T, fs FileSystem) {
	HelpBatchOpenClose(t, fs, 64, "/str-4-with-a-%d", ReadWrite, Create)
} // holding off on pushing open close more

func TestOpenCloseLeastFD(t *testing.T, fs FileSystem) {
	fd3A := HelpOpen(t, fs, "/A.txt", ReadWrite, Create)
	// Should be 3 because that's the lowest non-reserved non-active FD.
	assertEqualsFail(t, 3, fd3A)
	HelpClose(t, fs, fd3A)

	fd3B := HelpOpen(t, fs, "/B.txt", ReadWrite, Create)
	// Should be 3 again because A.txt was closed, so FD=3 is now non-active again.
	assertEqualsFail(t, 3, fd3B)
	// we're not closing it just yet

	fd4 := HelpOpen(t, fs, "/C.txt", ReadWrite, Create)
	// Should be 4 because 0-2 are reserved, 3 is taken, and 4 is next.
	assertEqualsFail(t, 4, fd4)

	HelpClose(t, fs, fd3B)

	fd3C := HelpOpen(t, fs, "/D.txt", ReadWrite, Create)
	// B.txt was closed, so FD=3 is now non-active again.
	assertEqualsFail(t, 3, fd3C)

	HelpClose(t, fs, fd3C)
	HelpClose(t, fs, fd4)
}

// XXX PLZ PING ME BEFORE BELIEVING ANY FAILURES IN THE BELOW TESTS RIGHT NOW!
// There are some semmantics to discuss to make sure the tests are aligned
// with the guarantees of our system, and test and dev should stay coupled
// and progress together. Consider this a sync point where I want to
// re-validate the below tests before throwing a bunch of potentially bogus
// issues... they should be mostly good but want to look again tomorrow.

// open and close files checking 128 FD limit that fd is always increasing
func TestOpenCloseDeleteFD128(t *testing.T, fs FileSystem) {
   prevFD := 2
   //@dedup use something like HelpBatchOpen ?
   for ix := 0; ix < 128; ix++ {
      fd := HelpOpen(t, fs, fmt.Sprintf("/least-fd-%d.txt", ix),
                     ReadWrite, Create)
      assertEqualsFail(t, fd > prevFD, true)
      prevFD = fd
   }
   //
   // please use assertExplain() to give helpful context to the failure
   //
   // my test system auto-dedups bugs if continuous testing hits dup failures
   // which facilitates rapid bug shakedowns to see the overall system health
   //
   assertExplain(t, prevFD == 127, "wanted first 127 but ended with %d", prevFD)

   //@dedup probably going to need a batch close and delete helper...
   for ix := 0; ix < 128; ix++ {
      HelpClose(t, fs, ix)
      HelpDelete(t, fs, fmt.Sprintf("/least-fd-%d.txt", ix))
   }
}

//TODO coming next is one that hits MaxFDsOpen...

func TestOpenCloseDeleteAcrossDirectories(t *testing.T, fs FileSystem) {
   HelpMkdir(t, fs, "/dir1")
   HelpMkdir(t, fs, "/dir2")
   HelpMkdir(t, fs, "/dir3")
   fd1 := HelpOpen(t, fs, "/dir1/foo", ReadWrite, Create)
   fd2 := HelpOpen(t, fs, "/dir2/bar", ReadWrite, Create)
   fd3 := HelpOpen(t, fs, "/dir3/baz", ReadWrite, Create)
   HelpClose(t, fs, fd1)
   HelpClose(t, fs, fd2)
   HelpClose(t, fs, fd3)
   HelpDelete(t, fs, "/dir1/foo")
   HelpDelete(t, fs, "/dir2/bar")
   HelpDelete(t, fs, "/dir3/baz")
   HelpDelete(t, fs, "/dir1")
   HelpDelete(t, fs, "/dir2")
   HelpDelete(t, fs, "/dir3")
}

//TODO larger trees coming soon...

// ===== END OPEN CLOSE TESTS =====

// ===== BEGIN READ WRITE TESTS =====

func TestReadWriteBasic(t *testing.T, fs FileSystem) {
	HelpReadWrite(t, fs, "/foo.txt", "bar")
}

func TestReadWriteBasic4(t *testing.T, fs FileSystem) {
	HelpReadWrite(t, fs, "/foo1.txt", "bar1")
	HelpReadWrite(t, fs, "/foo2.txt", "bar2")
	HelpReadWrite(t, fs, "/foo3.txt", "bar3")
	HelpReadWrite(t, fs, "/foo4.txt", "bar4")
}

// TODO longer file paths and contents coming soon...

// ===== BEGIN SEEK DELETE TESTS =====

func TestSeekErrorBadFD(t *testing.T, fs FileSystem) {
   // must open an invalid FD
   _, err := fs.Seek(123456, 0, FromBeginning)
   assertEqualsFail(t, err, InvalidFD)
}

func TestSeekErrorBadOffsetOperation(t *testing.T, fs FileSystem) {
	fd := HelpOpen(t, fs, "/bad-offset-operation.txt", ReadWrite, Create)
   // Enforce only one option
   _, err := fs.Seek(fd, 0, FromBeginning | FromCurrent | FromEnd)
   assertEqualsFail(t, err, InvalidFD)
   _, err = fs.Seek(fd, 0, FromBeginning | FromCurrent)
   assertEqualsFail(t, err, InvalidFD)
   _, err = fs.Seek(fd, 0, FromEnd | FromCurrent)
   assertEqualsFail(t, err, InvalidFD)

   HelpClose(t, fs, fd)

   // TODO check size
   HelpDelete(t, fs, "/bad-offset-operation-1byte.txt")
}

func TestSeekErrorBadOffset1(t *testing.T, fs FileSystem) {
	fd := HelpOpen(t, fs, "/bad-offset-1byte.txt", ReadWrite, Create)
   _, err := fs.Seek(fd, -1, FromBeginning) // can't be negative
   assertEqualsFail(t, err, InvalidFD)

   _ = HelpSeek(t, fs, fd, 0, FromEnd)      // valid - at byte 0 
   _ = HelpSeek(t, fs, fd, 0, FromCurrent)  // valid - at byte 0

   _ = HelpSeek(t, fs, fd, 1, FromBeginning)    // valid - off end of file at byte 1
   _ = HelpSeek(t, fs, fd, 2, FromBeginning)    // valid - off end of file at byte 2
   _ = HelpSeek(t, fs, fd, 3, FromBeginning)    // valid - off end of file at byte 3

   _ = HelpSeek(t, fs, fd, 3, FromBeginning)    // valid - off end of file at byte 3
   _ = HelpSeek(t, fs, fd, 2, FromBeginning)    // valid - off end of file at byte 2
   _ = HelpSeek(t, fs, fd, 1, FromBeginning)    // valid - off end of file at byte 1
   _ = HelpSeek(t, fs, fd, 0, FromBeginning)    // valid - at byte 0


   n := HelpWrite(t, fs, fd, "c")
   assertExplain(t, n == 1, "the wr didn't wr 1 byte")

   // TODO check size of the file

   HelpDelete(t, fs, "/bad-offset-1byte.txt")
}

// Coming soon...
//TODO variants for other byte sweeps {8, 16, 64, 4096, ... GB? } godspeed to us

// ===== BEGIN SWEEP AND WRITE TESTS ===== 

// TODO these will start writing larger amounts of data
// TODO next set of tests will create holes, seek around, fill files with data

// ===== BEGIN MKDIR TESTS =====

func TestMkdir(t *testing.T, fs FileSystem) {
   HelpMkdir(t, fs, "/a-dir-1")
   HelpMkdir(t, fs, "/a-dir-2")
   HelpMkdir(t, fs, "/a-dir-3")
   HelpMkdir(t, fs, "/a-dir-4")
   HelpDelete(t, fs, "/a-dir-1")
   HelpDelete(t, fs, "/a-dir-2")
   HelpDelete(t, fs, "/a-dir-3")
   HelpDelete(t, fs, "/a-dir-4")
}


func TestMkdirTree(t *testing.T, fs FileSystem) {
   HelpMkdir(t, fs, "/a-dir-1")
   HelpMkdir(t, fs, "/a-dir-2")
   HelpMkdir(t, fs, "/a-dir-3")
   HelpMkdir(t, fs, "/a-dir-4")
   HelpMkdir(t, fs, "/a-dir-1/sub1")
   HelpMkdir(t, fs, "/a-dir-2/sub2/sub3")
   HelpMkdir(t, fs, "/a-dir-3/sub2/sub3/sub4/sub5")
   HelpMkdir(t, fs, "/a-dir-4/sub2/sub3/sub4/sub5/sub6")
   HelpDelete(t, fs, "/a-dir-1")
   HelpDelete(t, fs, "/a-dir-2")
   HelpDelete(t, fs, "/a-dir-3")
   HelpDelete(t, fs, "/a-dir-4")
}

// TODO subdirectories next...

