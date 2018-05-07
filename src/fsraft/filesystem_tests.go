package fsraft

import "testing"
import "fmt"
import "math/rand"

// Functionality tests for a FileSystem go here.
// Functions in this file are NOT "unit tests" because they ill not be run by "go test" because
// this file ends in "_testS", plural and they have more than one argument.
// Instead, these functionality tests can run against any class that implements the FileSystem interface
// by creating a unit test suite for your implementation class that calls these tests.

// ===== BEGIN OPEN CLOSE DELETE HELPERS =====

func HelpDelete(t *testing.T, fs FileSystem,
	pathname string) {
	success, err := fs.Delete(pathname)
	assertExplain(t, success && err == nil, "err %s deleting %s", err, pathname)
}

func HelpOpen(t *testing.T, fs FileSystem,
	path string, mode OpenMode, flags OpenFlags) int {
	fd, err := fs.Open(path, mode, flags)
	assertExplain(t, fd > 0 && err == nil, "err %s opening %s", err, path)
	return fd
}

func HelpTestOpenNotFound(t *testing.T, fs FileSystem,
	mode OpenMode, flags OpenFlags) {
	fd, err := fs.Open("/dirs/dont/exist/file", mode, flags)
	assertExplain(t, err == NotFound, "1st case didnt produce error")
	assertExplain(t, fd == -1, "fd should be negative on error")

	fd, err = fs.Open("f_wo_root_slash", mode, flags) // should we handle this?
	assertExplain(t, err == NotFound, "2nd case didnt produce error")
	assertExplain(t, fd == -1, "fd should be negative on error")
}

func HelpClose(t *testing.T, fs FileSystem,
	fd int) {
	success, err := fs.Close(fd)
	assertExplain(t, success && err == nil, "err closing fd %d", fd)
}

func HelpOpenClose(t *testing.T, fs FileSystem,
	path string, mode OpenMode, flags OpenFlags) {
	HelpClose(t, fs, HelpOpen(t, fs, path, mode, flags))
}

func HelpBatchOpen(t *testing.T, fs FileSystem,
	nFiles int, fmtStr string, mode OpenMode, flags OpenFlags) []int {
	fds := make([]int, nFiles)
	for ix := 0; ix < nFiles; ix++ {
		fds[ix] = HelpOpen(t, fs, fmt.Sprintf(fmtStr, ix), mode, flags)
	}
	return fds
}

func HelpBatchClose(t *testing.T, fs FileSystem, fds []int) {
	for ix := 0; ix < len(fds); ix++ {
		HelpClose(t, fs, fds[ix])
	}
}

func HelpBatchOpenClose(t *testing.T, fs FileSystem,
	nFiles int, fmtStr string, mode OpenMode, flags OpenFlags) {
	fds := HelpBatchOpen(t, fs, nFiles, fmtStr, mode, flags)
	HelpBatchClose(t, fs, fds)
}

func HelpBatchDelete(t *testing.T, fs FileSystem,
	nFiles int, fmtStr string) {
	for ix := 0; ix < nFiles; ix++ {
		HelpDelete(t, fs, fmt.Sprintf(fmtStr, ix))
	}
}

// ====== END OPEN CLOSE DELETE HELPERS =====

// ===== BEGIN MKDIR HELPERS =====

func HelpMkdir(t *testing.T, fs FileSystem,
	path string) {
	success, err := fs.Mkdir(path)
	assertNoError(t, err)
	assert(t, success)
	assertExplain(t, success, "mkdir fail on %s", path)
}

// ===== END MKDIR HELPERS =====

// ===== BEGIN READ WRITE SEEK HELPERS =====

func HelpMakeBytes(t *testing.T, n int) []byte {
   rndBytes := make([]byte, n)
   num, err := rand.Read(rndBytes)
   assertExplain(t, num == n, "mkbyte %d instead of %d", num, n)
   assertExplain(t, err == nil, "mkbyte err %s", err)
   return rndBytes
}

// error checked helper
func HelpSeek(t *testing.T, fs FileSystem,
	fd int, offset int, mode SeekMode) int {
	newPosition, err := fs.Seek(fd, offset, mode)
	assertNoError(t, err)
	if mode == FromBeginning {
		assertEquals(t, offset, newPosition)
	} // can we auto-check more seek behavior...
	return newPosition
}

// error checked helper
func HelpRead(t *testing.T, fs FileSystem,
	fd int, contents string, numBytes int) (int, []byte) {
	numRead, data, err := fs.Read(fd, numBytes)
	assertNoError(t, err)
	assertEquals(t, numBytes, numRead)
	assertEquals(t, contents, data)
	return numRead, data
}

// error checked helper
func HelpWrite(t *testing.T, fs FileSystem,
	fd int, contents string) int {
	bytes := []byte(contents)
	numBytes := len(bytes)
	numWritten, err := fs.Write(fd, numBytes, bytes)
	assertNoError(t, err)
	assertEquals(t, numBytes, numWritten)
	return numWritten
}

// error checked helper
func HelpReadWrite(t *testing.T, fs FileSystem,
	path string, contents string) int {
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

// @tests
// Whenever you add a new functionality test, be sure to add it to this list.
// This list is used in test_setup.go to run every functionality test on every difficulty.
var FunctionalityTests = []func(t *testing.T, fs FileSystem){
	TestBasicOpenClose,
	TestDeleteNotFound,
	TestCloseClosed,
	TestOpenOpened,
	TestOpenNotFound,
	TestOpenAlreadyExists,
	TestOpenROClose,
	TestOpenROClose,
	TestOpenROClose4,
	TestOpenROClose64,
	TestOpenRWClose,
	TestOpenRWClose4,
	TestOpenRWClose64,
	TestOpenCloseLeastFD,
	TestOpenCloseDeleteMaxFD,
	TestOpenCloseDeleteRoot,
	TestOpenCloseDeleteRootMax,
	TestReadWriteBasic,
	TestReadWriteBasic4,
	TestSeekErrorBadFD,
	TestSeekErrorBadOffsetOperation,
	TestSeekOffEOF,
	TestWrite1Byte,
   TestWrite8Bytes,
   TestWrite1KBytes,
   TestWrite1MBytes,
   TestWrite10MBytes,
   TestWrite100MBytes,
	// ========= the line in the sand =======
	TestMkdir,
	TestMkdirTree,
	TestOpenCloseDeleteAcrossDirectories,
}

// ===== BEGIN OPEN CLOSE TESTS ======

func TestBasicOpenClose(t *testing.T, fs FileSystem) {
	HelpOpenClose(t, fs, "/foo.txt", ReadWrite, Create)
}

func TestDeleteNotFound(t *testing.T, fs FileSystem) {
	success, err := fs.Delete("/this-file-doesnt-exist")
	assertExplain(t, !success, "delete on missing file was successful")
	assertExplain(t, err == NotFound, "err was not NotFound")
}

func TestCloseClosed(t *testing.T, fs FileSystem) {
	success, err := fs.Close(5) //arbirtary closed FD
	assertExplain(t, !success, "close on closed FD was successful")
	assertExplain(t, err == InactiveFD, "error needs to show issue with FD")
}

func TestOpenOpened(t *testing.T, fs FileSystem) {
	fd := HelpOpen(t, fs, "/file-open-successfully", ReadWrite, Create)
	fd2, err := fs.Open("/file-open-successfully", ReadWrite, Create)
	assertExplain(t, err == AlreadyOpen, "opened file returned err %s", err)
	assertExplain(t, fd2 == -1, "-1 needed on open error")
	HelpClose(t, fs, fd)
	HelpDelete(t, fs, "/file-open-successfully")
}

func TestOpenNotFound(t *testing.T, fs FileSystem) {
	HelpTestOpenNotFound(t, fs, ReadWrite, Append)
	HelpTestOpenNotFound(t, fs, ReadWrite, Create)
	HelpTestOpenNotFound(t, fs, ReadWrite, Truncate)
	HelpTestOpenNotFound(t, fs, ReadOnly, Append)
	HelpTestOpenNotFound(t, fs, ReadOnly, Create)
	HelpTestOpenNotFound(t, fs, ReadOnly, Truncate)
}

func TestOpenAlreadyExists(t *testing.T, fs FileSystem) {
	_ = HelpOpen(t, fs, "/first_file", ReadWrite, Create)
	fd, err := fs.Open("/first_file", ReadWrite, Create)
	assertExplain(t, err == AlreadyOpen, "wanted AlreadyOpen but err %s", err)
	assertExplain(t, fd == -1, "-1 needed on open error")
}

func TestOpenROClose(t *testing.T, fs FileSystem) {
	HelpOpenClose(t, fs, "/fooRO.txt", ReadOnly, Create)
}

func TestOpenROClose4(t *testing.T, fs FileSystem) {
	HelpBatchOpenClose(t, fs, 4, "/a-str-with-a-%d", ReadOnly, Create)
}

func TestOpenROClose64(t *testing.T, fs FileSystem) {
	HelpBatchOpenClose(t, fs, 64, "/str-2-with-a-%d", ReadOnly, Create)
}

func TestOpenRWClose(t *testing.T, fs FileSystem) {
	HelpOpenClose(t, fs, "/fooRO.txt", ReadOnly, Create)
}

func TestOpenRWClose4(t *testing.T, fs FileSystem) {
	HelpBatchOpenClose(t, fs, 4, "/str-3-with-a-%d", ReadWrite, Create)
}

func TestOpenRWClose64(t *testing.T, fs FileSystem) {
	HelpBatchOpenClose(t, fs, 64, "/str-4-with-a-%d", ReadWrite, Create)
}

func TestOpenCloseLeastFD(t *testing.T, fs FileSystem) {
	fd3A := HelpOpen(t, fs, "/A.txt", ReadWrite, Create)
	// Should be 3 because that's the lowest non-reserved non-active FD.
	assertEquals(t, 3, fd3A)
	HelpClose(t, fs, fd3A)

	fd3B := HelpOpen(t, fs, "/B.txt", ReadWrite, Create)
	// Should be 3 again because A.txt was closed, so FD=3 is now non-active again.
	assertEquals(t, 3, fd3B)
	// we're not closing it just yet

	fd4 := HelpOpen(t, fs, "/C.txt", ReadWrite, Create)
	// Should be 4 because 0-2 are reserved, 3 is taken, and 4 is next.
	assertEquals(t, 4, fd4)

	HelpClose(t, fs, fd3B)

	fd3C := HelpOpen(t, fs, "/D.txt", ReadWrite, Create)
	// B.txt was closed, so FD=3 is now non-active again.
	assertEquals(t, 3, fd3C)

	HelpClose(t, fs, fd3C)
	HelpClose(t, fs, fd4)
}

// open and close files checking all FDs open correctly up to limit,
// open a few past the limit, confirm we get errors, then close and delete all.
func TestOpenCloseDeleteMaxFD(t *testing.T, fs FileSystem) {
	maxFDCount := MaxActiveFDs
	maxFD := maxFDCount + 2 //max is offby1, & stdin, out, err...
	fmtStr := "/max-fd-%d.txt"
	prevFD := 0
	fds := make([]int, maxFDCount)
	for ix := 0; ix < maxFDCount; ix++ {
		fds[ix] = HelpOpen(t, fs, fmt.Sprintf(fmtStr, ix),
			ReadWrite, Create)
		assertExplain(t, fds[ix] > prevFD, "%d -> ? %d", prevFD, fds[ix])
		prevFD = fds[ix]
	}
	assertExplain(t, prevFD == maxFD,
		"wanted max FD %d but ended with %d", maxFD, prevFD)

	fd, err := fs.Open("/max-fd-one-more1.txt", ReadWrite, Create)
	assertExplain(t, err == TooManyFDsOpen, "RW 1 past max opened err: %s", err)
	assertExplain(t, fd == -1, "-1 needed on open error")
	fd, err = fs.Open("/max-fd-one-more2.txt", ReadWrite, Create)
	assertExplain(t, err == TooManyFDsOpen, "RW 2 past max opened err: %s", err)
	assertExplain(t, fd == -1, "-1 needed on open error")
	fd, err = fs.Open("/max-fd-one-more3.txt", ReadWrite, Create)
	assertExplain(t, err == TooManyFDsOpen, "RW 3 past max opened err: %s", err)
	assertExplain(t, fd == -1, "-1 needed on open error")

	fd, err = fs.Open("/max-fd-one-more1-ro.txt", ReadOnly, Create)
	assertExplain(t, err == TooManyFDsOpen, "R0 4 past max opened err: %s", err)
	assertExplain(t, fd == -1, "-1 needed on open error")
	fd, err = fs.Open("/max-fd-one-more2-ro.txt", ReadOnly, Create)
	assertExplain(t, err == TooManyFDsOpen, "R0 5 past max opened err: %s", err)
	assertExplain(t, fd == -1, "-1 needed on open error")
	fd, err = fs.Open("/max-fd-one-more3-ro.txt", ReadOnly, Create)
	assertExplain(t, err == TooManyFDsOpen, "R0 6 past max opened err: %s", err)
	assertExplain(t, fd == -1, "-1 needed on open error")

	HelpBatchClose(t, fs, fds)
	HelpBatchDelete(t, fs, maxFDCount, fmtStr)
}

func TestOpenCloseDeleteRoot(t *testing.T, fs FileSystem) {
	fd1 := HelpOpen(t, fs, "/foo", ReadWrite, Create)
	fd2 := HelpOpen(t, fs, "/bar", ReadWrite, Create)
	fd3 := HelpOpen(t, fs, "/baz", ReadWrite, Create)
	HelpClose(t, fs, fd1)
	HelpClose(t, fs, fd2)
	HelpClose(t, fs, fd3)
	HelpDelete(t, fs, "/foo")
	HelpDelete(t, fs, "/bar")
	HelpDelete(t, fs, "/baz")
}

func TestOpenCloseDeleteRootMax(t *testing.T, fs FileSystem) {
	maxFD := 64 //XXX update once we set it!!
	fds := HelpBatchOpen(t, fs, 64, "/max-root-opens-%d", ReadWrite, Create)
	HelpBatchClose(t, fs, fds)
	HelpBatchDelete(t, fs, maxFD, "/max-root-opens-%d")
}

// TODO next is same set of tests involving subdirs

//  ================== the line in the sand ====================
//  keeps moving down as tests begin passing and stay passing!

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
// TODO need a debug interface to simulate the test datastore runs out of space...
// TODO need a debug interface to simulate the test datastore has an IO error...

// ===== BEGIN SEEK DELETE TESTS =====

func TestSeekErrorBadFD(t *testing.T, fs FileSystem) {
	// must open an invalid FD
	_, err := fs.Seek(123456, 0, FromBeginning)
	assertEquals(t, err, InactiveFD)
}

func TestSeekErrorBadOffsetOperation(t *testing.T, fs FileSystem) {
	filename := "/bad-offset-operation.txt"
	fd := HelpOpen(t, fs, filename, ReadWrite, Create)
	// Enforce only one option
	_, err := fs.Seek(fd, 0, -1)
	assertExplain(t, err == IllegalArgument, "illegal seek mode wrong err")
	_, err = fs.Seek(fd, 0, 3)
	assertExplain(t, err == IllegalArgument, "illegal seek mode wrong err")
	_, err = fs.Seek(fd, 0, 0)
	assertExplain(t, err == nil, "illegal seek mode err")
	_, err = fs.Seek(fd, 0, 1)
	assertExplain(t, err == nil, "illegal seek mode err")
	_, err = fs.Seek(fd, 0, 2)
	assertExplain(t, err == nil, "illegal seek mode err")
	HelpClose(t, fs, fd)
	HelpDelete(t, fs, filename)
}

func TestSeekOffEOF(t *testing.T, fs FileSystem) {
	fd := HelpOpen(t, fs, "/seek-eof.txt", ReadWrite, Create)
	_, err := fs.Seek(fd, -1, FromBeginning) // can't be negative
	assertExplain(t, err == IllegalArgument, "illegal offset err %s", err)

	HelpSeek(t, fs, fd, 0, FromEnd)     // valid - at byte 0
	HelpSeek(t, fs, fd, 0, FromCurrent) // valid - at byte 0

	HelpSeek(t, fs, fd, 1, FromBeginning) // valid - off end of file at byte 1
	HelpSeek(t, fs, fd, 2, FromBeginning) // valid - off end of file at byte 2
	HelpSeek(t, fs, fd, 3, FromBeginning) // valid - off end of file at byte 3

	HelpSeek(t, fs, fd, 3, FromBeginning) // valid - off end of file at byte 3
	HelpSeek(t, fs, fd, 2, FromBeginning) // valid - off end of file at byte 2
	HelpSeek(t, fs, fd, 1, FromBeginning) // valid - off end of file at byte 1
	HelpSeek(t, fs, fd, 0, FromBeginning) // valid - at byte 0

	HelpDelete(t, fs, "/seek-eof.txt")
}

func TestWriteNBytesIter(t *testing.T, fs FileSystem, fileName string, nBytes int, iters int) {
	fd := HelpOpen(t, fs, fileName, ReadWrite, Create)
   data := make([]byte, 0)
   for i := 0; i < iters; i++ {
      data = HelpMakeBytes(t, nBytes)
      assertExplain(t, len(data) == nBytes, "made %d len array", len(data))
      n, err := fs.Write(fd, nBytes, data)
      assertExplain(t, err == nil, "err %s", err)
      assertExplain(t, n == nBytes, "wr %d", n)
   }
   HelpClose(t, fs, fd)
   HelpDelete(t, fs, fileName)
}

/*
func TestWriteReadNBytesIter(t *testing.T, fs FileSystem, fileName string, nBytes int, iters int) {
	fd := HelpOpen(t, fs, fileName, ReadWrite, Create)
   data := make([]byte, 0)
   for i := 0; i < iters; i++ {
      data = HelpMakeBytes(t, nBytes)
      assertExplain(t, len(data) == nBytes, "made %d len array", len(data))
      n, err := fs.Write(fd, nBytes, data)
      assertExplain(t, err == nil, "err %s", err)
      assertExplain(t, n == nBytes, "wr %d", n)
   }
   HelpClose(t, fs, fd)
   HelpDelete(t, fs, fileName)
}*/

func TestWrite1Byte(t *testing.T, fs FileSystem) {
   TestWriteNBytesIter(t, fs, "/wr-1.txt", 1, 5)
}

func TestWrite8Bytes(t *testing.T, fs FileSystem) {
   TestWriteNBytesIter(t, fs, "/wr-8.txt", 8, 5)
}

func TestWrite1KBytes(t *testing.T, fs FileSystem) {
   TestWriteNBytesIter(t, fs, "/wr-1k.txt", 1000, 5)
}

func TestWrite1MBytes(t *testing.T, fs FileSystem) {
   TestWriteNBytesIter(t, fs, "/wr-1m.txt", 1000000, 5)
}

func TestWrite10MBytes(t *testing.T, fs FileSystem) {
   TestWriteNBytesIter(t, fs, "/wr-10m.txt", 10000000, 5)
}

func TestWrite100MBytes(t *testing.T, fs FileSystem) {
   TestWriteNBytesIter(t, fs, "/wr-100m.txt", 100000000, 3)
}

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
