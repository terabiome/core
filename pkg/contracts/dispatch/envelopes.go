package dispatch

import (
	"encoding/json"
	"errors"
	"fmt"
	"uuid"

	manifestcontracts "github.com/terabiome/core/pkg/contracts/manifest"
)

type DispatchEnvelope struct {
	Metadata map[string]any  `json:"metadata"`
	Payload  json.RawMessage `json:"payload"`
}

type Identity struct {
	UUID    uuid.UUID
	UOWUUID uuid.UUID
}

func (e DispatchEnvelope) Identity() (Identity, error) {
	dispatchUUID, err := metadataUUID(e.Metadata, "dispatch_uuid")
	if err != nil {
		return Identity{}, err
	}

	uowUUID, err := metadataUUID(e.Metadata, "uow_uuid")
	if err != nil {
		return Identity{}, err
	}

	return Identity{UUID: dispatchUUID, UOWUUID: uowUUID}, nil
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

func metadataUUID(metadata map[string]any, key string) (uuid.UUID, error) {
	value, ok := metadata[key]
	if !ok {
		return uuid.UUID{}, fmt.Errorf("dispatch metadata has no %s", key)
	}

	switch v := value.(type) {
	case uuid.UUID:
		return v, nil
	case string:
		id, err := uuid.Parse(v)
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("dispatch metadata %s is not a uuid: %w", key, err)
		}
		return id, nil
	default:
		return uuid.UUID{}, fmt.Errorf("dispatch metadata %s has unexpected type %T", key, value)
	}
}
