package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"prun/internal/config"
	"prun/internal/runner"
)

const (
	exitCodeConfigNotFound = 2
	exitCodeParseFailed    = 3
	exitCodeRunFailed      = 1
)

func main() {
	// Parse CLI flags
	configPath := flag.String("c", "prun.toml", "path to config file")
	flag.StringVar(configPath, "config", "prun.toml", "path to config file")

	verbose := flag.Bool("v", false, "enable verbose logging")
	flag.BoolVar(verbose, "verbose", false, "enable verbose logging")

	list := flag.Bool("l", false, "list tasks and exit")
	flag.BoolVar(list, "list", false, "list tasks and exit")

	showHelp := flag.Bool("h", false, "show help")
	flag.BoolVar(showHelp, "help", false, "show help")

	flag.Parse()

	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	// Check if config file exists
	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "prun: no %s found — run `prun --help` to see usage\n", *configPath)
		os.Exit(exitCodeConfigNotFound)
	}

	// Load and parse config
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "prun: failed to parse config: %v\n", err)
		os.Exit(exitCodeParseFailed)
	}

	// List tasks if requested
	if *list {
		fmt.Println("Configured tasks:")
		for _, taskName := range cfg.Tasks {
			taskDef := cfg.TaskDefs[taskName]
			fmt.Printf("  %s: %s\n", taskName, taskDef.Cmd)
		}
		os.Exit(0)
	}

	// Get tasks to run
	tasksToRun, err := cfg.GetTasksToRun(flag.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, "prun: %v\n", err)
		os.Exit(exitCodeRunFailed)
	}

	if len(tasksToRun) == 0 {
		fmt.Fprintln(os.Stderr, "prun: no tasks to run")
		os.Exit(0)
	}

	// Create runner
	r := runner.New(cfg, tasksToRun, *verbose)

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Run tasks in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- r.Run(ctx)
	}()

	// Wait for completion or signal
	select {
	case <-sigChan:
		if *verbose {
			fmt.Fprintln(os.Stderr, "\nprun: received interrupt signal, shutting down...")
		}
		cancel()
		// Wait a bit for graceful shutdown
		err := <-errChan
		if err != nil && *verbose {
			fmt.Fprintf(os.Stderr, "prun: %v\n", err)
		}
		os.Exit(130) // Standard exit code for SIGINT
	case err := <-errChan:
		if err != nil {
			fmt.Fprintf(os.Stderr, "prun: %v\n", err)
			os.Exit(exitCodeRunFailed)
		}
	}
}

func printHelp() {
	fmt.Println(`prun - run multiple commands in parallel

Usage:
  prun [flags] [task1 task2 ...]

Flags:
  -c, --config <path>   Path to config file (default: prun.toml)
  -v, --verbose         Enable verbose logging
  -l, --list            List configured tasks and exit
  -h, --help            Show this help message

Examples:
  prun                  Run all tasks defined in prun.toml
  prun app server       Run only 'app' and 'server' tasks
  prun -c dev.toml      Use dev.toml instead of prun.toml
  prun --list           List all configured tasks

Config format (prun.toml):
  [tasks]
  - app
  - server

  [task.app]
  cmd = "npm run dev"

  [task.server]
  cmd = "./server"
  path = "/path/to/server"
  
For more information, see PROJECT_SPEC.md`)
}
