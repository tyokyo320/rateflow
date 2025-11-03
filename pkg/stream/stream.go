// Package stream provides utilities for working with Go 1.23+ range over function iterators.
package stream

import "iter"

// Filter creates a new iterator that only yields items matching the predicate.
func Filter[T any](seq iter.Seq[T], predicate func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for item := range seq {
			if predicate(item) {
				if !yield(item) {
					return
				}
			}
		}
	}
}

// Map creates a new iterator by transforming each item.
func Map[T, U any](seq iter.Seq[T], transform func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for item := range seq {
			if !yield(transform(item)) {
				return
			}
		}
	}
}

// Take creates a new iterator that yields at most n items.
func Take[T any](seq iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		count := 0
		for item := range seq {
			if count >= n {
				return
			}
			if !yield(item) {
				return
			}
			count++
		}
	}
}

// Skip creates a new iterator that skips the first n items.
func Skip[T any](seq iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		count := 0
		for item := range seq {
			if count < n {
				count++
				continue
			}
			if !yield(item) {
				return
			}
		}
	}
}

// Collect gathers all items from the iterator into a slice.
func Collect[T any](seq iter.Seq[T]) []T {
	var result []T
	for item := range seq {
		result = append(result, item)
	}
	return result
}

// Count returns the number of items in the iterator.
func Count[T any](seq iter.Seq[T]) int {
	count := 0
	for range seq {
		count++
	}
	return count
}

// First returns the first item from the iterator, or false if empty.
func First[T any](seq iter.Seq[T]) (T, bool) {
	for item := range seq {
		return item, true
	}
	var zero T
	return zero, false
}

// Last returns the last item from the iterator, or false if empty.
func Last[T any](seq iter.Seq[T]) (T, bool) {
	var last T
	found := false
	for item := range seq {
		last = item
		found = true
	}
	return last, found
}

// Any returns true if any item matches the predicate.
func Any[T any](seq iter.Seq[T], predicate func(T) bool) bool {
	for item := range seq {
		if predicate(item) {
			return true
		}
	}
	return false
}

// All returns true if all items match the predicate.
func All[T any](seq iter.Seq[T], predicate func(T) bool) bool {
	for item := range seq {
		if !predicate(item) {
			return false
		}
	}
	return true
}

// Reduce applies an accumulator function over the iterator.
func Reduce[T, U any](seq iter.Seq[T], initial U, accumulator func(U, T) U) U {
	result := initial
	for item := range seq {
		result = accumulator(result, item)
	}
	return result
}

// ForEach executes a function for each item in the iterator.
func ForEach[T any](seq iter.Seq[T], fn func(T)) {
	for item := range seq {
		fn(item)
	}
}

// Chain concatenates two iterators.
func Chain[T any](first, second iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for item := range first {
			if !yield(item) {
				return
			}
		}
		for item := range second {
			if !yield(item) {
				return
			}
		}
	}
}

// Zip combines two iterators into pairs.
func Zip[T, U any](first iter.Seq[T], second iter.Seq[U]) iter.Seq2[T, U] {
	return func(yield func(T, U) bool) {
		next1, stop1 := iter.Pull(first)
		defer stop1()
		next2, stop2 := iter.Pull(second)
		defer stop2()

		for {
			v1, ok1 := next1()
			v2, ok2 := next2()
			if !ok1 || !ok2 {
				return
			}
			if !yield(v1, v2) {
				return
			}
		}
	}
}

// Enumerate adds an index to each item in the iterator.
func Enumerate[T any](seq iter.Seq[T]) iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		index := 0
		for item := range seq {
			if !yield(index, item) {
				return
			}
			index++
		}
	}
}

// Chunk groups items into chunks of the specified size.
func Chunk[T any](seq iter.Seq[T], size int) iter.Seq[[]T] {
	return func(yield func([]T) bool) {
		chunk := make([]T, 0, size)
		for item := range seq {
			chunk = append(chunk, item)
			if len(chunk) == size {
				if !yield(chunk) {
					return
				}
				chunk = make([]T, 0, size)
			}
		}
		if len(chunk) > 0 {
			yield(chunk)
		}
	}
}

// FromSlice creates an iterator from a slice.
func FromSlice[T any](slice []T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, item := range slice {
			if !yield(item) {
				return
			}
		}
	}
}
