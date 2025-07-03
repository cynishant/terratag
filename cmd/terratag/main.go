package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/cloudyali/terratag"
	"github.com/cloudyali/terratag/cli"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/logutils"
)

var (
	version = "dev"
	// Global context for graceful shutdown
	globalCtx    context.Context
	globalCancel context.CancelFunc
	shutdownWg   sync.WaitGroup
)

func main() {
	args, err := cli.InitArgs()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Usage: terratag -tags='{ \"some_tag\": \"value\" }' [-dir=\".\"]")

		return
	}

	if args.Version {
		var versionPrefix string

		if !strings.HasPrefix(version, "v") {
			versionPrefix = "v"
		}

		fmt.Printf("Terratag %s%s\n", versionPrefix, version)

		return
	}

	initLogFiltering(args.Verbose)

	// Set up graceful shutdown handling
	globalCtx, globalCancel = context.WithCancel(context.Background())
	defer globalCancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Start signal handler in a goroutine
	go handleShutdownSignals(sigChan)

	// In API server mode, just keep the process running
	if args.APIServerMode {
		log.Println("[INFO] Running in API server mode - process will stay active")
		log.Println("[INFO] Send SIGINT, SIGTERM, or SIGQUIT to gracefully shutdown")
		
		// TODO: Start actual API server here when implemented
		// For now, just wait for shutdown signal
		<-globalCtx.Done()
		
		log.Println("[INFO] Received shutdown signal, stopping API server mode")
		gracefulShutdown()
		return
	}

	// For regular operations, run with context
	if err := runWithContext(globalCtx, args); err != nil {
		log.Printf("[ERROR] execution failed due to an error\n%v", err)
		os.Exit(1)
	}

	// Wait for any remaining operations to complete
	gracefulShutdown()
}

// handleShutdownSignals handles graceful shutdown signals
func handleShutdownSignals(sigChan chan os.Signal) {
	sig := <-sigChan
	log.Printf("[INFO] Received signal %v, initiating graceful shutdown...", sig)
	
	// Cancel the global context to signal all operations to stop
	globalCancel()
}

// gracefulShutdown waits for all operations to complete with a timeout
func gracefulShutdown() {
	log.Println("[INFO] Waiting for operations to complete...")
	
	// Create a timeout context for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	
	// Wait for all operations with timeout
	done := make(chan struct{})
	go func() {
		shutdownWg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		log.Println("[INFO] All operations completed successfully")
	case <-shutdownCtx.Done():
		log.Println("[WARN] Shutdown timeout reached, forcing exit")
	}
}

// runWithContext runs terratag with context support for cancellation
func runWithContext(ctx context.Context, args cli.Args) error {
	// Add this operation to the wait group
	shutdownWg.Add(1)
	defer shutdownWg.Done()
	
	// Create a channel to receive the result
	resultChan := make(chan error, 1)
	
	// Run terratag in a goroutine
	go func() {
		resultChan <- terratag.Terratag(args)
	}()
	
	// Wait for either completion or cancellation
	select {
	case err := <-resultChan:
		return err
	case <-ctx.Done():
		log.Println("[INFO] Operation cancelled by shutdown signal")
		return ctx.Err()
	}
}

func initLogFiltering(verbose bool) {
	level := "INFO"
	if verbose {
		level = "DEBUG"
	}

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "TRACE", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(level),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)
	hclog.DefaultOutput = filter
}
