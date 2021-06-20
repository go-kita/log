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

type nopLogger struct {
}

var _ Logger = (*nopLogger)(nil)

func (l *nopLogger) AtLevel(_ context.Context, _ Level) Printer {
	return NewNopPrinter()
}

func nopProvider(_ string) Logger {
	return &nopLogger{}
}

func getLoggerProvider() LoggerProvider {
	fp := (*LoggerProvider)(_loggerProvider.Load())
	if fp == nil {
		return nil
	}
	return *fp
}

func TestUseProvider(t *testing.T) {
	op := getLoggerProvider()
	defer UseProvider(op)
	UseProvider(nopProvider)
	if getLoggerProvider() == nil {
		t.Errorf("expect the underlying LoggerProvider not nil, but is nil")
	}
	log := getLoggerProvider()("test")
	if _, ok := log.(*nopLogger); !ok {
		t.Errorf("expect a nopLogger, but not: %T", log)
	}
	UseProvider(nil)
	if getLoggerProvider() != nil {
		t.Errorf("expect the underlying LoggerProvider be nil, but not nil")
	}
}

func TestGet(t *testing.T) {
	op := getLoggerProvider()
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
