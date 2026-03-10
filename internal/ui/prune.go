package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type pruneItem struct {
	name     string
	selected bool
}

type pruneModel struct {
	items    []pruneItem
	cursor   int
	done     bool
	aborted  bool
}

func (m pruneModel) Init() tea.Cmd { return nil }

func (m pruneModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case " ", "x":
			m.items[m.cursor].selected = !m.items[m.cursor].selected
		case "a":
			allSelected := true
			for _, item := range m.items {
				if !item.selected {
					allSelected = false
					break
				}
			}
			for i := range m.items {
				m.items[i].selected = !allSelected
			}
		case "enter":
			m.done = true
			return m, tea.Quit
		case "q", "ctrl+c", "esc":
			m.aborted = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m pruneModel) View() string {
	if m.done || m.aborted {
		return ""
	}

	var b strings.Builder
	b.WriteString("Select worktrees to prune (space to toggle, a to toggle all, enter to confirm):\n\n")

	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		check := "[ ]"
		if item.selected {
			check = "[x]"
		}
		b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, check, item.name))
	}

	b.WriteString("\npress q to cancel")
	return b.String()
}

// RunPruneSelector shows a multi-select list for pruning and returns selected names.
func RunPruneSelector(names []string) ([]string, error) {
	if len(names) == 0 {
		return nil, nil
	}

	items := make([]pruneItem, len(names))
	for i, name := range names {
		items[i] = pruneItem{name: name, selected: true}
	}

	m := pruneModel{items: items}
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	result := finalModel.(pruneModel)
	if result.aborted {
		return nil, nil
	}

	var selected []string
	for _, item := range result.items {
		if item.selected {
			selected = append(selected, item.name)
		}
	}
	return selected, nil
}
