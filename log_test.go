package log

import (
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
