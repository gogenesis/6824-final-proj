package ad

import (
	"fmt"
	"path"
	"runtime"
	"sync"
	"time"
)

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
	debugPrivate(level, "", formatStr, a...)
}

// Indicate that an object should print its state when debugging.
type Debuggable interface {
	DebugPrefix() string
}

// Debug with some state about an object.
func DebugObj(db Debuggable, level int, formatStr string, a ...interface{}) {
	debugPrivate(level, db.DebugPrefix(), formatStr, a...)
}

func debugPrivate(level int, stateStr string, formatStr string, a ...interface{}) {
	debugMutex.Lock()
	defer debugMutex.Unlock()
	t := time.Now().Format("3:04:05.000")

	if CURRENT_DEBUG_LEVEL <= level {
		loggingLevelNames := [...]string{"TRACE", " RPC ", "WARN "}
		levelName := loggingLevelNames[level]
		// 2 to use the function 2 above this in the stack trace
		_, fileWithPath, lineNum, _ := runtime.Caller(2)
		_, file := path.Split(fileWithPath)
		// 12 because most file names are 12 lines or fewer
		fileWithPadding := fmt.Sprintf("%12v", file)

		fmt.Printf("[%v] %v %v:%03d %v %v\n", levelName, t, fileWithPadding, lineNum, stateStr,
			fmt.Sprintf(formatStr, a...))
	}
}
