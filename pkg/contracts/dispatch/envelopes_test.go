package dispatch_test

import (
	"encoding/json"
	"testing"

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
