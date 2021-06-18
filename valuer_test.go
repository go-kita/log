package log

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

type testKey struct{}

func TestValue(t *testing.T) {
	var testValuer Valuer
	testValuer = func(ctx context.Context) interface{} {
		return ctx.Value(testKey{})
	}
	ctx := context.WithValue(context.Background(), testKey{}, "value")
	tests := []struct {
		v      interface{}
		expect interface{}
	}{
		{"abc", "abc"},
		{1, 1},
		{nil, nil},
		{testValuer, "value"},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("test%d", i), func(t *testing.T) {
			value := Value(ctx, test.v)
			if !reflect.DeepEqual(value, test.expect) {
				t.Errorf("expect %q, got %q", test.expect, value)
			}
		})
	}
}
