package fsraft

import (
	"fmt"
	"testing"
)

// You can change these from panic to t.Fatalf if it would make your life easier
func assertNoError(e error) {
	if e != nil {
		panic(e.Error())
	}
}

func assertEquals(expected, actual interface{}) {
	if expected != actual {
		panic(fmt.Sprintf("Assertion error! Expected %+v, got %+v\n", expected, actual))
	}
}

func assert(cond bool) {
	if !cond {
		panic("Assertion error!")
	}
}

func TestBasicReadWrite(t *testing.T) {
	fileName := "/foo.txt" // arbitrarily
	contents := "bar"      // also arbitrarily
	bytes := []byte(contents)
	numBytes := len(bytes)

	cfg := make_config(t, 1, false, -1)
	ck := cfg.makeClient(cfg.All())

	fd, err := ck.Open(fileName, ReadWrite, Create)
	assertNoError(err)

	numWritten, err := ck.Write(fd, numBytes, bytes)
	assertNoError(err)
	assertEquals(numBytes, numWritten)

	newPosition, err := ck.Seek(fd, 0, FromBeginning)
	assertNoError(err)
	assertEquals(0, newPosition)

	numRead, data, err := ck.Read(fd, numBytes)
	assertNoError(err)
	assertEquals(numBytes, numRead)
	assertEquals(bytes, data)

	success, err := ck.Close(fd)
	assertNoError(err)
	assert(success)
}
