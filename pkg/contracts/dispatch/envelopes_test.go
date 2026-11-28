package dispatch_test

import (
	"encoding/json"
	"testing"
	"uuid"

	"github.com/terabiome/core/pkg/contracts/dispatch"
)

func TestPublicDispatchEnvelope(t *testing.T) {
	t.Parallel()

	raw := []byte(`{"metadata":{"uow_uuid":"x"},"payload":[{"name":"os-hardening"}]}`)

	var envelope dispatch.DispatchEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("json.Unmarshal() = %v, want nil", err)
	}
	if len(envelope.Payload) == 0 {
		t.Fatal("payload is empty")
	}
	if envelope.Metadata["uow_uuid"] != "x" {
		t.Fatalf("metadata[uow_uuid] = %v, want x", envelope.Metadata["uow_uuid"])
	}
}

func TestDispatchEnvelopeIdentity(t *testing.T) {
	t.Parallel()

	want := dispatch.Identity{
		UUID:    uuid.MustParse("018f0000-0000-7000-8000-000000000001"),
		UOWUUID: uuid.MustParse("018f0000-0000-7000-8000-000000000002"),
	}
	raw := []byte(`{"metadata":{"dispatch_uuid":"018f0000-0000-7000-8000-000000000001","uow_uuid":"018f0000-0000-7000-8000-000000000002"},"payload":[{"name":"smoke"}]}`)

	var envelope dispatch.DispatchEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("json.Unmarshal() = %v", err)
	}

	got, err := envelope.Identity()
	if err != nil {
		t.Fatalf("Identity() = %v", err)
	}
	if got != want {
		t.Fatalf("Identity() = %+v, want %+v", got, want)
	}
}

func TestDispatchEnvelopeIdentityRejectsBadMetadata(t *testing.T) {
	t.Parallel()

	tests := map[string]map[string]any{
		"missing dispatch_uuid": {"uow_uuid": "018f0000-0000-7000-8000-000000000002"},
		"missing uow_uuid":      {"dispatch_uuid": "018f0000-0000-7000-8000-000000000001"},
		"invalid dispatch_uuid": {"dispatch_uuid": "nope", "uow_uuid": "018f0000-0000-7000-8000-000000000002"},
		"wrong type":            {"dispatch_uuid": 42, "uow_uuid": "018f0000-0000-7000-8000-000000000002"},
	}

	for name, metadata := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := (dispatch.DispatchEnvelope{Metadata: metadata}).Identity(); err == nil {
				t.Fatal("Identity() = nil, want error")
			}
		})
	}
}
