package log

import (
	"testing"
	"unsafe"
)

func resetLevelNames() {
	_levelNames.Swap(unsafe.Pointer(&map[Level]string{
		DebugLevel: "DEBUG",
		InfoLevel:  "INFO",
		WarnLevel:  "WARN",
		ErrorLevel: "ERROR",
	}))
}

func TestRegisterLevelName(t *testing.T) {
	resetLevelNames()
	r1 := *(*map[Level]string)(_levelNames.Load())
	RegisterLevelName(Level(99), "L99")
	r2 := *(*map[Level]string)(_levelNames.Load())
	if len(r1) == len(r2) {
		t.Errorf("expect size not equals, but equals")
	}
	if _, ok := r1[Level(99)]; ok {
		t.Errorf("old levelNames should not contains L99")
	}
	if _, ok := r2[Level(99)]; !ok {
		t.Errorf("new levelNames should contains L99")
	}
}

func TestLevel_String(t *testing.T) {
	defer resetLevelNames()
	resetLevelNames()
	if InfoLevel.String() != "INFO" {
		t.Errorf("string of InfoLevel is not INFO!")
	}
	if Level(99).String() != "Level(99)" {
		t.Errorf("string of unregister Level(99) should be \"Level(99)\", but is %s",
			Level(99).String())
	}
	RegisterLevelName(DebugLevel, "")
	if DebugLevel.String() != "Level(-1)" {
		t.Errorf("after deregister, string of DebugLevel should be \"Level(-1)\", but is %s",
			DebugLevel.String())
	}
	RegisterLevelName(Level(99), "HELP")
	if Level(99).String() != "HELP" {
		t.Errorf("after register, string of Level(99) should be \"HELP\", but is %s",
			Level(99).String())
	}
}
