package skir_client

// Optional[T] holds a value that may or may not be present.
// It is deeply immutable when T is deeply immutable.
type Optional[T any] struct {
	value   T
	present bool
}

// OptionalOf returns an Optional containing v.
func OptionalOf[T any](v T) Optional[T] {
	return Optional[T]{value: v, present: true}
}

// OptionalOfPtr converts a pointer to an Optional:
// nil pointer → absent, non-nil pointer → present with *p.
func OptionalOfPtr[T any](p *T) Optional[T] {
	if p == nil {
		return Optional[T]{}
	}
	return Optional[T]{value: *p, present: true}
}

// IsPresent reports whether a value is present.
func (o Optional[T]) IsPresent() bool { return o.present }

// IsEmpty reports whether no value is present.
func (o Optional[T]) IsEmpty() bool { return !o.present }

// Get returns the value. Panics if not present.
func (o Optional[T]) Get() T {
	if !o.present {
		panic("Optional.Get(): value is not present")
	}
	return o.value
}

// GetOrDefault returns the value if present, or the zero value of T otherwise.
func (o Optional[T]) GetOrDefault() T {
	return o.value
}
