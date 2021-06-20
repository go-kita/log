package log

import (
	"context"
	"log"
	"unsafe"

	ua "go.uber.org/atomic"
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

// nopPrinter is a Printer that print nothing.
type nopPrinter struct {
}

// NewNopPrinter returns a Printer which print nothing.
func NewNopPrinter() Printer {
	return &nopPrinter{}
}

func (p *nopPrinter) Print(_ ...interface{}) {
}

func (p *nopPrinter) Printf(_ string, _ ...interface{}) {
}

func (p *nopPrinter) Println(_ ...interface{}) {
}

func (p *nopPrinter) With(_ string, _ interface{}) Printer {
	return p
}

// Logger represents a logger that can provide Printers. A Logger has a name,
// and there may be a lower limit of logging Level the Logger supported.
type Logger interface {
	// AtLevel get a Printer wrapping the provided context.Context at specified
	// logging Level. If the Level is not enabled to the Logger, nothing will be
	// print when calling the Print functions of returned Printer.
	// If the context.Context provided is nil, context.Background() will be used.
	AtLevel(ctx context.Context, level Level) Printer
}

// LoggerProvider is provider function that provide a non-nil Logger by name.
type LoggerProvider func(name string) Logger

var _loggerProvider = ua.NewUnsafePointer(unsafe.Pointer((*LoggerProvider)(nil)))

// UseProvider register a LoggerProvider for use by default.
// If this function is called more than once, the last call wins.
func UseProvider(provider LoggerProvider) {
	for {
		old := _loggerProvider.Load()
		if _loggerProvider.CAS(old, unsafe.Pointer(&provider)) {
			break
		}
	}
}

// Get return a Logger by name.
func Get(name string) Logger {
	provider := *(*LoggerProvider)(_loggerProvider.Load())
	if provider == nil {
		return NewStdLogger(name, NewStdOutPutter(log.Default()))
	}
	return (*(*LoggerProvider)(_loggerProvider.Load()))(name)
}
