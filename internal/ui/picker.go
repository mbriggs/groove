package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// PickerItem represents a selectable item in the picker.
type PickerItem struct {
	Name string
	Desc string
}

func (i PickerItem) Title() string       { return i.Name }
func (i PickerItem) Description() string { return i.Desc }
func (i PickerItem) FilterValue() string { return i.Name }

type pickerModel struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m pickerModel) Init() tea.Cmd { return nil }

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if item, ok := m.list.SelectedItem().(PickerItem); ok {
				m.choice = item.Name
			}
			m.quitting = true
			return m, tea.Quit
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height)
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m pickerModel) View() string {
	if m.quitting {
		return ""
	}
	return m.list.View()
}

// RunPicker shows an interactive fuzzy-search list and returns the selected item name.
func RunPicker(title string, items []PickerItem) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no items to pick from")
	}

	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(listItems, delegate, 60, 14)
	l.Title = title
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	m := pickerModel{list: l}
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	result := finalModel.(pickerModel)
	if result.choice == "" {
		return "", fmt.Errorf("no selection made")
	}
	return result.choice, nil
}

// Confirm asks a yes/no question and returns the answer.
func Confirm(prompt string) (bool, error) {
	fmt.Fprintf(os.Stderr, "%s [y/N] ", prompt)
	var answer string
	if _, err := fmt.Scanln(&answer); err != nil {
		return false, nil
	}
	return answer == "y" || answer == "Y" || answer == "yes" || answer == "Yes", nil
}
