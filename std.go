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

const (
	// LevelKey is field key for logging Level.
	LevelKey = "level"
	// LoggerKey is field key for logger name.
	LoggerKey = "logger"
)

func init() {
	UseLevelStore(NewStdLevelStore())
	UseProvider(NewStdLoggerProvider(NewStdOutput(log.Default())))
}

// Field represent a key/value pair.
type Field struct {
	// Key is the key, string.
	Key string
	// Value is the value, may be a Valuer.
	Value interface{}
}

// Output is the real final output of builtin standard Logger and Printer.
// It calls the underlying logging / printing infrastructures.
type Output interface {
	// Output output msg, fields at a specific level to the underlying
	// logging / printing infrastructures.
	Output(ctx context.Context, level Level, msg string, fields []Field)
}

// stdOutput is an Output implementation based on Go SDK log.Logger.
type stdOutput struct {
	out     *log.Logger
	bufPool *sync.Pool
}

// NewStdOutput create a Output based on Go SDK log.Logger.
// It output to the provided log.Logger.
func NewStdOutput(out *log.Logger) Output {
	return &stdOutput{
		out: out,
		bufPool: &sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
}

func (s *stdOutput) Output(ctx context.Context, _ Level, msg string, fields []Field) {
	buf := s.bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		s.bufPool.Put(buf)
	}()
	for _, field := range fields {
		_, _ = fmt.Fprintf(buf, "%s=%v ", field.Key, Value(ctx, field.Value))
	}
	_, _ = fmt.Fprint(buf, msg)
	_ = s.out.Output(2, buf.String())
}

// stdPrinter is the builtin implementation of Printer.
type stdPrinter struct {
	output  Output
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
	p.output.Output(p.ctx, p.level, buf.String(), p.fields)
}

func (p *stdPrinter) Printf(format string, v ...interface{}) {
	buf := p.bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		p.bufPool.Put(buf)
	}()
	_, _ = fmt.Fprintf(buf, format, v...)
	p.output.Output(p.ctx, p.level, buf.String(), p.fields)
}

func (p *stdPrinter) Println(v ...interface{}) {
	buf := p.bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		p.bufPool.Put(buf)
	}()
	buf.WriteString(fmt.Sprintln(v...))
	buf.Truncate(buf.Len() - 1)
	p.output.Output(p.ctx, p.level, buf.String(), p.fields)
}

func (p *stdPrinter) With(key string, value interface{}) Printer {
	if key == "" {
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

// stdLogger is the builtin implementation of Logger
type stdLogger struct {
	output Output
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

func (l *stdLogger) AtLevel(level Level, ctx context.Context) Printer {
	if !l.levelEnabled(level) {
		return NewNopPrinter()
	}
	return &stdPrinter{
		output: l.output,
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
// Output to print logging messages.
func NewStdLogger(name string, output Output) Logger {
	return &stdLogger{
		output: output,
		name:   name,
	}
}

// NewStdLoggerProvider make a LoggerProvider which produce Logger via
// NewStdLogger function.
func NewStdLoggerProvider(output Output) LoggerProvider {
	return func(name string) Logger {
		return NewStdLogger(name, output)
	}
}

// stdLevelStore is builtin implementation of LevelStore.
// It store and update levels with a Copy-On-Write map.
type stdLevelStore struct {
	store *map[string]Level
}

func (l *stdLevelStore) Get(name string) Level {
	store := *l.store
	if store == nil {
		return InfoLevel
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
	if lvl, ok := store[""]; ok {
		return lvl
	}
	return InfoLevel
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
