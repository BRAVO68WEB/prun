# prun

A simple CLI tool to run multiple commands in parallel with real-time output streaming.

## Installation

Build from source:

```bash
go build -o prun ./cmd/prun
```

Then move the binary to your PATH:

```bash
sudo mv prun /usr/local/bin/
```

Or use it directly:

```bash
./prun
```

## Quick Start

1. Create a `prun.toml` file in your project:

```toml
tasks = ["app", "redis_server", "server"]

[task.app]
cmd = "pnpm run dev"

[task.redis_server]
cmd = "redis-server"

[task.server]
cmd = "./server"
path = "/home/user/my-server"
```

2. Run all tasks:

```bash
prun
```

## Usage

```bash
prun [flags] [task1 task2 ...]
```

### Flags

- `-c, --config <path>` - Path to config file (default: `prun.toml`)
- `-i, --interactive` - Run in interactive TUI mode
- `-w, --watch` - Watch files and restart all tasks on changes
- `-v, --verbose` - Enable verbose logging
- `-l, --list` - List configured tasks and exit
- `-h, --help` - Show help message

## Interactive Mode

Run `prun` with the `-i` or `--interactive` flag to launch an interactive TUI:

```bash
prun -i
```

### Interactive Mode Features

- **Task List (Left Pane)**: Shows all tasks with status indicators
  - `▲` Running task
  - `✓` Completed successfully
  - `✗` Failed
  - ` ` Idle/pending
- **Log View (Right Pane)**: Shows real-time logs for the selected task
- **Keyboard Controls**:
  - `↑/↓` or `k/j` - Navigate between tasks
  - `PgUp/PgDn` - Scroll logs up/down
  - `Home/End` - Jump to top/bottom of logs
  - `Space` - Page down in logs
  - `q` or `Esc` or `Ctrl-C` - Quit and stop all tasks
  - Task selection shows logs filtered for that specific task

### Interactive Mode Screenshot

The interactive mode provides a clean, organized view similar to tools like Turborepo, making it easy to monitor multiple services during development.

## File Watching

The `--watch` flag enables automatic restarts when files change, perfect for development workflows.

### Global Watch Mode

Restart **all** tasks when any file changes:

```bash
prun -w              # Non-interactive with watch
prun -i -w           # Interactive TUI with watch
```

### Per-Task Watch Mode

Control which tasks restart on file changes in your `prun.toml`:

```toml
tasks = ["app", "build", "server"]

[task.app]
cmd = "npm run dev"
watch = true          # Restart this task on file changes

[task.build]
cmd = "npm run build:watch"
watch = false         # Don't restart (already has its own watch)

[task.server]
cmd = "python -m http.server"
watch = true          # Restart on changes
```

### Watch Behavior

- **Watched directories**: Tasks watch their `path` directory (or current directory if not specified)
- **Debouncing**: Changes are debounced (500ms) to avoid excessive restarts
- **Excluded directories**: `.git`, `node_modules`, `vendor`, `dist`, `build`, and hidden directories are automatically excluded
- **File events**: Watches for `Write` and `Create` events only
- **Intelligent restart**: Only tasks with `watch = true` (or all tasks with `-w` flag) are restarted

### Examples

Run all tasks:
```bash
prun
```

Run in interactive mode:
```bash
prun -i
```

Run specific tasks:
```bash
prun app server
```

Use a different config file:
```bash
prun -c dev.toml
```

List all configured tasks:
```bash
prun --list
```

Enable verbose output:
```bash
prun --verbose
```

Interactive mode with custom config:
```bash
prun -i -c examples/kan-demo.toml
```

Run with file watching (restart on changes):
```bash
prun -w
```

Interactive mode with watch:
```bash
prun -i -w
```

## Configuration

The `prun.toml` file uses TOML format:

### Required Fields

- `tasks` - Array of task names to run (in order)
- `[task.<name>]` - Task definition
  - `cmd` - Command to execute (required)

### Optional Fields

- `path` - Working directory for the command
- `env` - Environment variables (key-value pairs)
- `shell` - Use shell to execute command (default: true)
- `watch` - Restart task when files change (default: false)

### Example Configuration

```toml
tasks = ["frontend", "backend", "database"]

[task.frontend]
cmd = "npm run dev"
path = "./frontend"
env = { NODE_ENV = "development", PORT = "3000" }
watch = true  # Automatically restart on file changes

[task.backend]
cmd = "go run main.go"
path = "./backend"
env = { PORT = "8080" }
watch = true  # Automatically restart on file changes

[task.database]
cmd = "docker-compose up postgres"
watch = false  # Don't restart this task
```

## Features

- ✅ Run multiple commands in parallel
- ✅ Real-time output streaming with task prefixes
- ✅ **Interactive TUI mode** with task list and filtered logs
- ✅ **File watching** - automatically restart tasks on file changes (per-task or global)
- ✅ **Scrollable logs** - PgUp/PgDn, Home/End navigation in TUI
- ✅ **Word wrapping** - long log lines wrap instead of truncate
- ✅ Graceful shutdown on Ctrl-C (SIGINT)
- ✅ Automatic cleanup when any task fails
- ✅ Per-task working directories
- ✅ Per-task environment variables
- ✅ Run specific tasks by name
- ✅ List configured tasks
- ✅ Beautiful terminal UI with colors and status indicators

## How It Works

1. `prun` reads the `prun.toml` configuration file
2. Spawns each task as a separate process
3. Captures stdout/stderr from all tasks
4. Prefixes each line with the task name `[task_name]`
5. Streams output to your terminal in real-time
6. On error or interrupt, cancels all running tasks

## Signal Handling

- **SIGINT (Ctrl-C)**: Forwards signal to all tasks and waits for graceful shutdown
- **SIGTERM**: Forwards signal to all tasks and waits for graceful shutdown
- **Task Failure**: If any task exits with non-zero status, all other tasks are cancelled

## Exit Codes

- `0` - Success (all tasks completed successfully)
- `1` - Task execution failed
- `2` - Config file not found
- `3` - Config file parse error
- `130` - Interrupted by user (SIGINT)

## Development

Run tests:
```bash
make test
```

Build:
```bash
make build
```

Run with example:
```bash
make run
```

## Project Structure

```
.
├── cmd/
│   └── prun/
│       └── main.go          # CLI entrypoint
├── internal/
│   ├── config/
│   │   └── config.go        # TOML parsing and validation
│   └── runner/
│       └── runner.go        # Process management and output streaming
├── examples/                # Example configuration files
│   ├── simple.toml
│   └── dev-servers.toml
├── tests/                   # Test files and configs
│   ├── test.sh              # Integration test suite
│   ├── sample.toml
│   ├── example.toml
│   └── error-test.toml
├── go.mod
├── Makefile                 # Build automation
├── PROJECT_SPEC.md          # Detailed specification
└── README.md
```

## License

MIT

## Contributing

See `PROJECT_SPEC.md` for detailed implementation specifications and acceptance criteria.
