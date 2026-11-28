package journalers

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type BatchHandler func([]JournalEntry)

type JournalEntry struct {
	StreamUUID string
	Timestamp  time.Time
	Source     string
	Message    string
}

type JournalerParameters struct {
	MaxRecords    int
	FlushInterval time.Duration
	BatchFn       BatchHandler
}

func (p JournalerParameters) Validate() error {
	if p.MaxRecords < 1 {
		return fmt.Errorf("max_records must be >= 1")
	}
	if p.BatchFn == nil {
		return fmt.Errorf("batch handler cannot be nil")
	}
	return nil
}

type JournalerMetadata struct {
	StreamUUID string
	Source     string
}

const (
	journalIdle uint32 = iota
	journalRunning
	journalClosed
)

type Journaler struct {
	state         atomic.Uint32
	start         func()
	stop          func()
	stopCh        chan struct{}
	doneCh        chan struct{}
	inputCh       chan JournalEntry
	batchFn       BatchHandler
	maxRecords    int
	flushInterval time.Duration
}

func NewJournaler(params JournalerParameters) (*Journaler, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	tj := &Journaler{
		stopCh:        make(chan struct{}),
		doneCh:        make(chan struct{}),
		inputCh:       make(chan JournalEntry, 1<<10),
		maxRecords:    params.MaxRecords,
		flushInterval: params.FlushInterval,
		batchFn:       params.BatchFn,
	}
	tj.start = sync.OnceFunc(func() {
		if !tj.state.CompareAndSwap(journalIdle, journalRunning) {
			return
		}
		go tj.process()
	})
	tj.stop = sync.OnceFunc(func() {
		if tj.state.Swap(journalClosed) == journalRunning {
			close(tj.stopCh)
			<-tj.doneCh
		}
	})
	return tj, nil
}

func (t *Journaler) Start() {
	t.start()
}

func (t *Journaler) Write(p []byte) (n int, err error) {
	if len(p) == 0 || t.state.Load() == journalClosed {
		return len(p), nil
	}

	select {
	case t.inputCh <- JournalEntry{
		Timestamp: time.Now(),
		Source:    "stream",
		Message:   string(p),
	}:
	default:
	}
	return len(p), nil
}

func (t *Journaler) Log(msg string, metadata JournalerMetadata) {
	if t.state.Load() == journalClosed {
		return
	}

	select {
	case t.inputCh <- JournalEntry{
		Timestamp:  time.Now(),
		Message:    msg,
		StreamUUID: metadata.StreamUUID,
		Source:     metadata.Source,
	}:
	default:
	}
}

func (t *Journaler) process() {
	defer close(t.doneCh)

	ticker := time.NewTicker(t.flushInterval)
	defer ticker.Stop()

	buffer := make([]JournalEntry, 0, t.maxRecords)
	flush := func() {
		if len(buffer) == 0 {
			return
		}
		batch := buffer
		buffer = make([]JournalEntry, 0, t.maxRecords)
		t.batchFn(batch)
	}

	for {
		select {
		case entry := <-t.inputCh:
			buffer = append(buffer, entry)
			if len(buffer) >= t.maxRecords {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-t.stopCh:
			for {
				select {
				case entry := <-t.inputCh:
					buffer = append(buffer, entry)
					if len(buffer) >= t.maxRecords {
						flush()
					}
				default:
					flush()
					return
				}
			}
		}
	}
}

func (t *Journaler) Close() error {
	t.stop()
	return nil
}

// TaggerJournaler wraps a Journaler and injects a specific source tag.
type TaggerJournaler struct {
	logger   *Journaler
	metadata JournalerMetadata
}

func NewTaggerJournaler(logger *Journaler, metadata JournalerMetadata) *TaggerJournaler {
	return &TaggerJournaler{
		logger:   logger,
		metadata: metadata,
	}
}

// Write satisfies io.Writer. It delegates to the logger with the pre-defined source.
func (tw *TaggerJournaler) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	tw.logger.Log(string(p), tw.metadata)
	return len(p), nil
}
