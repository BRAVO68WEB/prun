# prun — Project Specification

Short description
-----------------
`prun` is a small command-line tool that reads a `prun.toml` configuration file, starts multiple tasks (commands) in parallel, and streams their combined output to the console in real time. It's designed for local development workflows where several processes (app, server, databases, watchers) need to be run together.

Goals
-----
- Simple, predictable CLI for running multiple processes in parallel.
- Clear, easy-to-edit configuration file (`prun.toml`).
- Real-time, prefixed output per-task so users can distinguish logs.
- Signal forwarding and graceful shutdown.
- Minimal dependencies and easy packaging for Linux/macOS/Windows.

Config format
-------------
`prun` will look for `prun.toml` in the current working directory (or a location provided by a `--config` flag). The file is a TOML document with a `tasks` list and per-task tables.

Example:

```
[tasks]
- app
- redis_server
- server

[task.app]
cmd = "pnpm run dev"

[task.redis_server]
cmd = "redis-server"

[task.server]
cmd = "./server"
path = "/home/user/my-server"
```

Semantics:
- `tasks` is an ordered list of task names. Order is primarily for user readability; the runner will start tasks in the order listed but run them concurrently.
- Each `[task.<name>]` table must include `cmd` (string). Optionally it may include:
  - `path` (string) — working directory to run the command in.
  - `env` (table) — map of environment variables (string -> string) to set for the task.
  - `restart` (boolean|string) — whether/how to restart a task on exit: `false` (default), `true` (restart always), or a policy like `on-failure`.

CLI behavior
------------
Invocation: `prun [flags] [--] [task1 task2 ...]`

Flags:
- `-c, --config <path>`: path to `prun.toml`. Defaults to `./prun.toml`.
- `-v, --verbose` : enable verbose logs for prun internals.
- `-l, --list` : print tasks defined and exit.
- `-h, --help` : show usage.

Behavior:
- On start, `prun` searches for the config file. If not found, it prints a short message: "prun: no prun.toml found — run `prun help` to create one" and exits with status code 2.
- If found, `prun` parses the TOML. If parsing fails, it prints the parse error and exits with status code 3.
- If the user passes specific task names as arguments, only those tasks (in the order provided) are started. If no tasks are passed, all tasks listed under `[tasks]` are started, in that order.

Process lifecycle and output
---------------------------
- `prun` starts all selected tasks as child processes.
- For each task, `prun` captures stdout and stderr, prefixes each line with a short task label (e.g., `[app] `, `[redis_server] `) and writes to the main stdout/stderr stream. The prefix helps distinguish interleaved logs.
- The prefixing should support configurable width or colorization when stdout is a TTY. When output is not a TTY (e.g., piped to a file), `prun` should omit colors and may shorten prefixes.
- Output should be unbuffered and near-real-time. Use a line-based scanner or incremental flush to avoid large delays.

Exit codes and failure policies
------------------------------
- If any task exits with a non-zero status, `prun` should by default terminate all other tasks and exit with a non-zero status reflecting failure.
- The `restart` policy per-task (if implemented) can override this behavior — e.g., tasks with `restart = true` will be restarted automatically and won't cause `prun` to exit unless explicitly configured.
- On receiving termination signals (SIGINT, SIGTERM), `prun` should forward the signal to child processes and wait for them to exit gracefully (with a short timeout, e.g., 5s) before forcing termination.

Edge cases and considerations
-----------------------------
- Commands producing binary or very long lines — prefixing should be done safely (don't assume UTF-8 or small sizes).
- Process groups and shells: commands in `cmd` may be shell forms. By default, `prun` should run the command through a shell (like `/bin/sh -c`) to support complex commands, but provide a `shell = false` option to run exec directly if desired.
- Environment inheritance: tasks should inherit the parent environment unless overridden in `env`.
- Port collisions and graceful restarts are out-of-scope for initial implementation.

Acceptance criteria
-------------------
Minimum viable product (MVP):
- `prun` searches for `./prun.toml`. If absent, prints: "no prun.toml found — run `prun help`" and exits code 2.
- Parses a valid `prun.toml` with at least `tasks` and `cmd` entries.
- Starts all defined tasks concurrently, streams output with task prefixes, and shows both stdout and stderr combined in real time.
- On a non-zero exit from any task, `prun` terminates remaining tasks and exits non-zero.
- Handles SIGINT (Ctrl-C) by forwarding and cleanly shutting down processes.

Optional / Nice-to-have (post-MVP):
- Colorized prefixes and adjustable width.
- Per-task restart policies.
- `--watch` to restart tasks when files change.
- `--parallelism` to limit number of concurrently running tasks.
- Built-in command to generate a sample `prun.toml` (`prun init`).
- Windows-specific behavior and proper signal handling on Windows.

Testing
-------
Unit tests:
- TOML parsing tests for valid and invalid configs.
- Task selection logic (all tasks vs. subset by args).
- Prefix formatting and TTY detection.

Integration tests (fast):
- Start two small processes (e.g., `sh -c 'echo a; sleep 0.1; echo done'`) and assert `prun` output contains both prefixes.
- Verify `prun` exits non-zero when a task fails.
- Signal forwarding test: send SIGINT to `prun` and assert child processes receive term signal.

Security and safety
-------------------
- Avoid shell injection risks by documenting that `cmd` is executed via a shell; if users supply untrusted config files, commands will run with the user's privileges.
- The tool will not attempt to sandbox commands.

Implementation plan (high level)
--------------------------------
1. Parse CLI flags and find config path.
2. Read and parse `prun.toml` into a typed config object.
3. Resolve the list of tasks to run.
4. For each task, spawn a child process with appropriate working dir and env.
5. Start goroutines to read stdout/stderr line-by-line, prefix, and write to the main stdout/stderr with synchronization.
6. Monitor child exits; on failure or signal, implement shutdown procedure.
7. Add tests and sample config files.

Files to add
------------
- `cmd/prun/main.go` — CLI entrypoint and wiring.
- `internal/config/config.go` — TOML parsing (types + parser).
- `internal/runner/runner.go` — process management, output streaming.
- `tests/` — unit and integration tests described above.
- `PROJECT_SPEC.md` — this spec.

Open questions / assumptions
--------------------------
- Assumed implementation language: Go (based on repo), using `github.com/pelletier/go-toml` or `BurntSushi/toml` for parsing.
- Assume running commands through `/bin/sh -c` by default for Unix-like systems.
- Default timeout for graceful shutdown: 5s (configurable later).

Next steps
----------
- Implement the config parser and unit tests for it.
- Implement a basic runner that starts tasks and streams output.
- Add signal handling and graceful shutdown.

Completion
----------
This file defines the project behaviour and acceptance criteria for `prun`. Implementers should follow the acceptance criteria for the MVP and add optional features later.
