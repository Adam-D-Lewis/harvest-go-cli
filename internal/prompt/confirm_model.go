package prompt

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// confirmModel is the bubbletea model for y/n confirmation prompts.
type confirmModel struct {
	label    string
	result   bool
	quitting bool
	aborted  bool
}

func newConfirmModel(label string) confirmModel {
	return confirmModel{label: label}
}

func (m confirmModel) Init() tea.Cmd {
	return nil
}

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.result = true
			m.quitting = true
			return m, tea.Quit
		case "n", "N", "enter":
			m.result = false
			m.quitting = true
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.aborted = true
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m confirmModel) View() string {
	if m.quitting {
		if m.result {
			return selectedStyle.Render("* "+m.label+": yes") + "\n"
		}
		return faintStyle.Render("* "+m.label+": no") + "\n"
	}
	return labelStyle.Render(m.label) + " " + faintStyle.Render("[y/N]") + " "
}

// runConfirm runs a bubbletea y/n confirmation prompt.
func runConfirm(label string) (bool, error) {
	m := newConfirmModel(label)
	p := tea.NewProgram(m)

	final, err := p.Run()
	if err != nil {
		return false, fmt.Errorf("confirm error: %w", err)
	}

	result := final.(confirmModel)
	if result.aborted {
		return false, nil
	}

	return result.result, nil
}
