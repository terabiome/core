package helpers

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

func (h mapHelpers) GetInt64[T mapType](input T, key string) (int64, bool, bool) {
	return h.get[T, int64](input, key)
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
