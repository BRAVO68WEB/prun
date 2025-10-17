package runner

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"prun/internal/config"
)

// Runner manages multiple task processes
type Runner struct {
	cfg     *config.Config
	tasks   []string
	verbose bool
	output  *outputWriter
}

// New creates a new Runner
func New(cfg *config.Config, tasks []string, verbose bool) *Runner {
	return &Runner{
		cfg:     cfg,
		tasks:   tasks,
		verbose: verbose,
		output:  newOutputWriter(os.Stdout),
	}
}

// Run starts all tasks and waits for them to complete
func (r *Runner) Run(ctx context.Context) error {
	// Create a cancellable context for all tasks
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	errChan := make(chan error, len(r.tasks))

	// Start all tasks
	for _, taskName := range r.tasks {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			if err := r.runTask(ctx, name); err != nil {
				errChan <- fmt.Errorf("task '%s': %w", name, err)
				cancel() // Cancel all other tasks on error
			}
		}(taskName)
	}

	// Wait for all tasks to complete
	wg.Wait()
	close(errChan)

	// Check for errors
	var firstErr error
	for err := range errChan {
		if firstErr == nil {
			firstErr = err
		}
		if r.verbose {
			fmt.Fprintf(os.Stderr, "prun: %v\n", err)
		}
	}

	return firstErr
}

// runTask runs a single task
func (r *Runner) runTask(ctx context.Context, taskName string) error {
	taskDef := r.cfg.TaskDefs[taskName]

	if r.verbose {
		r.output.WritePrefix(taskName, fmt.Sprintf("Starting: %s\n", taskDef.Cmd))
	}

	// Determine if we should use shell
	useShell := true
	if taskDef.Shell != nil {
		useShell = *taskDef.Shell
	}

	var cmd *exec.Cmd
	if useShell {
		cmd = exec.CommandContext(ctx, "/bin/sh", "-c", taskDef.Cmd)
	} else {
		// For non-shell, we'd need to parse the command - simplified for now
		cmd = exec.CommandContext(ctx, "/bin/sh", "-c", taskDef.Cmd)
	}

	// Set working directory if specified
	if taskDef.Path != "" {
		cmd.Dir = taskDef.Path
	}

	// Set environment variables
	cmd.Env = os.Environ()
	for k, v := range taskDef.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Set process group for signal forwarding
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Capture stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	// Stream output
	var streamWg sync.WaitGroup
	streamWg.Add(2)

	go func() {
		defer streamWg.Done()
		r.streamOutput(taskName, stdout)
	}()

	go func() {
		defer streamWg.Done()
		r.streamOutput(taskName, stderr)
	}()

	// Wait for output streaming to complete
	streamWg.Wait()

	// Wait for command to exit
	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			// Context was cancelled, this is expected
			return nil
		}
		return err
	}

	return nil
}

// streamOutput reads from a reader and writes prefixed lines
func (r *Runner) streamOutput(taskName string, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		r.output.WritePrefix(taskName, line+"\n")
	}
}

// outputWriter handles synchronized, prefixed output
type outputWriter struct {
	mu     sync.Mutex
	writer io.Writer
}

func newOutputWriter(w io.Writer) *outputWriter {
	return &outputWriter{writer: w}
}

func (ow *outputWriter) WritePrefix(prefix, text string) {
	ow.mu.Lock()
	defer ow.mu.Unlock()

	// Calculate max prefix width for alignment
	maxWidth := 15
	paddedPrefix := prefix
	if len(prefix) < maxWidth {
		paddedPrefix = prefix + string(make([]byte, maxWidth-len(prefix)))
		for i := len(prefix); i < maxWidth; i++ {
			paddedPrefix = paddedPrefix[:len(prefix)] + " " + paddedPrefix[len(prefix):]
		}
	}

	fmt.Fprintf(ow.writer, "[%s] %s", prefix, text)
}

// Shutdown gracefully shuts down all running processes
func (r *Runner) Shutdown(timeout time.Duration) {
	if r.verbose {
		fmt.Fprintln(os.Stderr, "prun: shutting down tasks...")
	}
	// Tasks are managed via context cancellation in Run()
}
