package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"prun/internal/config"

	"github.com/fsnotify/fsnotify"
)

// Watcher manages file watching and task restarts
type Watcher struct {
	cfg          *config.Config
	tasks        []string
	verbose      bool
	globalWatch  bool
	eventChan    chan LogEvent
	fsWatcher    *fsnotify.Watcher
	restartChans map[string]chan struct{}
	mu           sync.Mutex
}

// NewWatcher creates a new file watcher
func NewWatcher(cfg *config.Config, tasks []string, verbose bool, globalWatch bool) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	return &Watcher{
		cfg:          cfg,
		tasks:        tasks,
		verbose:      verbose,
		globalWatch:  globalWatch,
		fsWatcher:    fsWatcher,
		restartChans: make(map[string]chan struct{}),
	}, nil
}

// SetEventChannel sets a channel for publishing log events
func (w *Watcher) SetEventChannel(ch chan LogEvent) {
	w.eventChan = ch
}

// Start begins watching files and running tasks
func (w *Watcher) Start(ctx context.Context) error {
	// Setup watchers for each task
	for _, taskName := range w.tasks {
		taskDef := w.cfg.TaskDefs[taskName]
		shouldWatch := w.globalWatch || taskDef.Watch

		if shouldWatch {
			watchDir := taskDef.Path
			if watchDir == "" {
				watchDir = "."
			}

			// Add the directory to watch
			if err := w.addWatchRecursive(watchDir); err != nil {
				return fmt.Errorf("failed to watch directory for task '%s': %w", taskName, err)
			}

			if w.verbose {
				w.logEvent(taskName, fmt.Sprintf("Watching directory: %s", watchDir))
			}
		}
	}

	// Start file watcher event loop
	go w.watchLoop(ctx)

	// Start all tasks
	var wg sync.WaitGroup
	for _, taskName := range w.tasks {
		w.restartChans[taskName] = make(chan struct{}, 1)
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			w.runTaskWithRestart(ctx, name)
		}(taskName)
	}

	wg.Wait()
	return nil
}

// addWatchRecursive adds a directory and all its subdirectories to the watcher
func (w *Watcher) addWatchRecursive(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and node_modules, .git, etc.
		if info.IsDir() {
			base := filepath.Base(path)
			if base[0] == '.' || base == "node_modules" || base == "vendor" || base == "dist" || base == "build" {
				return filepath.SkipDir
			}
			return w.fsWatcher.Add(path)
		}
		return nil
	})
}

// watchLoop monitors file system events
func (w *Watcher) watchLoop(ctx context.Context) {
	// Debounce timer to avoid too many restarts
	var debounceTimer *time.Timer
	debounceDuration := 500 * time.Millisecond

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return
			}

			// Only watch Write and Create events
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				if w.verbose {
					w.logEvent("watcher", fmt.Sprintf("File changed: %s", event.Name))
				}

				// Reset debounce timer
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(debounceDuration, func() {
					w.triggerRestarts()
				})
			}
		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return
			}
			if w.verbose {
				w.logEvent("watcher", fmt.Sprintf("Error: %v", err))
			}
		}
	}
}

// triggerRestarts signals all watched tasks to restart
func (w *Watcher) triggerRestarts() {
	w.mu.Lock()
	defer w.mu.Unlock()

	for taskName, restartChan := range w.restartChans {
		taskDef := w.cfg.TaskDefs[taskName]
		if w.globalWatch || taskDef.Watch {
			select {
			case restartChan <- struct{}{}:
				if w.verbose {
					w.logEvent(taskName, "Restarting due to file change...")
				}
			default:
				// Channel already has a pending restart
			}
		}
	}
}

// runTaskWithRestart runs a task and restarts it when signaled
func (w *Watcher) runTaskWithRestart(ctx context.Context, taskName string) {
	taskDef := w.cfg.TaskDefs[taskName]
	shouldWatch := w.globalWatch || taskDef.Watch
	restartChan := w.restartChans[taskName]

	for {
		// Create a cancellable context for this task instance
		taskCtx, cancel := context.WithCancel(ctx)

		// Run the task in a goroutine
		done := make(chan error, 1)
		go func() {
			r := New(w.cfg, []string{taskName}, w.verbose)
			if w.eventChan != nil {
				r.SetEventChannel(w.eventChan)
			}
			done <- r.runTask(taskCtx, taskName)
		}()

		// Wait for completion, restart signal, or context cancellation
		select {
		case <-ctx.Done():
			cancel()
			return
		case <-restartChan:
			if shouldWatch {
				// Cancel current task and restart
				cancel()
				<-done // Wait for task to finish
				w.logEvent(taskName, "Restarted")
				continue
			}
		case err := <-done:
			cancel()
			if err != nil && w.verbose {
				w.logEvent(taskName, fmt.Sprintf("Exited with error: %v", err))
			}

			// If not watching, exit after first run
			if !shouldWatch {
				return
			}

			// Otherwise wait for restart signal
			select {
			case <-ctx.Done():
				return
			case <-restartChan:
				w.logEvent(taskName, "Restarted")
				continue
			}
		}
	}
}

// logEvent sends a log event
func (w *Watcher) logEvent(taskName, message string) {
	if w.eventChan != nil {
		w.eventChan <- LogEvent{
			Task:  taskName,
			Line:  message,
			IsErr: false,
			Time:  time.Now(),
		}
	} else {
		fmt.Printf("[%s] %s\n", taskName, message)
	}
}

// Close closes the watcher
func (w *Watcher) Close() error {
	return w.fsWatcher.Close()
}
