package prompt

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// selectItem is a generic item for the list model.
type selectItem struct {
	title string
	desc  string
	index int
}

func (i selectItem) FilterValue() string { return i.title + " " + i.desc }

// selectDelegate renders each item in the selection list.
type selectDelegate struct{}

func (d selectDelegate) Height() int                             { return 1 }
func (d selectDelegate) Spacing() int                            { return 0 }
func (d selectDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d selectDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(selectItem)
	if !ok {
		return
	}

	cursor := "  "
	var str string

	if index == m.Index() {
		cursor = "> "
		str = activeStyle.Render(i.title)
		if i.desc != "" {
			str += " " + faintStyle.Render("("+i.desc+")")
		}
	} else {
		str = i.title
		if i.desc != "" {
			str += " " + faintStyle.Render("("+i.desc+")")
		}
	}

	fmt.Fprint(w, cursor+str)
}

// selectModel is the bubbletea model for selection prompts.
type selectModel struct {
	list     list.Model
	label    string
	choice   int
	quitting bool
	aborted  bool
}

func newSelectModel(label string, items []selectItem) selectModel {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	l := list.New(listItems, selectDelegate{}, 60, min(len(items), 10)+2)
	l.Title = label
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = labelStyle
	l.DisableQuitKeybindings()

	return selectModel{
		list:   l,
		label:  label,
		choice: -1,
	}
}

func (m selectModel) Init() tea.Cmd {
	return nil
}

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		// Don't intercept keys when filtering
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch msg.String() {
		case "enter":
			item, ok := m.list.SelectedItem().(selectItem)
			if ok {
				m.choice = item.index
				m.quitting = true
				return m, tea.Quit
			}
		case "ctrl+c", "esc":
			m.aborted = true
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m selectModel) View() string {
	if m.quitting {
		if m.choice >= 0 {
			item := m.list.SelectedItem().(selectItem)
			return selectedStyle.Render("* "+item.title) + "\n"
		}
		return ""
	}
	return m.list.View()
}

// runSelect runs a bubbletea selection prompt and returns the chosen index.
func runSelect(label string, items []selectItem) (int, error) {
	if len(items) == 0 {
		return -1, fmt.Errorf("no items to select from")
	}

	// If only one item, auto-select it
	if len(items) == 1 {
		fmt.Println(selectedStyle.Render("* " + items[0].title))
		return 0, nil
	}

	m := newSelectModel(label, items)
	p := tea.NewProgram(m)

	final, err := p.Run()
	if err != nil {
		return -1, fmt.Errorf("selection error: %w", err)
	}

	result := final.(selectModel)
	if result.aborted {
		return -1, fmt.Errorf("selection cancelled")
	}
	if result.choice < 0 {
		return -1, fmt.Errorf("no selection made")
	}

	return result.choice, nil
}

