package helpers

import "encoding/json"

type mapHelpers struct {
}

type mapType interface {
	map[string]any // currently support this
}

func Map() mapHelpers {
	return mapHelpers{}
}

func (h mapHelpers) Merge[T mapType](base T, inputs ...T) T {
	if base == nil {
		base = T{}
	}

	for _, input := range inputs {
		for k, v := range input {
			base[k] = v
		}
	}
	return base
}

func (h mapHelpers) ImmutableMerge[T mapType](inputs ...T) T {
	return h.Merge(T{}, inputs...)
}

// can generalize this
func (h mapHelpers) GetString[T mapType](input T, key string) (string, bool, bool) {
	return h.get[T, string](input, key)
}

// GetInt64 reads a number stored in a map decoded by encoding/json.
// JSON numbers come back as float64, not int64.
func (h mapHelpers) GetInt64[T mapType](input T, key string) (int64, bool, bool) {
	raw, isSet := input[key]
	if !isSet {
		return 0, false, false
	}

	switch n := raw.(type) {
	case int64:
		return n, true, true
	case int:
		return int64(n), true, true
	case float64:
		return int64(n), true, true
	case json.Number:
		i, err := n.Int64()
		if err != nil {
			return 0, true, false
		}
		return i, true, true
	default:
		return 0, true, false
	}
}

// rule: can't read -> can't cast, shortcut usage involves reading the 3rd result
func (h mapHelpers) get[T mapType, V any](input T, key string) (V, bool, bool) {
	var val V
	rawVal, ok := input[key]
	if !ok {
		return val, false, false
	}
	val, canCast := rawVal.(V)
	return val, true, canCast
}
