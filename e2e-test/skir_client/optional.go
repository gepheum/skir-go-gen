package skir_client

import "fmt"

// Optional[T] holds a value that may or may not be present.
// It is deeply immutable when T is deeply immutable.
type Optional[T any] struct {
	value   T
	present bool
}

// Absent returns an absent Optional. Equivalent to Optional[T]{}.
func Absent[T any]() Optional[T] {
	return Optional[T]{}
}

// OptionalOf returns an Optional containing v.
func OptionalOf[T any](v T) Optional[T] {
	return Optional[T]{value: v, present: true}
}

// OptionalOfNilable converts a possibly-nil pointer to an Optional:
// nil pointer → absent, non-nil pointer → present with *p.
func OptionalOfNilable[T any](p *T) Optional[T] {
	if p == nil {
		return Optional[T]{}
	}
	return Optional[T]{value: *p, present: true}
}

// IsPresent reports whether a value is present.
func (o Optional[T]) IsPresent() bool { return o.present }

// Get returns the value. Panics if not present.
func (o Optional[T]) Get() T {
	if !o.present {
		panic("Optional.Get(): value is not present")
	}
	return o.value
}

// GetOr returns the value if present, or the provided default value otherwise.
func (o Optional[T]) GetOr(def T) T {
	if o.present {
		return o.value
	}
	return def
}

func (o Optional[T]) String() string {
	if !o.present {
		return "null"
	}
	return fmt.Sprint(o.value)
}
