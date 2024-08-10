package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	account       string
	personalName  string
	personalEmail string
	workName      string
	workEmail     string
	step          int
	input         string
}

func initialModel() model {
	currentName, _ := getGitConfig("user.name")
	currentEmail, _ := getGitConfig("user.email")
	return model{
		account:       "personal",
		personalName:  strings.TrimSpace(currentName),
		personalEmail: strings.TrimSpace(currentEmail),
		step:          1,
	}
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			m.handleInput()
		default:
			if m.step > 1 {
				m.input += msg.String()
			}
		}
	}
	return m, nil
}

func (m *model) handleInput() {
	switch m.step {
	case 2:
		m.workName = strings.TrimSpace(m.input)
		m.input = ""
		m.step = 3
	case 3:
		m.workEmail = strings.TrimSpace(m.input)
		m.input = ""
		m.step = 4
	}
}

func (m model) View() string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	accountStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))

	header := headerStyle.Render("GitSwitch CLI")
	var body string

	switch m.step {
	case 1:
		body = fmt.Sprintf("Detected current GitHub account:\n\nUsername: %s\nEmail: %s\n\nPress 'Enter' to continue", accountStyle.Render(m.personalName), accountStyle.Render(m.personalEmail))
	case 2:
		body = "Enter username for the work account: " + m.input
	case 3:
		body = "Enter email for the work account: " + m.input
	case 4:
		currentAccount := fmt.Sprintf("Current GitHub account: %s (%s)", accountStyle.Render(m.account), m.getEmail())
		options := "Press 1 to switch to personal account\nPress 2 to switch to work account\nPress q to quit"
		body = fmt.Sprintf("%s\n\n%s\n\n%s", header, currentAccount, options)
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(fmt.Sprintf("%s\n\n%s", header, body))
}

func (m model) getEmail() string {
	if m.account == "personal" {
		return m.personalEmail
	}
	return m.workEmail
}

func switchAccount(username, email string) {
	exec.Command("git", "config", "--global", "user.name", username).Run()
	exec.Command("git", "config", "--global", "user.email", email).Run()
}

func getGitConfig(key string) (string, error) {
	cmd := exec.Command("git", "config", "--global", key)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
