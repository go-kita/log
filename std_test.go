package log

import (
	"bytes"
	"context"
	"io"
	"log"
	"strings"
	"sync"
	"testing"
)

func TestNewStdOutPutter(t *testing.T) {
	outPutter := NewStdOutPutter(nil)
	outPutter.OutPut(
		context.Background(),
		"",
		InfoLevel,
		"abc",
		[]Field{
			{LevelKey, InfoLevel},
			{LoggerKey, ""},
		},
		3)
}

func TestOutPutFilter_OutPut(t *testing.T) {
	o := &OutPutFilter{}
	o.OutPut(context.Background(), "", InfoLevel, "abc", nil, 0)
}

func TestFilterEnable(t *testing.T) {
	w := &bytes.Buffer{}
	o := NewStdOutPutter(log.New(w, "", 0))
	o = FilterEnable(o, func(ctx context.Context, name string, level Level) bool {
		return name != "sensitive" && level >= WarnLevel
	})
	o.OutPut(context.Background(), "abc", ErrorLevel, "something wrong", nil, 3)
	got := w.String()
	expect := "something wrong\n"
	if got != expect {
		t.Errorf("expect %q, got %q", expect, got)
	}
	w.Reset()
	o.OutPut(context.Background(), "sensitive", WarnLevel, "it's a secret.", nil, 3)
	got = w.String()
	expect = ""
	if got != expect {
		t.Errorf("expect %q, got %q", expect, got)
	}
	w.Reset()
	o.OutPut(context.Background(), "xyz", InfoLevel, "hello", nil, 3)
	expect = ""
	got = w.String()
	if got != expect {
		t.Errorf("expect %q, got %q", expect, got)
	}
	o = FilterEnable(nil, nil)
	if o != nil {
		t.Errorf("expect nil, but not")
	}
}

func TestFilterRemoveField(t *testing.T) {
	w := &bytes.Buffer{}
	o := NewStdOutPutter(log.New(w, "", 0))
	o = FilterRemoveField(o, "passwd")
	o.OutPut(
		context.Background(),
		"",
		InfoLevel,
		"logged in...",
		[]Field{
			{"username", "mike"},
			{"passwd", "dwssap"},
		},
		3)
	got := w.String()
	expect := "username=mike logged in...\n"
	if got != expect {
		t.Errorf("expect %q, got %q", expect, got)
	}
	o = FilterRemoveField(nil, "abc")
	if o != nil {
		t.Errorf("expect nil, but not")
	}
}

func TestFilterCoverField(t *testing.T) {
	w := &bytes.Buffer{}
	o := NewStdOutPutter(log.New(w, "", 0))
	o = FilterCoverField(o, "passwd", "***")
	o.OutPut(
		context.Background(),
		"",
		InfoLevel,
		"logged in...",
		[]Field{
			{"username", "mike"},
			{"passwd", "dwssap"},
		},
		3)
	got := w.String()
	expect := "username=mike passwd=*** logged in...\n"
	if got != expect {
		t.Errorf("expect %q, got %q", expect, got)
	}
	o = FilterCoverField(nil, "abc", "...")
	if o != nil {
		t.Errorf("expect nil, but not")
	}
}

func buildStdPrinter(ctx context.Context, w io.Writer) *stdPrinter {
	return &stdPrinter{
		logger: buildStdLogger("", w),
		level:  InfoLevel,
		fields: []Field{
			{LevelKey, InfoLevel},
			{LoggerKey, ""},
		},
		ctx: ctx,
		bufPool: &sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
}

func TestStdPrinter_Print(t *testing.T) {
	w := &bytes.Buffer{}
	printer := buildStdPrinter(context.Background(), w)
	printer.Print("1", "2", "3")
	expect := "level=INFO logger= 123\n"
	actual := w.String()
	if expect != actual {
		t.Errorf("expect %s, got %s", expect, actual)
	}
}

func TestStdPrinter_Printf(t *testing.T) {
	w := &bytes.Buffer{}
	printer := buildStdPrinter(context.Background(), w)
	printer.Printf("%d %s", 1, "abc")
	expect := "level=INFO logger= 1 abc\n"
	actual := w.String()
	if expect != actual {
		t.Errorf("expect %s, got %s", expect, actual)
	}
}

func TestStdPrinter_Println(t *testing.T) {
	w := &bytes.Buffer{}
	printer := buildStdPrinter(context.Background(), w)
	printer.Println(1, "abc", true)
	expect := "level=INFO logger= 1 abc true\n"
	actual := w.String()
	if expect != actual {
		t.Errorf("expect %s, got %s", expect, actual)
	}
}

func TestStdPrinter_With(t *testing.T) {
	w := &bytes.Buffer{}
	printer := buildStdPrinter(context.Background(), w)
	printer.With("module", "test").Print("abc")
	expect := "level=INFO logger= module=test abc\n"
	got := w.String()
	if expect != got {
		t.Errorf("expect %q, got %q", expect, got)
	}
	w.Reset()
	printer = buildStdPrinter(context.Background(), w)
	printer.With("", "ignored").Printf("abc")
	expect = "level=INFO logger= abc\n"
	got = w.String()
	if expect != got {
		t.Errorf("expect %q, got %q", expect, got)
	}
	w.Reset()
	printer = buildStdPrinter(context.Background(), w)
	printer.With("module", "test1").Printf("abc")
	expect = "level=INFO logger= module=test1 abc\n"
	got = w.String()
	if expect != got {
		t.Errorf("expect %q, got %q", expect, got)
	}
	w.Reset()
	printer = buildStdPrinter(context.Background(), w)
	printer.
		With("module", "test1").
		With("module", "test2").
		Printf("abc")
	expect = "level=INFO logger= module=test2 abc\n"
	got = w.String()
	if expect != got {
		t.Errorf("expect %q, got %q", expect, got)
	}
}

func buildStdLogger(name string, w io.Writer) *stdLogger {
	return &stdLogger{
		output: NewStdOutPutter(log.New(w, "", 0)),
		name:   name,
	}
}

func TestStdLogger_LevelEnabled(t *testing.T) {
	w := &bytes.Buffer{}
	store := GetLevelStore()
	store.Set("", WarnLevel).Set("pkg", DebugLevel)
	defer func() {
		store.Set("", InfoLevel).UnSet("pkg")
	}()
	sub := buildStdLogger("pkg/sub", w)
	if !sub.levelEnabled(DebugLevel) {
		t.Errorf("expect logger %s enabled DebugLevel as it parent, but not enabled",
			"pkg/sub")
	}
	xyz := buildStdLogger("xyz", w)
	if !xyz.levelEnabled(WarnLevel) {
		t.Errorf("expect logger %s enabled WarnLevel as root, but not enabled", "xyz")
	}
	if xyz.levelEnabled(InfoLevel) {
		t.Errorf("expect logger %s not enabled InfoLevel as root, but enabled", "xyz")
	}
}

func TestStdLogger_AtLevel(t *testing.T) {
	w := &bytes.Buffer{}
	GetLevelStore().Set("closed", ClosedLevel)
	defer func() {
		GetLevelStore().UnSet("closed")
	}()
	root := buildStdLogger("", w)
	w.Reset()
	root.AtLevel(context.Background(), DebugLevel).Print("print nothing")
	if w.Len() != 0 {
		t.Errorf("should print nothing")
	}
	w.Reset()
	root.AtLevel(context.Background(), WarnLevel).Print("HAHAHA")
	expect := "level=WARN logger= HAHAHA\n"
	if expect != w.String() {
		t.Errorf("expect %q, got %q", expect, w.String())
	}

	w.Reset()
	closed := buildStdLogger("closed", w)
	closed.AtLevel(context.Background(), InfoLevel).Print("closed")
	closed.AtLevel(context.Background(), InfoLevel).Printf("closed %s", "nothing")
	closed.AtLevel(context.Background(), InfoLevel).Println("closed %s", "nothing")
	closed.AtLevel(context.Background(), ErrorLevel).Println("abc")
	if w.Len() != 0 {
		t.Errorf("should print nothing, got %q", w.String())
	}

	GetLevelStore().UnSet("closed")
	w.Reset()
	closed.AtLevel(context.Background(), InfoLevel).Println("hello")
	expect = "level=INFO logger=closed hello\n"
	got := w.String()
	if expect != got {
		t.Errorf("expect %q, got %q", expect, got)
	}

	p := root.AtLevel(nil, InfoLevel)
	sp := p.(*stdPrinter)
	if sp.ctx == nil {
		t.Errorf("expect sp.ctx not nil but got nil")
	}
}

func TestCaller(t *testing.T) {
	w := &bytes.Buffer{}
	logger := NewStdLogger("", NewStdOutPutter(log.New(w, "", log.Lshortfile)))
	logger.AtLevel(context.Background(), InfoLevel).Print()
	got := w.String()
	if !strings.Contains(got, "_test.go") {
		t.Errorf("expect contains current test file name, but got: %q", got)
	}
	nl := NewStdLogger("", nil)
	sl := nl.(*stdLogger)
	if sl.output == nil {
		t.Errorf("expect sl.output not nil, but is nil")
	}
}
