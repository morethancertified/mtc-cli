package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/morethancertified/mtc-cli/internal/types"
)

const (
	listWidth  = 40
	detailWidth = 60
)

var (
	titleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#25A065")).
		Padding(0, 1)

	itemStyle = lipgloss.NewStyle().
		PaddingLeft(4)

	selectedItemStyle = lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(lipgloss.Color("170"))

	paginationStyle = list.DefaultStyles().PaginationStyle.
		PaddingLeft(4)

	helpStyle = list.DefaultStyles().HelpStyle.
		PaddingLeft(4).
		PaddingBottom(1)

	quitTextStyle = lipgloss.NewStyle().
		Margin(1, 0, 2, 4)

	detailBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1).
		Width(detailWidth)
)

type item struct {
	task types.Task
}

func (i item) FilterValue() string { return i.task.Title }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	var statusIcon string
	switch i.task.Status {
	case "COMPLETED":
		statusIcon = "✅"
	case "FAILED":
		statusIcon = "❌"
	default:
		statusIcon = "⏳"
	}

	str := fmt.Sprintf("%s %s", statusIcon, i.task.Title)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	list              list.Model
	viewport          viewport.Model
	choice            string
	quitting          bool
	tasks             []types.Task
	showingExplanation bool
	lessonToken       string
	resubmitting      bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(listWidth)
		m.list.SetHeight(msg.Height - 7) // Reduced by 3 more lines to account for footer
		m.viewport.Width = detailWidth
		m.viewport.Height = msg.Height - 9 // Reduced by 3 more lines to account for footer
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, list.DefaultKeyMap().Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, list.DefaultKeyMap().CursorUp):
			m.list, _ = m.list.Update(msg)
			m.showingExplanation = false // Reset explanation view when navigating
			m.updateDetail()
			return m, nil

		case key.Matches(msg, list.DefaultKeyMap().CursorDown):
			m.list, _ = m.list.Update(msg)
			m.showingExplanation = false // Reset explanation view when navigating
			m.updateDetail()
			return m, nil

		case msg.String() == "enter":
			// Toggle AI explanation display
			m.showingExplanation = !m.showingExplanation
			m.updateDetail()
			return m, nil

		case msg.String() == "r":
			// Resubmit lesson
			m.resubmitting = true
			return m, tea.Quit

		// Handle viewport scrolling for long AI explanations
		case msg.String() == "j" || msg.String() == "down":
			if m.showingExplanation {
				m.viewport.LineDown(1)
				return m, nil
			}
		case msg.String() == "k" || msg.String() == "up":
			if m.showingExplanation {
				m.viewport.LineUp(1)
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	m.updateDetail()
	return m, cmd
}

func (m *model) updateDetail() {
	if len(m.list.Items()) == 0 {
		return
	}

	selectedItem := m.list.SelectedItem()
	if selectedItem == nil {
		return
	}

	task := selectedItem.(item).task
	
	var statusText string
	switch task.Status {
	case "COMPLETED":
		statusText = "✅ COMPLETED"
	case "FAILED":
		statusText = "❌ FAILED"
	default:
		statusText = "⏳ PENDING"
	}

	content := fmt.Sprintf("Details\n\n%s\n\n", statusText)
	
	if m.showingExplanation {
		if task.AiExplanation != "" {
			// Wrap the AI explanation text to fit the viewport width
			wrappedText := wordwrap.String(task.AiExplanation, detailWidth-6) // Account for padding and borders
			content += fmt.Sprintf("%s\n", wrappedText)
		} else {
			content += "No AI explanation available.\nGrader may not have AI feedback enabled.\n"
		}
	} else {
		if task.AiExplanation != "" {
			content += "Press Enter to view AI explanation"
		} else {
			content += "No AI explanation available.\nGrader may not have AI feedback enabled."
		}
	}

	m.viewport.SetContent(content)
}

func (m model) View() string {
	if m.quitting {
		return quitTextStyle.Render("Thanks for using sprintctl!")
	}

	listView := m.list.View()
	
	// Set the detail box height to match the list height
	detailBoxHeight := m.list.Height() + 2 // Account for borders
	detailViewStyled := detailBoxStyle.Height(detailBoxHeight)
	detailView := detailViewStyled.Render(m.viewport.View())

	// Add footer with resubmit option
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("\n\nPress 'r' to resubmit lesson • Press 'q' to quit")

	mainView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		listView,
		detailView,
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		mainView,
		footer,
	)
}

func RunGradingReport(tasks []types.Task, lessonToken string) (bool, error) {
	items := make([]list.Item, len(tasks))
	for i, task := range tasks {
		items[i] = item{task: task}
	}

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, 14)
	l.Title = "Sprint Grading Report"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := model{
		list:        l,
		tasks:       tasks,
		lessonToken: lessonToken,
	}
	m.viewport = viewport.New(detailWidth-4, 15) // Reduced height to account for footer
	m.updateDetail()

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return false, err
	}

	// Check if user wants to resubmit
	if finalModel, ok := finalModel.(model); ok {
		return finalModel.resubmitting, nil
	}

	return false, nil
}
