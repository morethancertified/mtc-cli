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
	"github.com/morethancertified/mtc-cli/internal/styles"
	"github.com/morethancertified/mtc-cli/internal/types"
)

// Note: dynamic sizing is handled via WindowSizeMsg; no fixed widths

var (
	titleStyle = lipgloss.NewStyle().
		Foreground(styles.White).
		Background(styles.PrimaryBlue).
		Padding(0, 2).
		Bold(true)

	itemStyle = lipgloss.NewStyle().
		PaddingLeft(4).
		Foreground(styles.White)

	selectedItemStyle = lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(styles.White).
		Background(styles.PrimaryBlue).
		Bold(true)

	paginationStyle = list.DefaultStyles().PaginationStyle.
		PaddingLeft(4).
		Foreground(styles.Gray)

	helpStyle = list.DefaultStyles().HelpStyle.
		PaddingLeft(4).
		PaddingBottom(1).
		Foreground(styles.Gray)

	quitTextStyle = lipgloss.NewStyle().
		Margin(1, 0, 2, 4).
		Foreground(styles.PrimaryGreen).
		Bold(true)

	screenStyle = lipgloss.NewStyle().
		Background(styles.PrimaryBlue).
		Foreground(styles.White)
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

	statusIcon := styles.StatusIcon(i.task.Status)
	str := fmt.Sprintf("%s %s", statusIcon, i.task.Title)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("▶ " + strings.Join(s, " "))
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
	width             int
	height            int
	topHeight         int
	bottomHeight      int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Ensure minimum dimensions
		if m.width < 10 {
			m.width = 10
		}
		if m.height < 10 {
			m.height = 10
		}
		
		// Reserve space for separator and footer
		reservedHeight := 4 // separator + footer + padding
		availableHeight := m.height - reservedHeight
		if availableHeight < 4 {
			availableHeight = 4
		}
		
		// Split remaining space between list and feedback
		m.topHeight = availableHeight / 2
		m.bottomHeight = availableHeight - m.topHeight

		// Set list dimensions with minimum width
		listWidth := m.width - 2
		if listWidth < 1 {
			listWidth = 1
		}
		m.list.SetWidth(listWidth)
		m.list.SetHeight(m.topHeight)

		// Set viewport dimensions with minimum width
		viewportWidth := m.width - 4
		if viewportWidth < 1 {
			viewportWidth = 1
		}
		m.viewport.Width = viewportWidth
		m.viewport.Height = m.bottomHeight
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, list.DefaultKeyMap().Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, list.DefaultKeyMap().CursorUp):
			m.list, _ = m.list.Update(msg)
			m.showingExplanation = false // Reset explanation view when navigating
			m.viewport.SetContent("") // Clear viewport content completely
			m.viewport.GotoTop() // Reset scroll position
			m.updateDetail()
			return m, nil

		case key.Matches(msg, list.DefaultKeyMap().CursorDown):
			m.list, _ = m.list.Update(msg)
			m.showingExplanation = false // Reset explanation view when navigating
			m.viewport.SetContent("") // Clear viewport content completely
			m.viewport.GotoTop() // Reset scroll position
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
		m.viewport.SetContent("")
		m.viewport.GotoTop()
		return
	}

	selectedItem := m.list.SelectedItem()
	if selectedItem == nil {
		m.viewport.SetContent("")
		m.viewport.GotoTop()
		return
	}

	task := selectedItem.(item).task
	
	// Always clear viewport first and reset scroll position to prevent content remnants
	m.viewport.SetContent("")
	m.viewport.GotoTop()
	
	content := styles.SectionHeaderStyle.Render("🤖 AI FEEDBACK") + "\n\n"
	
	if m.showingExplanation {
		if task.AiExplanation != "" {
			// Wrap text based on current viewport width (fallback to a sensible width)
			wrapWidth := m.viewport.Width
			if wrapWidth <= 0 {
				wrapWidth = 80
			}
			wrappedText := wordwrap.String(task.AiExplanation, wrapWidth-4)
			content += wrappedText + "\n"
		} else {
			content += styles.WarningStyle.Render(" NO AI FEEDBACK ") + "\n"
			content += "Grader may not have AI feedback enabled.\n"
		}
	} else {
		if task.AiExplanation != "" {
			content += styles.InfoStyle.Render(" Press Enter to view AI explanation ")
		} else {
			content += styles.WarningStyle.Render(" NO AI FEEDBACK ") + "\n"
			content += "Grader may not have AI feedback enabled."
		}
	}

	m.viewport.SetContent(content)
}

func (m model) View() string {
	if m.quitting {
		return quitTextStyle.Render("Thanks for using sprintctl!")
	}

	// Top: tasks list (full width)
	listView := m.list.View()
	
	// Bottom: AI feedback with proper spacing
	feedbackContent := m.viewport.View()
	
	// Create a clean separator with proper width
	separatorWidth := m.width - 4
	if separatorWidth < 1 {
		separatorWidth = 1
	}
	separatorLine := strings.Repeat("─", separatorWidth)
	separator := lipgloss.NewStyle().
		Width(m.width).
		Background(styles.PrimaryBlue).
		Foreground(styles.Gray).
		Padding(0, 1).
		Render(separatorLine)

	// Footer with resubmit instructions
	footerText := "Press 'r' to resubmit lesson • Press 'q' to quit"
	if !m.showingExplanation {
		footerText = "Press Enter to view AI feedback • Press 'r' to resubmit • Press 'q' to quit"
	}
	footer := lipgloss.NewStyle().
		Width(m.width).
		Foreground(styles.White).
		Background(styles.DarkGray).
		Padding(0, 1).
		Render(footerText)

	// Combine everything with proper spacing
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		listView,
		separator,
		feedbackContent,
		footer,
	)

	return screenStyle.
		Width(m.width).
		Height(m.height).
		Padding(1).
		Render(content)
}

func RunGradingReport(tasks []types.Task, lessonToken string) (bool, error) {
	items := make([]list.Item, len(tasks))
	for i, task := range tasks {
		items[i] = item{task: task}
	}

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, 14)
	l.Title = "🚀 Sprint Grading Report"
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
	// Initial viewport; will be resized on WindowSizeMsg
	m.viewport = viewport.New(80, 10)
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
