package fsraft

import (
	"fmt"
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
