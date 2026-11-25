package messaging

type Subscription[T any] struct {
	Source Source
	Raw    T
}

type Source struct {
	Stream  string
	Subject string
	Options map[string]any
}

type Destination struct {
	Subject string
	Options map[string]any
}
