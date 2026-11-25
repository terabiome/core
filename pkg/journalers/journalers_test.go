package journalers

import (
	"sync"
	"testing"
	"time"
)

func TestJournalerBatchOutlivesFlush(t *testing.T) {
	t.Parallel()

	var batches [][]JournalEntry
	journaler, err := NewJournaler(JournalerParameters{
		MaxRecords:    1,
		FlushInterval: time.Hour,
		BatchFn: func(entries []JournalEntry) {
			batches = append(batches, entries)
		},
	})
	if err != nil {
		t.Fatalf("NewJournaler() = %v", err)
	}

	for _, message := range []string{"first", "second"} {
		if _, err := journaler.Write([]byte(message)); err != nil {
			t.Fatalf("Write(%q) = %v", message, err)
		}
	}
	if err := journaler.Close(); err != nil {
		t.Fatalf("Close() = %v", err)
	}

	if len(batches) != 2 {
		t.Fatalf("batches = %d, want 2", len(batches))
	}
	for i, want := range []string{"first", "second"} {
		if got := batches[i][0].Message; got != want {
			t.Fatalf("batch %d message = %q, want %q", i, got, want)
		}
	}
}

func TestJournalerCloseIsIdempotentAndStopsWrites(t *testing.T) {
	t.Parallel()

	journaler, err := NewJournaler(JournalerParameters{
		MaxRecords:    4,
		FlushInterval: time.Hour,
		BatchFn:       func([]JournalEntry) {},
	})
	if err != nil {
		t.Fatalf("NewJournaler() = %v", err)
	}

	if err := journaler.Close(); err != nil {
		t.Fatalf("Close() = %v", err)
	}
	if err := journaler.Close(); err != nil {
		t.Fatalf("second Close() = %v", err)
	}

	message := []byte("after close")
	if n, err := journaler.Write(message); n != len(message) || err != nil {
		t.Fatalf("Write() after close = (%d, %v), want (%d, nil)", n, err, len(message))
	}
	journaler.Log("after close", JournalerMetadata{Source: "test"})
}

func TestJournalerWriteDuringClose(t *testing.T) {
	t.Parallel()

	journaler, err := NewJournaler(JournalerParameters{
		MaxRecords:    8,
		FlushInterval: time.Hour,
		BatchFn:       func([]JournalEntry) {},
	})
	if err != nil {
		t.Fatalf("NewJournaler() = %v", err)
	}

	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				journaler.Write([]byte("message"))
			}
		}()
	}

	if err := journaler.Close(); err != nil {
		t.Fatalf("Close() = %v", err)
	}
	wg.Wait()
}
