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
