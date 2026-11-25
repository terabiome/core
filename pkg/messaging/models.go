package messaging

type Subscription struct {
	source Source
	stop   func()
	drain  func()
	done   <-chan struct{}
}

func NewSubscription(source Source, stop, drain func(), done <-chan struct{}) *Subscription {
	return &Subscription{
		source: source,
		stop:   stop,
		drain:  drain,
		done:   done,
	}
}

func (s *Subscription) Source() Source {
	return s.source
}

func (s *Subscription) Stop() {
	if s != nil && s.stop != nil {
		s.stop()
	}
}

func (s *Subscription) Drain() {
	if s != nil && s.drain != nil {
		s.drain()
	}
}

func (s *Subscription) Done() <-chan struct{} {
	if s == nil {
		return nil
	}
	return s.done
}

type Source struct {
	Stream  string
	Subject string
	Options any // carries transport-specific subscription settings.
}

type Destination struct {
	Subject string
	Options any // carries transport-specific publish settings.
}
