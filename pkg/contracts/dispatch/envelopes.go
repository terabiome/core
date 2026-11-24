package dispatch

import "encoding/json"

type DispatchEnvelope struct {
	Metadata map[string]any  `json:"metadata"`
	Payload  json.RawMessage `json:"payload"`
}
