package log

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"unsafe"

	ua "go.uber.org/atomic"
)

// Define some builtin and standard logging field key.
const (
	// LevelKey is field key for logging Level.
	LevelKey = "level"
	// LoggerKey is field key for logger name.
	LoggerKey = "logger"
)

// Field represent a key/value pair.
type Field struct {
	// Key is the key, string.
	Key string
	// Value is the value, may be a Valuer.
	Value interface{}
}

// ======== OutPutter =========

// OutPutter is the real final output of builtin standard Logger and Printer.
// It calls the underlying logging / printing infrastructures.
type OutPutter interface {
	// OutPut output msg, fields at a specific level to the underlying
	// logging / printing infrastructures.
	OutPut(ctx context.Context, name string, level Level, msg string, fields []Field, callDepth int)
}

// stdOutPutter is an OutPutter implementation based on Go SDK log.Logger.
type stdOutPutter struct {
	out     *log.Logger
	bufPool *sync.Pool
}

// NewStdOutPutter create a OutPutter based on Go SDK log.Logger.
// It output to the provided log.Logger.
// If nil is passed to the function, log.Default() will be called to get a
// log.Logger.
func NewStdOutPutter(out *log.Logger) OutPutter {
	if out == nil {
		out = log.Default()
	}
	return &stdOutPutter{
		out: out,
		bufPool: &sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
}

func (s *stdOutPutter) OutPut(ctx context.Context, _ string, _ Level, msg string, fields []Field, callDepth int) {
	buf := s.bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		s.bufPool.Put(buf)
	}()
	for _, field := range fields {
		if len(field.Key) == 0 {
			continue
		}
		_, _ = fmt.Fprintf(buf, "%s=%v ", field.Key, Value(ctx, field.Value))
	}
	_, _ = fmt.Fprint(buf, msg)
	_ = s.out.Output(callDepth+3, buf.String())
}

// ======== OutPutFilter =========

var _ OutPutter = &OutPutFilter{}

// OutPutFilter is a filter wrapped a OutPutter.
type OutPutFilter struct {
	underlying      OutPutter
	enableFunc      func(ctx context.Context, name string, level Level) bool
	fieldModifyFunc func(ctx context.Context, field *Field)
}

// OutPut do checking and modification before calling the OutPut() method of the wrapped OutPutter.
func (o *OutPutFilter) OutPut(
	ctx context.Context, name string, level Level, msg string, fields []Field, callDepth int) {
	if o.underlying == nil {
		return
	}
	if o.enableFunc != nil && !o.enableFunc(ctx, name, level) {
		return
	}
	for i := 0; i < len(fields); i++ {
		o.fieldModifyFunc(ctx, &fields[i])
	}
	o.underlying.OutPut(ctx, name, level, msg, fields, callDepth+1)
}

// FilterEnable build a OutPutFilter wrapping the provided OutPutter. The output
// will be skipped if the enable checking function return false.
func FilterEnable(o OutPutter, f func(ctx context.Context, name string, level Level) bool) OutPutter {
	if o == nil {
		return o
	}
	return &OutPutFilter{
		underlying: o,
		enableFunc: f,
	}
}

// FilterRemoveField build a OutPutFilter wrapping the provided OutPutter. Any field
// with the specific name will be skipped.
func FilterRemoveField(o OutPutter, name string) OutPutter {
	if o == nil {
		return o
	}
	return &OutPutFilter{
		underlying: o,
		fieldModifyFunc: func(ctx context.Context, field *Field) {
			if field.Key == name {
				field.Key = ""
			}
		},
	}
}

// FilterCoverField build a OutPutFilter wrapping the provided OutPutter.
// The value of the field with specific name will be replaced.
func FilterCoverField(o OutPutter, name string, replace interface{}) OutPutter {
	if o == nil {
		return o
	}
	return &OutPutFilter{
		underlying: o,
		fieldModifyFunc: func(ctx context.Context, field *Field) {
			if field.Key != name {
				return
			}
			field.Value = replace
		},
	}
}

// ======== Printer =========

var _ Printer = (*stdPrinter)(nil)

// stdPrinter is the builtin implementation of Printer.
type stdPrinter struct {
	logger  *stdLogger
	level   Level
	fields  []Field
	ctx     context.Context
	bufPool *sync.Pool
}

func (p *stdPrinter) Print(v ...interface{}) {
	buf := p.bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		p.bufPool.Put(buf)
	}()
	_, _ = fmt.Fprint(buf, v...)
	p.logger.output.OutPut(p.ctx, p.logger.name, p.level, buf.String(), p.fields, 0)
}

func (p *stdPrinter) Printf(format string, v ...interface{}) {
	buf := p.bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		p.bufPool.Put(buf)
	}()
	_, _ = fmt.Fprintf(buf, format, v...)
	p.logger.output.OutPut(p.ctx, p.logger.name, p.level, buf.String(), p.fields, 0)
}

func (p *stdPrinter) Println(v ...interface{}) {
	buf := p.bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		p.bufPool.Put(buf)
	}()
	buf.WriteString(fmt.Sprintln(v...))
	buf.Truncate(buf.Len() - 1)
	p.logger.output.OutPut(p.ctx, p.logger.name, p.level, buf.String(), p.fields, 0)
}

func (p *stdPrinter) With(key string, value interface{}) Printer {
	if len(key) == 0 {
		return p
	}
	for i := 0; i < len(p.fields); i++ {
		if p.fields[i].Key == key {
			p.fields[i].Value = value
			return p
		}
	}
	p.fields = append(p.fields, Field{key, value})
	return p
}

// ======== Logger =========

var _ Logger = (*stdLogger)(nil)

// stdLogger is the builtin implementation of Logger
type stdLogger struct {
	output OutPutter
	name   string
}

// NewStdLogger create a Logger by name. The Logger returned will use the provided
// OutPutter to print logging messages.
func NewStdLogger(name string, output OutPutter) Logger {
	if output == nil {
		output = NewStdOutPutter(log.Default())
	}
	return &stdLogger{
		output: output,
		name:   name,
	}
}

func (l *stdLogger) levelEnabled(level Level) bool {
	store := GetLevelStore()
	ll := InfoLevel
	if store != nil {
		ll = store.Get(l.name)
	}
	return ll != ClosedLevel && ll <= level
}

func (l *stdLogger) AtLevel(ctx context.Context, level Level) Printer {
	if !l.levelEnabled(level) {
		return NewNopPrinter()
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return &stdPrinter{
		logger: l,
		level:  level,
		fields: []Field{
			{LevelKey, level},
			{LoggerKey, l.name},
		},
		ctx: ctx,
		bufPool: &sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
}

// ======== LevelStore =========

// LevelStore stores and provides the lowest logging Level limit of a Logger by
// name.
type LevelStore interface {
	// Get provide the lowest logging Level that a Logger can support by name.
	Get(name string) Level
	// Set update the lowest logging Level that a Logger can support by name.
	// If this method is called more than once, the last call wins.
	Set(name string, level Level) LevelStore
	// UnSet clear the set lowest logging Level that a Logger can support by
	// name.
	UnSet(name string) LevelStore

	// Restore clear all known levels and reset levels according to
	// the provided level map.
	Restore(mp map[string]Level)
	// Levels return all known levels as map[string]Level.
	Levels() map[string]Level
}

var _levelStore = &stdLevelStore{
	store: ua.NewUnsafePointer(unsafe.Pointer(&map[string]Level{
		"": InfoLevel,
	})),
}

// GetLevelStore returns the registered LevelStore for use by default.
// Nil may be returned, if no LevelStore is registered or the registered
// LevelStore has be cleared.
func GetLevelStore() LevelStore {
	return _levelStore
}

// stdLevelStore is builtin implementation of LevelStore.
// It store and update levels with a Copy-On-Write map.
type stdLevelStore struct {
	store *ua.UnsafePointer
}

var _ LevelStore = (*stdLevelStore)(nil)

func (l *stdLevelStore) Get(name string) Level {
	store := *(*map[string]Level)(l.store.Load())
	for name != "" {
		if lvl, ok := store[name]; ok {
			return lvl
		}
		lastIndex := strings.LastIndexFunc(name, func(r rune) bool {
			return r == '.' || r == '/'
		})
		if lastIndex == -1 {
			break
		}
		name = name[:lastIndex]
	}
	return store[""]
}

func (l *stdLevelStore) Set(name string, level Level) LevelStore {
	for {
		store := map[string]Level{}
		old := (*map[string]Level)(l.store.Load())
		for oldName, oldLevel := range *old {
			store[oldName] = oldLevel
		}
		store[name] = level
		if l.store.CAS(unsafe.Pointer(old), unsafe.Pointer(&store)) {
			break
		}
	}
	return l
}

func (l *stdLevelStore) UnSet(name string) LevelStore {
	for {
		store := map[string]Level{}
		old := (*map[string]Level)(l.store.Load())
		for oldName, oldLevel := range *old {
			if oldName == name {
				continue
			}
			store[oldName] = oldLevel
		}
		if l.store.CAS(unsafe.Pointer(old), unsafe.Pointer(&store)) {
			break
		}
	}
	return l
}

func (l *stdLevelStore) Restore(mp map[string]Level) {
	store := make(map[string]Level, len(mp))
	for name, level := range mp {
		store[name] = level
	}
	l.store.Swap(unsafe.Pointer(&store))
}

func (l *stdLevelStore) Levels() map[string]Level {
	store := *(*map[string]Level)(l.store.Load())
	mp := make(map[string]Level, len(store))
	for name, level := range store {
		mp[name] = level
	}
	return mp
}

// ======== LoggerProvider =========

// NewStdLoggerProvider make a LoggerProvider which produce Logger via
// NewStdLogger function.
func NewStdLoggerProvider(outPutter OutPutter) LoggerProvider {
	return func(name string) Logger {
		return NewStdLogger(name, outPutter)
	}
}

func init() {
	UseProvider(NewStdLoggerProvider(NewStdOutPutter(log.Default())))
}
