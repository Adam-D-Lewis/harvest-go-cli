package prompt

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// inputModel is the bubbletea model for text input prompts.
type inputModel struct {
	textInput textinput.Model
	label     string
	validate  func(string) error
	err       string
	value     string
	quitting  bool
	aborted   bool
}

func newInputModel(label, placeholder, defaultValue string, validate func(string) error) inputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	if defaultValue != "" {
		ti.SetValue(defaultValue)
	}

	return inputModel{
		textInput: ti,
		label:     label,
		validate:  validate,
	}
}

func (m inputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			val := m.textInput.Value()
			if m.validate != nil {
				if err := m.validate(val); err != nil {
					m.err = err.Error()
					return m, nil
				}
			}
			m.value = val
			m.quitting = true
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.aborted = true
			m.quitting = true
			return m, tea.Quit
		}
	}

	// Clear error on new input
	m.err = ""
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m inputModel) View() string {
	if m.quitting {
		if m.aborted {
			return ""
		}
		return selectedStyle.Render("* "+m.label+": "+m.value) + "\n"
	}

	s := labelStyle.Render(m.label) + "\n"
	s += m.textInput.View() + "\n"
	if m.err != "" {
		s += errorStyle.Render("  "+m.err) + "\n"
	}
	return s
}

// runInput runs a bubbletea text input prompt and returns the entered value.
func runInput(label, placeholder, defaultValue string, validate func(string) error) (string, error) {
	m := newInputModel(label, placeholder, defaultValue, validate)
	p := tea.NewProgram(m)

	final, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("input error: %w", err)
	}

	result := final.(inputModel)
	if result.aborted {
		return "", fmt.Errorf("input cancelled")
	}

	return result.value, nil
}
