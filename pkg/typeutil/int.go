// Package typeutil provides small type conversion helpers shared across Chronix.
package typeutil

// AsInt64 converts common numeric scalar and pointer values to int64.
func AsInt64(v any) int64 {
	switch t := v.(type) {
	case int64:
		return t
	case *int64:
		if t != nil {
			return *t
		}
	case int32:
		return int64(t)
	case *int32:
		if t != nil {
			return int64(*t)
		}
	case int:
		return int64(t)
	case *int:
		if t != nil {
			return int64(*t)
		}
	case float64:
		return int64(t)
	case *float64:
		if t != nil {
			return int64(*t)
		}
	case float32:
		return int64(t)
	case *float32:
		if t != nil {
			return int64(*t)
		}
	}
	return 0
}
