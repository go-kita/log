package log

import "testing"

func TestGetManager(t *testing.T) {
	mng, err := GetManager()
	if err != nil || mng == nil {
		t.Errorf("expect mng not nil and err nil, but got mng: %q, err: %q",
			mng, err)
	}
	defer UseManager(mng)
	UseManager(nil)
	mng, err = GetManager()
	if err == nil || mng != nil {
		t.Errorf("expect mng nil and err not nil, but got mng: %q, err: %q",
			mng, err)
	}
}

func TestGet(t *testing.T) {
	managed := Get("abc")
	mng, _ := GetManager()
	mng.Level("abc", InfoLevel)
	if !managed.LevelEnabled(InfoLevel) {
		t.Errorf("expect managed log %s enables InfoLevel, but not enabled", "abc")
	}
	mng.Level("abc", WarnLevel)
	if managed.LevelEnabled(InfoLevel) {
		t.Errorf("expect managed log %s not enables InfoLevel, but enabled", "abc")
	}
	defer UseManager(mng)
	UseManager(nil)
	free := Get("abc")
	UseManager(mng)
	mng.Level("abc", InfoLevel)
	if !free.LevelEnabled(DebugLevel) {
		t.Errorf("expect free log %s enables DebugLevel, but not enabled", "abc")
	}
}
