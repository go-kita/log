package log

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"
)

// Define some builtin and standard logging field key.
const (
	// LevelKey is field key for logging Level.
	LevelKey = "level"
	// LoggerKey is field key for logger name.
	LoggerKey = "logger"
)

func init() {
	UseLevelStore(NewStdLevelStore())
	UseProvider(NewStdLoggerProvider(NewStdOutPutter(log.Default())))
}

// Field represent a key/value pair.
type Field struct {
	// Key is the key, string.
	Key string
	// Value is the value, may be a Valuer.
	Value interface{}
}

// OutPutter is the real final output of builtin standard Logger and Printer.
// It calls the underlying logging / printing infrastructures.
type OutPutter interface {
	// OutPut output msg, fields at a specific level to the underlying
	// logging / printing infrastructures.
	OutPut(ctx context.Context, name string, level Level, msg string, fields []Field, callDepth int)
}

var _ OutPutter = &stdOutPutter{}

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

var _ Logger = (*stdLogger)(nil)

// stdLogger is the builtin implementation of Logger
type stdLogger struct {
	output OutPutter
	name   string
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

var _ LevelStore = (*stdLevelStore)(nil)

// stdLevelStore is builtin implementation of LevelStore.
// It store and update levels with a Copy-On-Write map.
type stdLevelStore struct {
	store *map[string]Level
}

func (l *stdLevelStore) Get(name string) Level {
	store := *l.store
	if store == nil {
		return allLevel
	}
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

func (l *stdLevelStore) Set(name string, level Level) {
	for {
		store := map[string]Level{}
		oldStore := l.store
		for oldName, oldLevel := range *oldStore {
			store[oldName] = oldLevel
		}
		store[name] = level
		addr := (*unsafe.Pointer)(unsafe.Pointer(&l.store))
		if atomic.CompareAndSwapPointer(addr, unsafe.Pointer(oldStore), unsafe.Pointer(&store)) {
			break
		}
	}
}

func (l *stdLevelStore) UnSet(name string) {
	for {
		store := map[string]Level{}
		oldStore := l.store
		for oldName, oldLevel := range *oldStore {
			if oldName == name {
				continue
			}
			store[oldName] = oldLevel
		}
		addr := (*unsafe.Pointer)(unsafe.Pointer(&l.store))
		if atomic.CompareAndSwapPointer(addr, unsafe.Pointer(oldStore), unsafe.Pointer(&store)) {
			break
		}
	}
}

// StdLevelStoreOption is option for NewStdLevelStore constructor function.
type StdLevelStoreOption func(ls *stdLevelStore)

// LoggerLevel is a option which define Level for name.
func LoggerLevel(name string, level Level) StdLevelStoreOption {
	return func(ls *stdLevelStore) {
		ls.Set(name, level)
	}
}

// RootLevel is a option which define the root Level.
func RootLevel(level Level) StdLevelStoreOption {
	return func(ls *stdLevelStore) {
		ls.Set("", level)
	}
}

// NewStdLevelStore create a LevelStore.
func NewStdLevelStore(opts ...StdLevelStoreOption) LevelStore {
	store := &stdLevelStore{
		store: &map[string]Level{
			"": InfoLevel,
		},
	}
	for _, opt := range opts {
		opt(store)
	}
	return store
}

// NewStdLoggerProvider make a LoggerProvider which produce Logger via
// NewStdLogger function.
func NewStdLoggerProvider(outPutter OutPutter) LoggerProvider {
	return func(name string) Logger {
		return NewStdLogger(name, outPutter)
	}
}
