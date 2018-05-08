package raft

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

type debuggable interface {
	debugPrefix() string
}

// logging levels
const (
	TRACE               = iota // everything
	CURRENT                    // currently debugging
	RPC                        // no more than one per RPC or state change
	WARN                       // big warnings
	NONE                       // for submission
	CURRENT_DEBUG_LEVEL = NONE
)

// Make sure that logging a debug message is atomic
var debugMutex sync.Mutex

// Write some stuff to stdout iff this instance is alive
func debug(db debuggable, level int, formatStr string, a ...interface{}) {
	debugMutex.Lock()
	defer debugMutex.Unlock()
	t := time.Now().Format("05.000")

	if CURRENT_DEBUG_LEVEL <= level {
		loggingLevelNames := [...]string{"R TRAC", "R CURR", "R RPC ", "R WARN"}
		levelName := loggingLevelNames[level]
		// 1 to use the function 1 above this in the stack trace
		_, _, lineNum, _ := runtime.Caller(1)
		var debugStr string
		switch len(a) {
		case 0:
			debugStr = formatStr
		case 1:
			debugStr = fmt.Sprintf(formatStr, a[0])
		default:
			debugStr = fmt.Sprintf(formatStr, a...)
		}

		statusStr := db.debugPrefix()

		fmt.Printf("[%v] %v %03d [%v] %v\n", levelName, t, lineNum, statusStr, debugStr)
	}
}
