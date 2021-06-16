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

func buildStdPrinter(w io.Writer, ctx context.Context) *stdPrinter {
	return &stdPrinter{
		output: NewStdOutput(log.New(w, "", 0)),
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
	printer := buildStdPrinter(w, context.Background())
	printer.Print("1", "2", "3")
	expect := "level=INFO logger= 123\n"
	actual := w.String()
	if expect != actual {
		t.Errorf("expect %s, got %s", expect, actual)
	}
}

func TestStdPrinter_Printf(t *testing.T) {
	w := &bytes.Buffer{}
	printer := buildStdPrinter(w, context.Background())
	printer.Printf("%d %s", 1, "abc")
	expect := "level=INFO logger= 1 abc\n"
	actual := w.String()
	if expect != actual {
		t.Errorf("expect %s, got %s", expect, actual)
	}
}

func TestStdPrinter_Println(t *testing.T) {
	w := &bytes.Buffer{}
	printer := buildStdPrinter(w, context.Background())
	printer.Println(1, "abc", true)
	expect := "level=INFO logger= 1 abc true\n"
	actual := w.String()
	if expect != actual {
		t.Errorf("expect %s, got %s", expect, actual)
	}
}

func TestStdPrinter_With(t *testing.T) {
	w := &bytes.Buffer{}
	printer := buildStdPrinter(w, context.Background())
	printer.With("module", "test").Print("abc")
	expect := "level=INFO logger= module=test abc\n"
	got := w.String()
	if expect != got {
		t.Errorf("expect %q, got %q", expect, got)
	}
	w.Reset()
	printer = buildStdPrinter(w, context.Background())
	printer.With("", "ignored").Printf("abc")
	expect = "level=INFO logger= abc\n"
	got = w.String()
	if expect != got {
		t.Errorf("expect %q, got %q", expect, got)
	}
	w.Reset()
	printer = buildStdPrinter(w, context.Background())
	printer.With("module", "test1").Printf("abc")
	expect = "level=INFO logger= module=test1 abc\n"
	got = w.String()
	if expect != got {
		t.Errorf("expect %q, got %q", expect, got)
	}
	w.Reset()
	printer = buildStdPrinter(w, context.Background())
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
		output: NewStdOutput(log.New(w, "", 0)),
		name:   name,
	}
}

func TestStdLogger_AtLevel(t *testing.T) {
	w := &bytes.Buffer{}
	ls := NewStdLevelStore(
		RootLevel(WarnLevel),
		LoggerLevel("pkg", DebugLevel),
	)
	ols := GetLevelStore()
	defer UseLevelStore(ols)
	UseLevelStore(ls)
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

func TestLogger_AtLevel(t *testing.T) {
	w := &bytes.Buffer{}
	ls := NewStdLevelStore(
		LoggerLevel("closed", ClosedLevel),
	)
	ols := GetLevelStore()
	defer UseLevelStore(ols)
	UseLevelStore(ls)
	root := buildStdLogger("", w)
	w.Reset()
	root.AtLevel(DebugLevel, context.Background()).Print("print nothing")
	if w.Len() != 0 {
		t.Errorf("should print nothing")
	}
	w.Reset()
	root.AtLevel(WarnLevel, context.Background()).Print("HAHAHA")
	expect := "level=WARN logger= HAHAHA\n"
	if expect != w.String() {
		t.Errorf("expect %q, got %q", expect, w.String())
	}

	w.Reset()
	closed := buildStdLogger("closed", w)
	closed.AtLevel(InfoLevel, context.Background()).Print("closed")
	closed.AtLevel(InfoLevel, context.Background()).Printf("closed %s", "nothing")
	closed.AtLevel(InfoLevel, context.Background()).Println("closed %s", "nothing")
	closed.AtLevel(ErrorLevel, context.Background()).Println("abc")
	if w.Len() != 0 {
		t.Errorf("should print nothing, got %q", w.String())
	}

	ls.UnSet("closed")
	w.Reset()
	closed.AtLevel(InfoLevel, context.Background()).Println("hello")
	expect = "level=INFO logger=closed hello\n"
	got := w.String()
	if expect != got {
		t.Errorf("expect %q, got %q", expect, got)
	}
}

func TestCaller(t *testing.T) {
	w := &bytes.Buffer{}
	logger := NewStdLogger("", NewStdOutput(log.New(w, "", log.Lshortfile)))
	logger.AtLevel(InfoLevel, context.Background()).Print()
	got := w.String()
	if !strings.Contains(got, "_test.go") {
		t.Errorf("expect contains current test file name, but got: %q", got)
	}
}
