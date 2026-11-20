package helpers

import "encoding/json"

type jsonHelpers struct{}

func JSON() jsonHelpers {
	return jsonHelpers{}
}

func (h jsonHelpers) Unmarshal[T any](data []byte) (T, error) {
	var v T
	err := json.Unmarshal(data, &v)
	return v, err
}
