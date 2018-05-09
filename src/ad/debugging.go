package ad

import (
	"fmt"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

// logging levels
const (
	NONE  = iota // for submission.
	WARN         // only for warnings or things that go wrong
	RPC          // no more than two per RPC or state change
	TRACE        // everything!
)

// Indicate that an object should print its state when debugging.
type Debuggable interface {
	DebugPrefix() string
}

var loggingLevelNames = [...]string{
	"NONE",
	"WARN",
	"RPC",
	"TRACE",
}

func debugLevelName(level int) string {
	return loggingLevelNames[level]
}

// Make sure that logging a debug message is atomic
var debugMutex sync.Mutex

// Write some stuff to stdout
func Debug(level int, formatStr string, a ...interface{}) {
	debugPrivate(level, "", formatStr, a...)
}

// Debug with some state about an object.
func DebugObj(db Debuggable, level int, formatStr string, a ...interface{}) {
	debugPrivate(level, db.DebugPrefix(), formatStr, a...)
}

func debugPrivate(level int, stateStr string, formatStr string, a ...interface{}) {
	debugMutex.Lock()
	defer debugMutex.Unlock()
	t := time.Now().Format("3:04:05.000")

	// 2 to use the function 2 above this in the stack trace
	_, fileWithPath, lineNum, _ := runtime.Caller(2)
	dir, file := path.Split(fileWithPath)
	packageName := strings.ToLower(path.Base(dir))
	levelName := debugLevelName(level)
	//fmt.Printf("%v %v ", packageName, levelName)
	if level <= packageNamesToDebugLevels[packageName] {
		//fmt.Printf("<= %v, printing\n", debugLevelName(packageNamesToDebugLevels[packageName]))
		// these are tbh pretty arbitrary amounts of padding
		// - for left-align
		packageWithPadding := fmt.Sprintf("%-8v", packageName)
		fileWithPadding := fmt.Sprintf("%12v", file)
		levelNameWithPadding := fmt.Sprintf("%-5v", levelName)

		fmt.Printf("[%v %v] %v %v:%03d [%v] %v\n", packageWithPadding, levelNameWithPadding, t, fileWithPadding, lineNum, stateStr,
			fmt.Sprintf(formatStr, a...))
	} else {
		//fmt.Printf("> %v, not printing\n", debugLevelName(packageNamesToDebugLevels[packageName]))
	}
}
