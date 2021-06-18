package log

import (
	"fmt"
	"math"
	"sync/atomic"
	"unsafe"
)

// A Level is a logging priority. Higher levels are more important.
type Level int8

// Define logging Levels. Default level is InfoLevel.
const (
	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel Level = iota - 1
	// InfoLevel is the default logging priority.
	InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel
	// ClosedLevel logs output nothing.
	ClosedLevel = math.MaxInt8
	// allLevel logs output anything.
	allLevel = math.MinInt8
)

var levelNames = &map[Level]string{
	DebugLevel: "DEBUG",
	InfoLevel:  "INFO",
	WarnLevel:  "WARN",
	ErrorLevel: "ERROR",
}

// RegisterLevelName register the name of one level. If the level is already exists,
// the function will overwrite it. If the name given is empty, it register nothing,
// or it deregister a level name.
func RegisterLevelName(level Level, name string) {
	for {
		mp := make(map[Level]string, len(*levelNames))
		for l, n := range *levelNames {
			mp[l] = n
		}
		mp[level] = name
		if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&levelNames)),
			unsafe.Pointer(levelNames), unsafe.Pointer(&mp)) {
			break
		}
	}
}

// String returns a lower-case ASCII representation of the log level.
func (l Level) String() string {
	name := (*levelNames)[l]
	if len(name) == 0 {
		return fmt.Sprintf("Level(%d)", l)
	}
	return name
}
