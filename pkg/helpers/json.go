package helpers

import (
	"bytes"
	"encoding/json"
)

type jsonHelpers struct{}

func JSON() jsonHelpers {
	return jsonHelpers{}
}

func (h jsonHelpers) Unmarshal[T any](data []byte) (T, error) {
	var v T
	// Numbers that are not decoded into an integer field stay json.Number, not float64.
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	err := dec.Decode(&v)
	return v, err
}
