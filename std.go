package log

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/cornelk/hashmap"
)

const (
	LevelKey  = "level"
	LoggerKey = "logger"
)

func init() {
	UseManager(NewManager(log.Default()))
}

// field represent a key/value pair.
type field struct {
	// Key is the key, string.
	Key string
	// Value is the value, may be a Valuer.
	Value interface{}
}

// printer is the builtin implementation of Printer.
type printer struct {
	logger  *logger
	level   Level
	fields  []field
	ctx     context.Context
	bufPool *sync.Pool
}

func (p *printer) Print(v ...interface{}) {
	if !p.logger.LevelEnabled(p.level) {
		return
	}
	buf := p.bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		p.bufPool.Put(buf)
	}()
	for _, field := range p.fields {
		_, _ = fmt.Fprintf(buf, "%s=%v ", field.Key, Value(p.ctx, field.Value))
	}
	_, _ = fmt.Fprint(buf, v...)
	_ = p.logger.out.Output(2, buf.String())
}

func (p *printer) Printf(format string, v ...interface{}) {
	if !p.logger.LevelEnabled(p.level) {
		return
	}
	buf := p.bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		p.bufPool.Put(buf)
	}()
	for _, field := range p.fields {
		_, _ = fmt.Fprintf(buf, "%s=%v ", field.Key, Value(p.ctx, field.Value))
	}
	_, _ = fmt.Fprintf(buf, format, v...)
	_ = p.logger.out.Output(2, buf.String())
}

func (p *printer) Println(v ...interface{}) {
	if !p.logger.LevelEnabled(p.level) {
		return
	}
	buf := p.bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		p.bufPool.Put(buf)
	}()
	for _, field := range p.fields {
		_, _ = fmt.Fprintf(buf, "%s=%v ", field.Key, Value(p.ctx, field.Value))
	}
	buf.WriteString(fmt.Sprintln(v...))
	buf.Truncate(buf.Len() - 1)
	_ = p.logger.out.Output(2, buf.String())
}

func (p *printer) With(key string, value interface{}) Printer {
	if key == "" {
		return p
	}
	for i := 0; i < len(p.fields); i++ {
		if p.fields[i].Key == key {
			p.fields[i].Value = value
			return p
		}
	}
	p.fields = append(p.fields, field{key, value})
	return p
}

// logger is the builtin implementation of Logger
type logger struct {
	out     *log.Logger
	name    string
	manager *manager
}

// level return the lowest Level the logger supports.
func (l *logger) level() Level {
	name := l.name
	for name != "" {
		if val, ok := l.manager.levels.Get(name); ok {
			return val.(Level)
		}
		lastIndex := strings.LastIndexFunc(name, func(r rune) bool {
			return r == '.' || r == '/'
		})
		if lastIndex == -1 {
			break
		}
		name = name[:lastIndex]
	}
	rootLevel, _ := l.manager.levels.Get("")
	return rootLevel.(Level)
}

func (l *logger) LevelEnabled(level Level) bool {
	ll := l.level()
	return ll != ClosedLevel && ll <= level
}

func (l *logger) Printer(ctx context.Context) Printer {
	return l.AtLevel(l.level(), ctx)
}

func (l *logger) Name() string {
	return l.name
}

func (l *logger) AtLevel(level Level, ctx context.Context) Printer {
	return &printer{
		logger: l,
		level:  level,
		fields: []field{
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

// manager is the builtin implementation of Manager.
type manager struct {
	levels    *hashmap.HashMap
	out       *log.Logger
	instances *hashmap.HashMap
}

func (m *manager) Get(name string) Logger {
	actual, _ := m.instances.GetOrInsert(name, &logger{
		out:     m.out,
		name:    name,
		manager: m,
	})
	return actual.(*logger)
}

func (m *manager) Level(name string, level Level) {
	m.levels.Set(name, level)
}

// Option is options for NewManager.
type Option func(m *manager)

// LoggerLevel produce one Option which set Level of name.
func LoggerLevel(name string, level Level) Option {
	return func(m *manager) {
		m.Level(name, level)
	}
}

// RootLevel produce one Option which set Level of root logger.
func RootLevel(level Level) Option {
	return func(m *manager) {
		m.Level("", level)
	}
}

// NewManager build and return a Manager.
func NewManager(out *log.Logger, opts ...Option) Manager {
	mng := &manager{
		levels:    &hashmap.HashMap{},
		out:       out,
		instances: &hashmap.HashMap{},
	}
	mng.Level("", InfoLevel)
	for _, o := range opts {
		o(mng)
	}
	return mng
}
