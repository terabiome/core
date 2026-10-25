package models

import (
	"encoding/json"
	"time"
	"uuid"
)

// UnitOfWork represents the master plan row.
type UnitOfWork struct {
	UUID       uuid.UUID
	Name       string
	Status     string
	ConfigJSON json.RawMessage
	CreatedAt  time.Time
	UpdatedAt  time.Time
	FinishedAt *time.Time
}

// Dispatch represents a per-target assignment.
type Dispatch struct {
	UUID      uuid.UUID
	UOWUUID   uuid.UUID
	Status    string
	Metadata  json.RawMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Sequence represents a materialized sequence for a specific dispatch.
type Sequence struct {
	UUID         uuid.UUID
	DispatchUUID uuid.UUID
	Status       string
	Metadata     json.RawMessage
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Task represents a materialized task for a specific sequence.
type Task struct {
	UUID         uuid.UUID
	SequenceUUID uuid.UUID
	Status       string
	Metadata     json.RawMessage
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// for linking relationships
type CreationPlan struct {
	UOW        UnitOfWork
	Dispatches []DispatchPlan
}

type DispatchPlan struct {
	Dispatch  Dispatch
	Sequences []SequencePlan
}

type SequencePlan struct {
	Sequence Sequence
	Tasks    []Task
}
