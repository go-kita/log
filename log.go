package log

import (
	"context"
	"errors"
	stdlog "log"
)

// Printer represents a stateful printer that is at specific level,
// carrying fields including logger name.
// It also usually carry a context.Context for the Value calculation.
// Users are expected NEVER new a Printer instance by hand.
// It's usually got from a Logger and left/throw away after doing print.
// NEVER share a Printer between methods and goroutines.
type Printer interface {
	// Print formats using the default formats for its operands and writes
	// result as message value.
	// Spaces are added between operands when neither is a string.
	// The message value and other key/value carried by the Printer will be
	// passed to the implementing logging layer together.
	Print(v ...interface{})
	// Printf formats according to a format specifier and writes result as
	// message value.
	// The message value and other key/value carried by the Printer will be
	// passed to the implementing logging layer together.
	Printf(format string, v ...interface{})
	// Println formats using the default formats for its operands and writes
	// result as message value.
	// Spaces are always added between operands.
	// The message value and other key/value carried by the Printer will be
	// passed to the implementing logging layer together.
	Println(v ...interface{})
	// With add key/value pair to the Printer. If the key is already exist,
	// the existing value will be override by the given value. So users should
	// not use these internally used keys: logger, level, and so on.
	// Depending on the implementation, the key `caller` (and some other keys,
	// please refer the manual of the implementation) may should be avoid, since
	// the caller key/value pair may be added by the implementation.
	With(key string, value interface{}) Printer
}

// Logger represents a logger that can provide Printers. A Logger has a name,
// and there may be a lower limit of logging Level the Logger supported.
type Logger interface {
	// LevelEnabled indicates whether the Logger can output line at the provided
	// level.
	LevelEnabled(level Level) bool
	// Printer get a Printer wrapping the provided context.Context.
	// If the Level of the Logger is ClosedLevel, nothing will be print when
	// calling the Print functions of returned Printer.
	Printer(ctx context.Context) Printer
	// Name returns the name of the Logger.
	Name() string
	// AtLevel get a Printer wrapping the provided context.Context at specified
	// logging Level. If the Level is not enabled to the Logger, nothing will be
	// print when calling the Print functions of returned Printer.
	AtLevel(level Level, ctx context.Context) Printer
}

// Manager represents a Logger manager. It provides Logger by name.
type Manager interface {
	// Get returns a Logger of name from the Manager.
	// If no logger of the provided name is configured, a parent or default
	// Logger should be returned. Refer to the manual of implementations.
	Get(name string) Logger
	// Level set logging Level for logger of name.
	// If you call this function with same name more than once, the last call
	// wins.
	Level(name string, level Level)
}

var registeredManager Manager

// UseManager register a Manager to use.
// If this function is called more than once, the last call wins.
// Note that, all got Loggers before UseManager call have no relation with
// the new registered Manager, you can not control their behaviors after you
// call this function.
func UseManager(mng Manager) {
	registeredManager = mng
}

// GetManager returns the registered Manager to use.
// If none Manager is registered, it will return a none nil error.
func GetManager() (Manager, error) {
	if registeredManager == nil {
		return nil, errors.New("no log manager registered")
	}
	return registeredManager, nil
}

// Get return a Logger from registered Manager.
// If none Manager is registered, it return a new built logger, and note that,
// you can not control the returned new built logger's behaviors, so you'd
// better always call UseManager registering a valid Manager before call this
// function.
// Note that, this method will never refill the registered Manager if absent.
func Get(name string) Logger {
	mng, err := GetManager()
	if err != nil {
		mng := NewManager(stdlog.Default(), RootLevel(DebugLevel))
		return mng.Get("")
	}
	return mng.Get(name)
}
