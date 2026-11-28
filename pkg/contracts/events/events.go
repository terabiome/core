package events

import (
	"encoding/json"
	"time"
	"uuid"
)

type Type string

const (
	TypeDispatchStart  Type = "dispatch.start"
	TypeDispatchFinish Type = "dispatch.finish"
)

const (
	TypeSequenceStart  Type = "sequence.start"
	TypeSequenceFinish Type = "sequence.finish"
)

const (
	TypeStepStart  Type = "step.start"
	TypeStepFinish Type = "step.finish"
)

type DispatchEvent struct {
	DispatchUUID uuid.UUID       `json:"dispatch_uuid"`
	UOWUUID      uuid.UUID       `json:"uow_uuid"`
	Type         Type            `json:"event_type"`
	Metadata     json.RawMessage `json:"metadata"`
	Data         json.RawMessage `json:"data"`
	GeneratedAt  time.Time       `json:"generated_at"`
}
