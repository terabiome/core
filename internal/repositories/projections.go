package repositories

import (
	"encoding/json"
	"uuid"

	"github.com/terabiome/core/internal/models"
)

type DispatchFilterRecord struct {
	UUID     uuid.UUID
	UOWUUID  uuid.UUID
	Target   json.RawMessage
	Metadata json.RawMessage
	Payload  json.RawMessage
}

type DispatchFilterOutput struct {
	Records []DispatchFilterRecord
	Count   int
}

type DispatchUpdateFilters struct {
	UUID *uuid.UUID
}

type DispatchUpdateFields struct {
	Status   models.DispatchStatus
	Target   *json.RawMessage
	Metadata *json.RawMessage
	Payload  *json.RawMessage
}

type DispatchUpdateInput struct {
	Filters DispatchUpdateFilters
	Fields  DispatchUpdateFields
}

type DispatchUpdateOutput struct {
}

// // TODO: implement
// type DispatchBulkUpdateInput struct {
// }

// // TODO: implement
// type DispatchBulkUpdateOutput struct {
// }
