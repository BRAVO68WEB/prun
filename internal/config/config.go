package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config represents the prun.toml configuration
type Config struct {
	Tasks    []string           `toml:"tasks"`
	TaskDefs map[string]TaskDef `toml:"task"`
}

// TaskDef represents a single task configuration
type TaskDef struct {
	Cmd     string            `toml:"cmd"`
	Path    string            `toml:"path"`
	Env     map[string]string `toml:"env"`
	Restart interface{}       `toml:"restart"` // bool or string
	Shell   *bool             `toml:"shell"`
}

// Load reads and parses the prun.toml file
func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	cfg.TaskDefs = make(map[string]TaskDef)

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}

	// Validate that all tasks in the list have definitions
	for _, taskName := range cfg.Tasks {
		if _, exists := cfg.TaskDefs[taskName]; !exists {
			return nil, fmt.Errorf("task '%s' listed but not defined", taskName)
		}
	}

	// Validate that all task definitions have a cmd
	for name, task := range cfg.TaskDefs {
		if task.Cmd == "" {
			return nil, fmt.Errorf("task '%s' missing required 'cmd' field", name)
		}
	}

	return &cfg, nil
}

// GetTasksToRun returns the list of tasks to run based on config and args
func (c *Config) GetTasksToRun(args []string) ([]string, error) {
	if len(args) == 0 {
		return c.Tasks, nil
	}

	// Validate that all requested tasks exist
	for _, taskName := range args {
		if _, exists := c.TaskDefs[taskName]; !exists {
			return nil, fmt.Errorf("task '%s' not defined in config", taskName)
		}
	}

	return args, nil
}
