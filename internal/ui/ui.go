package ui

import (
	"fmt"
	"strings"
	"time"

	"prun/internal/runner"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model implements a simple TUI with left task list and right log pane
type Model struct {
	tasks       []string
	statuses    map[string]string // "idle", "running", "done", "failed"
	logs        []string
	selected    int
	interacting bool
	width       int
	height      int
	autoScroll  bool // auto-scroll to bottom of logs
	logOffset   int  // scroll offset for logs pane
}

// StatusIcon returns the visual indicator for a task status
func StatusIcon(status string) string {
	switch status {
	case "running":
		return "▲" // triangle up
	case "done":
		return "✓" // checkmark
	case "failed":
		return "✗" // cross
	default:
		return " " // idle/pending
	}
}

// NewModel creates a new UI model
func NewModel(tasks []string) *Model {
	st := make(map[string]string)
	for _, t := range tasks {
		st[t] = "idle"
	}
	return &Model{
		tasks:      tasks,
		statuses:   st,
		logs:       []string{},
		width:      80, // default width
		height:     24, // default height
		autoScroll: true,
		logOffset:  0,
	}
}

// Msg types
type logMsg runner.LogEvent
type tickMsg time.Time

func (m *Model) Init() tea.Cmd {
	// send a tick to refresh UI every 200ms
	return tea.Batch(
		tea.Tick(time.Millisecond*200, func(t time.Time) tea.Msg { return tickMsg(t) }),
		tea.WindowSize(),
	)
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch md := msg.(type) {
	case logMsg:
		ev := runner.LogEvent(md)
		// append to logs and update status
		m.logs = append(m.logs, fmt.Sprintf("[%s] %s", ev.Task, ev.Line))
		m.statuses[ev.Task] = "running"
		// keep logs bounded
		if len(m.logs) > 500 {
			m.logs = m.logs[len(m.logs)-500:]
		}
		return m, nil
	case tea.KeyMsg:
		switch md.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.selected > 0 {
				m.selected--
				m.autoScroll = true // Reset to auto-scroll when switching tasks
			}
		case "down", "j":
			if m.selected < len(m.tasks)-1 {
				m.selected++
				m.autoScroll = true // Reset to auto-scroll when switching tasks
			}
		case "pgup":
			// Scroll logs up
			if m.logOffset > 0 {
				m.logOffset -= 10
				if m.logOffset < 0 {
					m.logOffset = 0
				}
				m.autoScroll = false
			}
		case "pgdown", " ":
			// Scroll logs down
			m.logOffset += 10
			m.autoScroll = false
		case "home":
			// Jump to top of logs
			m.logOffset = 0
			m.autoScroll = false
		case "end":
			// Jump to bottom of logs
			m.autoScroll = true
			m.logOffset = 0
		}
		return m, nil
	case tickMsg:
		// schedule next tick
		return m, tea.Tick(time.Millisecond*200, func(t time.Time) tea.Msg { return tickMsg(t) })
	case tea.WindowSizeMsg:
		m.width = md.Width
		m.height = md.Height
		// Force a full redraw on resize by returning a batch command
		return m, tea.Batch(
			tea.ClearScreen,
			tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg { return tickMsg(t) }),
		)
	}
	return m, nil
}

// View renders the UI
func (m *Model) View() string {
	// Handle very small terminal sizes gracefully
	minWidth := 60
	minHeight := 10

	if m.width < minWidth || m.height < minHeight {
		msg := fmt.Sprintf("Terminal too small. Need at least %dx%d, got %dx%d\nResize terminal to continue...",
			minWidth, minHeight, m.width, m.height)
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Padding(1, 2).
			Render(msg)
	}

	// Define colors
	yellow := lipgloss.Color("226")
	green := lipgloss.Color("10")
	red := lipgloss.Color("9")
	gray := lipgloss.Color("240")
	cyan := lipgloss.Color("14")

	// Title style
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Padding(0, 1)

	// build left column with task list
	var leftLines []string
	leftLines = append(leftLines, titleStyle.Render("Tasks"))
	leftLines = append(leftLines, "")

	for i, t := range m.tasks {
		status := m.statuses[t]
		icon := StatusIcon(status)

		// Color the icon based on status
		var iconStyled string
		switch status {
		case "running":
			iconStyled = lipgloss.NewStyle().Foreground(yellow).Render(icon)
		case "done":
			iconStyled = lipgloss.NewStyle().Foreground(green).Render(icon)
		case "failed":
			iconStyled = lipgloss.NewStyle().Foreground(red).Render(icon)
		default:
			iconStyled = lipgloss.NewStyle().Foreground(gray).Render(icon)
		}

		// Selection indicator
		prefix := " "
		taskColor := lipgloss.Color("15")
		if i == m.selected {
			prefix = ">"
			taskColor = cyan
		}

		taskStyled := lipgloss.NewStyle().Foreground(taskColor).Render(t)
		line := fmt.Sprintf(" %s %s %s", iconStyled, prefix, taskStyled)
		leftLines = append(leftLines, line)
	}

	// Calculate how many task lines can fit in available height
	// Account for: title (1) + empty line (1) + padding/borders (4) + footer (2) = 8 total overhead
	availableTaskHeight := m.height - 8
	if availableTaskHeight < 3 {
		availableTaskHeight = 3 // Minimum to show at least some tasks
	}

	// If we have more tasks than can fit, show tasks around the selected one
	displayedLeftLines := leftLines
	if len(leftLines) > availableTaskHeight+2 { // +2 for title and empty line
		// Calculate window around selected task
		// leftLines[0] = title, leftLines[1] = empty, leftLines[2+] = tasks
		selectedLineIndex := m.selected + 2 // +2 to account for title and empty line

		// Try to center the selected task in the view
		halfWindow := availableTaskHeight / 2
		startIdx := selectedLineIndex - halfWindow
		if startIdx < 2 { // Don't go before title and empty line
			startIdx = 2
		}
		endIdx := startIdx + availableTaskHeight
		if endIdx > len(leftLines) {
			endIdx = len(leftLines)
			startIdx = endIdx - availableTaskHeight
			if startIdx < 2 {
				startIdx = 2
			}
		}

		// Build display with title, then visible tasks
		displayedLeftLines = []string{leftLines[0], leftLines[1]}
		displayedLeftLines = append(displayedLeftLines, leftLines[startIdx:endIdx]...)
	}

	left := strings.Join(displayedLeftLines, "\n")

	// build right pane with recent logs
	var rightLines []string
	rightLines = append(rightLines, titleStyle.Render(fmt.Sprintf("Logs for %s", m.tasks[m.selected])))
	rightLines = append(rightLines, "")

	// Calculate available height for logs (total height - borders - padding - title - footer)
	availableHeight := m.height - 8 // 4 for borders/padding, 2 for title, 2 for footer
	if availableHeight < 5 {
		availableHeight = 5
	}

	// Calculate safe dimensions with minimums
	leftWidth := 35
	if leftWidth > m.width-10 {
		leftWidth = m.width / 3 // Use 1/3 of width if too narrow
		if leftWidth < 20 {
			leftWidth = 20
		}
	}

	rightWidth := m.width - leftWidth - 5 // 5 for spacing
	if rightWidth < 20 {
		rightWidth = 20
	}

	// Calculate max line width for wrapping (account for padding and borders)
	maxLineWidth := rightWidth - 6 // padding (2*2=4) and border (2) = 6 chars overhead
	if maxLineWidth < 10 {
		maxLineWidth = 10
	}

	if len(m.logs) == 0 {
		rightLines = append(rightLines, lipgloss.NewStyle().Foreground(gray).Render("(no logs yet)"))
	} else {
		// Filter logs for selected task
		selectedTask := m.tasks[m.selected]
		var filteredLogs []string
		for _, log := range m.logs {
			// Check if log starts with [taskname]
			if strings.HasPrefix(log, "["+selectedTask+"]") {
				// Strip the prefix for cleaner display
				cleaned := strings.TrimPrefix(log, "["+selectedTask+"] ")
				filteredLogs = append(filteredLogs, cleaned)
			}
		}

		if len(filteredLogs) == 0 {
			rightLines = append(rightLines, lipgloss.NewStyle().Foreground(gray).Render("(no logs for this task yet)"))
		} else {
			// Word wrap each log line to fit in the pane width
			var wrappedLogs []string
			for _, line := range filteredLogs {
				if len(line) <= maxLineWidth {
					wrappedLogs = append(wrappedLogs, line)
				} else {
					// Wrap long lines
					for len(line) > maxLineWidth {
						wrappedLogs = append(wrappedLogs, line[:maxLineWidth])
						line = line[maxLineWidth:]
					}
					if len(line) > 0 {
						wrappedLogs = append(wrappedLogs, line)
					}
				}
			}

			// Show only the last N lines that fit in available height
			maxLogLines := availableHeight
			start := 0
			if len(wrappedLogs) > maxLogLines {
				if m.autoScroll {
					// Show the most recent logs
					start = len(wrappedLogs) - maxLogLines
				} else {
					// Use scroll offset
					start = m.logOffset
					if start > len(wrappedLogs)-maxLogLines {
						start = len(wrappedLogs) - maxLogLines
					}
					if start < 0 {
						start = 0
					}
				}
			}

			// Only append lines that fit
			end := start + maxLogLines
			if end > len(wrappedLogs) {
				end = len(wrappedLogs)
			}
			rightLines = append(rightLines, wrappedLogs[start:end]...)
		}
	}

	right := strings.Join(rightLines, "\n")

	paneHeight := m.height - 4
	if paneHeight < 5 {
		paneHeight = 5
	}

	// style using lipgloss
	// Both panes use the SAME height to stay aligned
	// Left pane will show all tasks (no content truncation)
	leftStyle := lipgloss.NewStyle().
		Width(leftWidth).
		Height(paneHeight).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(gray).
		Padding(1, 2)

	// Right pane: Use Height to ensure logs fit exactly in available space
	rightStyle := lipgloss.NewStyle().
		Width(rightWidth).
		Height(paneHeight).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(gray).
		Padding(1, 2)

	cols := lipgloss.JoinHorizontal(lipgloss.Top, leftStyle.Render(left), rightStyle.Render(right))

	help := "q/esc: quit | ↑/↓: navigate tasks | PgUp/PgDn: scroll logs | Home/End: jump"
	if m.interacting {
		help = "Ctrl-z - Stop interacting"
	}

	footer := lipgloss.NewStyle().Foreground(gray).Padding(0, 2).Render(help)

	return cols + "\n" + footer
}

// Start starts the TUI and returns when it's finished. It accepts an events channel
// which should receive runner.LogEvent values. It runs the TUI and returns any error.
func Start(tasks []string, events <-chan runner.LogEvent) error {
	m := NewModel(tasks)

	// Use alt screen mode for cleaner rendering and resize handling
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// feed events into the TUI
	go func() {
		for ev := range events {
			p.Send(logMsg(ev))
		}
	}()

	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
