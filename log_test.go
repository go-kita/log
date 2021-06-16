package log

import (
	"context"
	"testing"
)

func TestNopPrinter_Print(t *testing.T) {
	p := NewNopPrinter()
	p.Print("abc")
}

func TestNopPrinter_Printf(t *testing.T) {
	p := NewNopPrinter()
	p.Printf("%s %d", "abc", 123)
}

func TestNopPrinter_Println(t *testing.T) {
	p := NewNopPrinter()
	p.Println("abc", 123)
}

func TestNopPrinter_With(t *testing.T) {
	p := NewNopPrinter()
	_ = p.With("key", "value")
}

func TestNopLogger_AtLevel(t *testing.T) {
	l := NewNopLogger()
	p := l.AtLevel(InfoLevel, context.Background())
	if _, ok := p.(*nopPrinter); !ok {
		t.Errorf("expect a nopPrinter, but not")
	}
}

func nopProvider(_ string) Logger {
	return NewNopLogger()
}

func TestUseProvider(t *testing.T) {
	op := defaultLoggerProvider
	defer UseProvider(op)
	defaultLoggerProvider = nil
	UseProvider(nopProvider)
	if defaultLoggerProvider == nil {
		t.Errorf("expect the underlying LoggerProvider not nil, but is nil")
	}
	log := defaultLoggerProvider("test")
	if _, ok := log.(*nopLogger); !ok {
		t.Errorf("expect a nopLogger, but not: %T", log)
	}
	UseProvider(nil)
	if defaultLoggerProvider != nil {
		t.Errorf("expect the underlying LoggerProvider be nil, but not nil")
	}
}

func TestUseLevelStore(t *testing.T) {
	ols := defaultLevelStore
	defer UseLevelStore(ols)
	UseLevelStore(nil)
	if defaultLevelStore != nil {
		t.Errorf("expect the underlying LevelStore be nil, but not nil")
	}
}

func TestNoLevelStore(t *testing.T) {
	ols := defaultLevelStore
	defer UseLevelStore(ols)
	NoLevelStore()
	if defaultLevelStore != nil {
		t.Errorf("expect the underlying LevelStore be nil, but not nil")
	}
}

func TestGet(t *testing.T) {
	op := defaultLoggerProvider
	defer UseProvider(op)
	logger := Get("abc")
	if logger == nil {
		t.Errorf("expect not nil, but got nil")
	}
	UseProvider(nil)
	logger = Get("xyz")
	if logger == nil {
		t.Errorf("expect not nil, but got nil")
	}
}
