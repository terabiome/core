package dispatch

import (
	"encoding/json"
	"errors"
	"fmt"

	manifestcontracts "github.com/terabiome/core/pkg/contracts/manifest"
)

type DispatchEnvelope struct {
	Metadata map[string]any  `json:"metadata"`
	Payload  json.RawMessage `json:"payload"`
}

// payload only holds only the sequences of a UOW, not the full manifest.
func (e DispatchEnvelope) Sequences() ([]manifestcontracts.UnitOfWorkSequence, error) {
	if len(e.Payload) == 0 {
		return nil, errors.New("dispatch envelope has no payload")
	}

	var sequences []manifestcontracts.UnitOfWorkSequence
	if err := json.Unmarshal(e.Payload, &sequences); err != nil {
		return nil, fmt.Errorf("failed to decode dispatch payload: %w", err)
	}
	if len(sequences) == 0 {
		return nil, errors.New("dispatch payload has no sequences")
	}
	return sequences, nil
}
