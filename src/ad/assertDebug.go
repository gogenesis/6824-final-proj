package ad

import (
	"fmt"
	"path"
	"runtime"
	"sync"
	"time"
)

func Assert(cond bool) {
	// Skip assertions when debugging is turned off for performance
	if (CURRENT_DEBUG_LEVEL != NONE) && (!cond) {
		panic("Assertion error!")
	}
}

// logging levels
const (
	TRACE               = iota // everything!
	RPC                        // no more than one per RPC or state change
	WARN                       // big warnings
	NONE                       // for submission.
	CURRENT_DEBUG_LEVEL = RPC
)

// Make sure that logging a debug message is atomic
var debugMutex sync.Mutex

// Write some stuff to stdout
func Debug(level int, formatStr string, a ...interface{}) {
	debugMutex.Lock()
	defer debugMutex.Unlock()
	t := time.Now().Format("3:04:05.000")

	if CURRENT_DEBUG_LEVEL <= level {
		loggingLevelNames := [...]string{"TRACE", "RPC", "WARN"}
		levelName := loggingLevelNames[level]
		// 1 to use the function 1 above this in the stack trace
		_, fileWithPath, lineNum, _ := runtime.Caller(1)
		_, file := path.Split(fileWithPath)
		// 12 because most file names are 12 lines or fewer
		fileWithPadding := fmt.Sprintf("%12v", file)

		fmt.Printf("[%v] %v %v:%03d %v\n", levelName, t, fileWithPadding, lineNum, fmt.Sprintf(formatStr, a...))
	}
}
