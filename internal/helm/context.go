package helm

import "context"

// runResult holds the return values of runWithContext, including a Wait
// function that blocks until the background goroutine completes.
// Callers must call Wait before releasing resources that the goroutine
// may still be using (e.g. temp directories).
type runResult[T any] struct {
	Val T
	Err error
	// Wait blocks until the background goroutine has finished.
	// It is safe to call multiple times.
	Wait func()
}

// runWithContext executes fn in a goroutine and respects context cancellation.
// If the context is cancelled before fn completes, it returns the context error
// immediately. The returned runResult.Wait function blocks until the goroutine
// finishes, which callers must use before cleaning up shared resources.
func runWithContext[T any](ctx context.Context, fn func() (T, error)) runResult[T] {
	type result struct {
		val T
		err error
	}

	done := make(chan struct{})
	resultCh := make(chan result, 1)
	go func() {
		defer close(done)
		val, err := fn()
		resultCh <- result{val, err}
	}()

	wait := func() { <-done }

	select {
	case <-ctx.Done():
		var zero T
		return runResult[T]{Val: zero, Err: ctx.Err(), Wait: wait}
	case r := <-resultCh:
		return runResult[T]{Val: r.val, Err: r.err, Wait: wait}
	}
}
