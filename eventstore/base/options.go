package base

func DefaultBufferSize(defaultBufferSize int) Option {
	return func(s *EventStream) {
		s.defaultBufferSize = defaultBufferSize
	}
}
