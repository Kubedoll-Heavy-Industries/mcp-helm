package helm

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunWithContext(t *testing.T) {
	t.Run("returns result when context not cancelled", func(t *testing.T) {
		ctx := context.Background()
		res := runWithContext(ctx, func() (string, error) {
			return "success", nil
		})
		if res.Err != nil {
			t.Errorf("unexpected error: %v", res.Err)
		}
		if res.Val != "success" {
			t.Errorf("got %q, want %q", res.Val, "success")
		}
	})

	t.Run("returns error from function", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := errors.New("function error")
		res := runWithContext(ctx, func() (string, error) {
			return "", expectedErr
		})
		if !errors.Is(res.Err, expectedErr) {
			t.Errorf("got error %v, want %v", res.Err, expectedErr)
		}
		if res.Val != "" {
			t.Errorf("got result %q, want empty", res.Val)
		}
	})

	t.Run("returns context error when cancelled before completion", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// Cancel immediately
		cancel()

		// Function blocks on a channel that is never closed; context cancellation unblocks runWithContext.
		block := make(chan struct{})
		defer close(block) // unblock goroutine so it can exit
		res := runWithContext(ctx, func() (string, error) {
			<-block
			return "should not see this", nil
		})

		if !errors.Is(res.Err, context.Canceled) {
			t.Errorf("got error %v, want context.Canceled", res.Err)
		}
		if res.Val != "" {
			t.Errorf("got result %q, want empty", res.Val)
		}
	})

	t.Run("returns context deadline exceeded on timeout", func(t *testing.T) {
		// Use a deadline in the past so the context is already expired.
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
		defer cancel()

		// Function blocks on a channel; the already-expired context unblocks runWithContext deterministically.
		block := make(chan struct{})
		defer close(block) // unblock goroutine so it can exit
		res := runWithContext(ctx, func() (string, error) {
			<-block
			return "should not see this", nil
		})

		if !errors.Is(res.Err, context.DeadlineExceeded) {
			t.Errorf("got error %v, want context.DeadlineExceeded", res.Err)
		}
		if res.Val != "" {
			t.Errorf("got result %q, want empty", res.Val)
		}
	})

	t.Run("works with different types", func(t *testing.T) {
		ctx := context.Background()

		// Test with int
		intRes := runWithContext(ctx, func() (int, error) {
			return 42, nil
		})
		if intRes.Err != nil || intRes.Val != 42 {
			t.Errorf("int test failed: got %d, err %v", intRes.Val, intRes.Err)
		}

		// Test with struct
		type testStruct struct {
			Value string
		}
		structRes := runWithContext(ctx, func() (testStruct, error) {
			return testStruct{Value: "test"}, nil
		})
		if structRes.Err != nil || structRes.Val.Value != "test" {
			t.Errorf("struct test failed: got %+v, err %v", structRes.Val, structRes.Err)
		}
	})

	t.Run("Wait blocks until goroutine finishes", func(t *testing.T) {
		// Cancel the context upfront so runWithContext returns immediately
		// with context.Canceled, while the goroutine is still running.
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		var finished atomic.Bool
		proceed := make(chan struct{})

		res := runWithContext(ctx, func() (string, error) {
			<-proceed // block until test says go
			finished.Store(true)
			return "done", nil
		})

		if !errors.Is(res.Err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", res.Err)
		}

		// The goroutine should still be blocked
		if finished.Load() {
			t.Fatal("goroutine finished before Wait was called")
		}

		// Let the goroutine proceed
		close(proceed)

		// Wait for goroutine to complete
		res.Wait()

		if !finished.Load() {
			t.Fatal("goroutine did not finish after Wait")
		}
	})

	t.Run("Wait is safe to call multiple times", func(t *testing.T) {
		ctx := context.Background()
		res := runWithContext(ctx, func() (string, error) {
			return "ok", nil
		})

		// Calling Wait multiple times should not block or panic
		res.Wait()
		res.Wait()
		res.Wait()
	})

	t.Run("cleanup after cancelled context does not race with goroutine", func(t *testing.T) {
		// This test verifies the core fix: waiting for the goroutine
		// before cleaning up a shared resource (temp directory).
		// Run with -race to detect data races.
		tmpDir := t.TempDir()

		ctx, cancel := context.WithCancel(context.Background())

		// Gate ensures the goroutine is still running when cancel() fires,
		// creating a genuine race between cleanup and goroutine file writes.
		gate := make(chan struct{})

		// Cancel context from a separate goroutine after a brief moment,
		// so runWithContext's select unblocks via ctx.Done().
		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		res := runWithContext(ctx, func() (string, error) {
			<-gate // block until test releases us after cancel
			// Simulate writing to a temp directory (like DownloadTo)
			f, err := os.Create(filepath.Join(tmpDir, "chart.tgz"))
			if err != nil {
				return "", err
			}
			defer func() { _ = f.Close() }()
			_, err = f.WriteString("fake chart data")
			return f.Name(), err
		})

		// runWithContext returned due to cancellation; now release the goroutine
		close(gate)

		// Wait for goroutine before cleanup — this is the fix
		res.Wait()

		// Now safe to remove the directory
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatalf("RemoveAll failed: %v", err)
		}
	})
}
