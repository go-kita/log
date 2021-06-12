package log

import "context"

// Valuer is function that calculating real value at call time.
// Note that the ctx may be nil.
type Valuer func(ctx context.Context) interface{}

// Value calculate and return the value if v is a Valuer, or just return v.
func Value(ctx context.Context, v interface{}) interface{} {
	for {
		if valuer, ok := v.(Valuer); ok {
			v = valuer(ctx)
			continue
		}
		break
	}
	return v
}
