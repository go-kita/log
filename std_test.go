package log

import (
	"bytes"
	"context"
	stdlog "log"
	"testing"
)

func TestPrinter_Print(t *testing.T) {
	w := &bytes.Buffer{}
	mng := NewManager(stdlog.New(w, "", 0))
	mng.Get("").Printer(context.Background()).Print("1", "2", "3")
	expect := "level=INFO logger= 123\n"
	actual := w.String()
	if expect != actual {
		t.Errorf("expect %s, got %s", expect, actual)
	}
}

func TestPrinter_Printf(t *testing.T) {
	w := &bytes.Buffer{}
	mng := NewManager(stdlog.New(w, "", 0))
	mng.Get("").Printer(context.Background()).Printf("%d %s", 1, "abc")
	expect := "level=INFO logger= 1 abc\n"
	actual := w.String()
	if expect != actual {
		t.Errorf("expect %s, got %s", expect, actual)
	}
}

func TestPrinter_Println(t *testing.T) {
	w := &bytes.Buffer{}
	mng := NewManager(stdlog.New(w, "", 0))
	mng.Get("").Printer(context.Background()).Println(1, "abc", true)
	expect := "level=INFO logger= 1 abc true\n"
	actual := w.String()
	if expect != actual {
		t.Errorf("expect %s, got %s", expect, actual)
	}
}

func TestPrinter_EmptyKey(t *testing.T) {
	w := &bytes.Buffer{}
	mng := NewManager(stdlog.New(w, "", 0))
	printer := mng.Get("").Printer(context.Background()).With("", "abc")
	printer.Print("abc")
	expect := "level=INFO logger= abc\n"
	actual := w.String()
	if expect != actual {
		t.Errorf("expect %q, got %q", expect, actual)
	}
	w.Reset()
	printer.Printf("%s", "abc")
	actual = w.String()
	if expect != actual {
		t.Errorf("expect %q, got %q", expect, actual)
	}
	w.Reset()
	printer.Println("abc")
	actual = w.String()
	if expect != actual {
		t.Errorf("expect %q, got %q", expect, actual)
	}
}

func TestPrinter_With(t *testing.T) {
	w := &bytes.Buffer{}
	mng := NewManager(stdlog.New(w, "", 0))
	logger := mng.Get("")
	logger.Printer(context.Background()).With("module", "test").Print("abc")
	expect := "level=INFO logger= module=test abc\n"
	got := w.String()
	if expect != got {
		t.Errorf("expect %q, got %q", expect, got)
	}
	w.Reset()
	logger.Printer(context.Background()).With("", "ignored").Printf("abc")
	expect = "level=INFO logger= abc\n"
	got = w.String()
	if expect != got {
		t.Errorf("expect %q, got %q", expect, got)
	}
	w.Reset()
	logger.Printer(context.Background()).With("module", "test1").Printf("abc")
	expect = "level=INFO logger= module=test1 abc\n"
	got = w.String()
	if expect != got {
		t.Errorf("expect %q, got %q", expect, got)
	}
	w.Reset()
	logger.Printer(context.Background()).
		With("module", "test1").
		With("module", "test2").
		Printf("abc")
	expect = "level=INFO logger= module=test2 abc\n"
	got = w.String()
	if expect != got {
		t.Errorf("expect %q, got %q", expect, got)
	}
}

func TestLogger_level(t *testing.T) {
	w := &bytes.Buffer{}
	mng := NewManager(stdlog.New(w, "", 0),
		RootLevel(WarnLevel),
		LoggerLevel("pkg", DebugLevel),
	)
	sub := mng.Get("pkg/sub")
	if !sub.LevelEnabled(DebugLevel) {
		t.Errorf("expect logger %s enabled DebugLevel as it parent, but not enabled",
			"pkg/sub")
	}
	xyz := mng.Get("xyz")
	if !xyz.LevelEnabled(WarnLevel) {
		t.Errorf("expect logger %s enabled WarnLevel as root, but not enabled", "xyz")
	}
	if xyz.LevelEnabled(InfoLevel) {
		t.Errorf("expect logger %s not enabled InfoLevel as root, but enabled", "xyz")
	}
}

func TestLogger_LevelEnabled(t *testing.T) {
	w := &bytes.Buffer{}
	mng := NewManager(stdlog.New(w, "", 0), LoggerLevel("info", InfoLevel))
	info := mng.Get("info")
	if info.LevelEnabled(DebugLevel) {
		t.Errorf("An InfoLevel logger should not enable DebugLevel")
	}
	if !info.LevelEnabled(InfoLevel) {
		t.Errorf("An InfoLevel logger should enable InfoLevel")
	}
	if !info.LevelEnabled(WarnLevel) {
		t.Errorf("An InfoLevel logger should enable WarnLevel")
	}
	info.Printer(context.Background()).Println("abc")
	expect := "level=INFO logger=info abc\n"
	got := w.String()
	if expect != got {
		t.Errorf("expect %q, got %q", expect, got)
	}
}

func TestLogger_Name(t *testing.T) {
	mng := NewManager(stdlog.Default())
	name1 := mng.Get("abc").Name()
	if name1 != "abc" {
		t.Errorf("expect %s, got %s", "abc", name1)
	}
	name2 := mng.Get("").Name()
	if name2 != "" {
		t.Errorf("expect %s, got %s", "", name2)
	}
}

func TestLogger_AtLevel(t *testing.T) {
	w := &bytes.Buffer{}
	mng := NewManager(stdlog.New(w, "", 0),
		LoggerLevel("closed", ClosedLevel),
	)
	info := mng.Get("")
	w.Reset()
	info.AtLevel(DebugLevel, context.Background()).Print("print nothing")
	if w.Len() != 0 {
		t.Errorf("should print nothing")
	}
	w.Reset()
	info.AtLevel(WarnLevel, context.Background()).Print("HAHAHA")
	expect := "level=WARN logger= HAHAHA\n"
	if expect != w.String() {
		t.Errorf("expect %q, got %q", expect, w.String())
	}

	w.Reset()
	closed := mng.Get("closed")
	closed.Printer(context.Background()).Print("closed")
	closed.Printer(context.Background()).Printf("closed %s", "nothing")
	closed.Printer(context.Background()).Println("closed %s", "nothing")
	closed.AtLevel(ErrorLevel, context.Background()).Println("abc")
	if w.Len() != 0 {
		t.Errorf("should print nothing, got %q", w.String())
	}
}

func TestManager_Level(t *testing.T) {
	w := &bytes.Buffer{}
	mng := NewManager(stdlog.New(w, "", 0))
	mng.Level("", WarnLevel)
	mng.Get("").Printer(context.Background()).Println("abc")
	expect := "level=WARN logger= abc\n"
	got := w.String()
	if expect != got {
		t.Errorf("expect %q, got %q", expect, got)
	}
}
