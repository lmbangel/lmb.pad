package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Task struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Urgency     string    `json:"urgency"`
	Status      string    `json:"status"`
	AssignedBy  string    `json:"assigned_by"`
	Attachments []string  `json:"attachments"`
	Comments    []string  `json:"comments"`
	TimeStamp   time.Time `json:"timestamp"`
	StartTime   time.Time `json:"starttime"`
	EndTime     time.Time `json:"endtime"`
	Due         time.Time `json:"due"`
}

type ToDo struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Attachments []string  `json:"attachments"`
	Comments    []string  `json:"comments"`
	TimeStamp   time.Time `json:"timestamp"`
	StartTime   time.Time `json:"starttime"`
	EndTime     time.Time `json:"endtime"`
}

func createTask(task *Task) {
	if _, err := os.Stat("db/tasks.json"); os.IsNotExist(err) {
		os.Create("db/tasks.json")
	}

	b, err := os.ReadFile("db/tasks.json")
	if err != nil {
		log.Fatal(err)
	}

	var tasks []*Task
	if len(b) != 0 {
		err = json.Unmarshal(b, &tasks)
		if err != nil {
			log.Fatal(err)
		}
	}
	tasks = append(tasks, task)
	data, err := json.Marshal(tasks)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile("db/tasks.json", data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle.Copy()
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))

	focusedCancelButton = focusedStyle.Copy().Render("[ Cancel ]")
	blurredCancelButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Cancel"))
)

type model struct {
	focusIndex int
	inputs     []textinput.Model
	cursorMode cursor.Mode
}

func initialModel() model {
	m := model{
		inputs: make([]textinput.Model, 6),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Title"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "Description"
			t.CharLimit = 64
		case 2:
			t.Placeholder = "Urgency"
			t.CharLimit = 64
		case 3:
			t.Placeholder = "Status"
			t.CharLimit = 64
		case 4:
			t.Placeholder = "Assigned By ( Email )"
			t.CharLimit = 64
		case 5:
			t.Placeholder = "Comments"
			t.CharLimit = 64
			// case 2:
			// 	t.Placeholder = "Password"
			// 	t.EchoMode = textinput.EchoPassword
			// 	t.EchoCharacter = '•'
		}

		m.inputs[i] = t
	}

	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Change cursor mode
		case "ctrl+r":
			m.cursorMode++
			if m.cursorMode > cursor.CursorHide {
				m.cursorMode = cursor.CursorBlink
			}
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				cmds[i] = m.inputs[i].Cursor.SetMode(m.cursorMode)
			}
			return m, tea.Batch(cmds...)

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "left", "right", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" && m.focusIndex == len(m.inputs) {
				return m, tea.Quit
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" || s == "left" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs)+1 {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m model) View() string {
	var b strings.Builder
	b.WriteString(focusedStyle.Render("  Create A New Task \n ____________________\n"))
	b.WriteRune('\n')
	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	cancelButton := &blurredCancelButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	if m.focusIndex == len(m.inputs)+1 {
		button = &blurredButton
		cancelButton = &focusedCancelButton
	}

	fmt.Fprintf(&b, "\n\n%s", *button)
	fmt.Fprintf(&b, "\t%s\n\n", *cancelButton)

	b.WriteString(helpStyle.Render("cursor mode is "))
	b.WriteString(cursorModeHelpStyle.Render(m.cursorMode.String()))
	b.WriteString(helpStyle.Render(" (ctrl+r to change style)"))

	return b.String()
}

func main() {
	// taskJson := `{"title": "royalty reposrts", "description": "please fix broken ropyalty reports", "urgency": "normal", "status": "pending","assigned_by": "lihle@kanso.co.za", "attachments": "", "comments" :"", "timstamp": "", "starttime": "", "endtime": "", "due" :""}`
	// task := new(Task)
	// json.Unmarshal([]byte(taskJson), &task)
	// createTask(task)
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}

}
