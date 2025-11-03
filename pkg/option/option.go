// Package option provides a generic Option type for handling optional values,
// inspired by Rust's Option type.
package option

// Option represents a value that may or may not be present.
type Option[T any] struct {
	value   T
	present bool
}

// Some creates an Option containing a value.
func Some[T any](value T) Option[T] {
	return Option[T]{value: value, present: true}
}

// None creates an empty Option.
func None[T any]() Option[T] {
	var zero T
	return Option[T]{value: zero, present: false}
}

// IsSome returns true if the Option contains a value.
func (o Option[T]) IsSome() bool {
	return o.present
}

// IsNone returns true if the Option is empty.
func (o Option[T]) IsNone() bool {
	return !o.present
}

// Unwrap returns the contained value.
// Panics if the Option is None.
func (o Option[T]) Unwrap() T {
	if !o.present {
		panic("called Unwrap on None")
	}
	return o.value
}

// UnwrapOr returns the contained value or a default value if None.
func (o Option[T]) UnwrapOr(defaultValue T) T {
	if !o.present {
		return defaultValue
	}
	return o.value
}

// UnwrapOrElse returns the contained value or computes a default using the provided function.
func (o Option[T]) UnwrapOrElse(fn func() T) T {
	if !o.present {
		return fn()
	}
	return o.value
}

// Map transforms the contained value using the provided function.
// Returns None if the Option is None.
func (o Option[T]) Map(fn func(T) T) Option[T] {
	if !o.present {
		return None[T]()
	}
	return Some(fn(o.value))
}

// MapOr transforms the contained value or returns a default.
func (o Option[T]) MapOr(defaultValue T, fn func(T) T) T {
	if !o.present {
		return defaultValue
	}
	return fn(o.value)
}

// MapOrElse transforms the contained value or computes a default.
func (o Option[T]) MapOrElse(defaultFn func() T, fn func(T) T) T {
	if !o.present {
		return defaultFn()
	}
	return fn(o.value)
}

// FlatMap transforms the contained value into another Option.
func FlatMap[T, U any](o Option[T], fn func(T) Option[U]) Option[U] {
	if !o.present {
		return None[U]()
	}
	return fn(o.value)
}

// Filter returns None if the Option is None, or if the predicate returns false.
func (o Option[T]) Filter(predicate func(T) bool) Option[T] {
	if !o.present {
		return None[T]()
	}
	if predicate(o.value) {
		return o
	}
	return None[T]()
}

// Inspect calls the provided function with the value if Some, for side effects.
func (o Option[T]) Inspect(fn func(T)) Option[T] {
	if o.present {
		fn(o.value)
	}
	return o
}

// OkOr converts the Option into a Result.
func (o Option[T]) OkOr(err error) (T, error) {
	if !o.present {
		var zero T
		return zero, err
	}
	return o.value, nil
}

// FromPtr creates an Option from a pointer.
// Returns None if the pointer is nil.
func FromPtr[T any](ptr *T) Option[T] {
	if ptr == nil {
		return None[T]()
	}
	return Some(*ptr)
}

// ToPtr converts the Option to a pointer.
// Returns nil if the Option is None.
func (o Option[T]) ToPtr() *T {
	if !o.present {
		return nil
	}
	return &o.value
}
