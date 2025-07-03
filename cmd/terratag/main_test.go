package main

import (
	"context"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/cloudyali/terratag/cli"
	"github.com/stretchr/testify/assert"
)

func TestHandleShutdownSignals(t *testing.T) {
	// Set up global context
	globalCtx, globalCancel = context.WithCancel(context.Background())
	defer globalCancel()

	sigChan := make(chan os.Signal, 1)
	
	// Start signal handler in background
	go handleShutdownSignals(sigChan)
	
	// Send a signal
	sigChan <- syscall.SIGINT
	
	// Wait for context to be cancelled
	select {
	case <-globalCtx.Done():
		// Expected behavior
	case <-time.After(1 * time.Second):
		t.Error("Context was not cancelled within timeout")
	}
}

func TestGracefulShutdown_WithTimeout(t *testing.T) {
	// Reset shutdown wait group
	shutdownWg = sync.WaitGroup{}
	
	// Add a long-running operation
	shutdownWg.Add(1)
	go func() {
		defer shutdownWg.Done()
		time.Sleep(100 * time.Millisecond) // Short sleep for test
	}()
	
	start := time.Now()
	gracefulShutdown()
	duration := time.Since(start)
	
	// Should complete quickly since the operation finishes
	assert.Less(t, duration, 1*time.Second)
}

func TestGracefulShutdown_ForcedTimeout(t *testing.T) {
	// This test is more complex and would require mocking
	// For now, we'll test the basic functionality
	t.Skip("Timeout behavior testing requires more complex setup")
}

func TestRunWithContext_Cancellation(t *testing.T) {
	// Create a context that gets cancelled
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create minimal args that won't actually run terratag
	args := cli.Args{
		ValidateOnly: true,
		StandardFile: "/nonexistent/file.yaml", // This will cause an error
		Dir:          ".",
		Type:         "terraform",
	}
	
	// Cancel immediately
	cancel()
	
	err := runWithContext(ctx, args)
	
	// Should return context cancellation error
	assert.ErrorIs(t, err, context.Canceled)
}

func TestRunWithContext_Success(t *testing.T) {
	// Create a context
	ctx := context.Background()
	
	// Create args for validation mode with non-existent file (will fail quickly)
	args := cli.Args{
		ValidateOnly: true,
		StandardFile: "/nonexistent/file.yaml",
		Dir:          ".",
		Type:         "terraform",
	}
	
	err := runWithContext(ctx, args)
	
	// Should return the actual terratag error, not context cancellation
	assert.Error(t, err)
	assert.NotErrorIs(t, err, context.Canceled)
}