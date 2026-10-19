package messaging

type Subscription[T any] struct {
	Source Source
	Raw    T
}

type Source struct {
	Topic      string
	Parameters map[string]any
}

type Destination struct {
	Topic     string
	Parameter map[string]any
}
