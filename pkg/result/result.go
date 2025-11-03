// Package result provides a generic Result type for elegant error handling,
// inspired by Rust's Result type.
package result

// Result represents a value that can be either a success or an error.
type Result[T any] struct {
	value T
	err   error
}

// Ok creates a successful Result containing the given value.
func Ok[T any](value T) Result[T] {
	return Result[T]{value: value}
}

// Err creates a failed Result containing the given error.
func Err[T any](err error) Result[T] {
	var zero T
	return Result[T]{value: zero, err: err}
}

// Unwrap returns the value and error from the Result.
func (r Result[T]) Unwrap() (T, error) {
	return r.value, r.err
}

// IsOk returns true if the Result is successful.
func (r Result[T]) IsOk() bool {
	return r.err == nil
}

// IsErr returns true if the Result contains an error.
func (r Result[T]) IsErr() bool {
	return r.err != nil
}

// UnwrapOr returns the contained value or a default value if there's an error.
func (r Result[T]) UnwrapOr(defaultValue T) T {
	if r.IsErr() {
		return defaultValue
	}
	return r.value
}

// UnwrapOrElse returns the contained value or computes a default using the provided function.
func (r Result[T]) UnwrapOrElse(fn func(error) T) T {
	if r.IsErr() {
		return fn(r.err)
	}
	return r.value
}

// Map transforms the success value using the provided function.
// If the Result is an error, it returns an error Result of the new type.
func Map[T, U any](r Result[T], fn func(T) U) Result[U] {
	if r.IsErr() {
		return Err[U](r.err)
	}
	return Ok(fn(r.value))
}

// FlatMap chains operations that return Results.
func FlatMap[T, U any](r Result[T], fn func(T) Result[U]) Result[U] {
	if r.IsErr() {
		return Err[U](r.err)
	}
	return fn(r.value)
}

// AndThen is an alias for FlatMap for better readability in chains.
func (r Result[T]) AndThen(fn func(T) Result[T]) Result[T] {
	return FlatMap(r, fn)
}

// OrElse returns the Result if it's Ok, otherwise calls fn with the error.
func (r Result[T]) OrElse(fn func(error) Result[T]) Result[T] {
	if r.IsOk() {
		return r
	}
	return fn(r.err)
}

// Inspect calls the provided function with the value if Ok, for side effects.
func (r Result[T]) Inspect(fn func(T)) Result[T] {
	if r.IsOk() {
		fn(r.value)
	}
	return r
}

// InspectErr calls the provided function with the error if Err, for side effects.
func (r Result[T]) InspectErr(fn func(error)) Result[T] {
	if r.IsErr() {
		fn(r.err)
	}
	return r
}
