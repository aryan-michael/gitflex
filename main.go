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
	currentName    string
	currentEmail   string
	workName       string
	workEmail      string
	personalName   string
	personalEmail  string
	step           int
	input          string
	accountType    string // "work" or "personal"
	displayMessage string
}

func initialModel() model {
	currentName, _ := getGitConfig("user.name")
	currentEmail, _ := getGitConfig("user.email")
	return model{
		currentName:  strings.TrimSpace(currentName),
		currentEmail: strings.TrimSpace(currentEmail),
		step:         1,
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
		switch msg.Type {
		case tea.KeyEnter:
			m.handleInput()
		case tea.KeyBackspace:
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		case tea.KeyRunes:
			m.input += msg.String()
		}
	}
	return m, nil
}

func (m *model) handleInput() {
	switch m.step {
	case 1:
		m.accountType = strings.ToLower(strings.TrimSpace(m.input))
		if m.accountType == "work" || m.accountType == "personal" {
			if m.accountType == "personal" {
				m.personalName = m.currentName
				m.personalEmail = m.currentEmail
			} else {
				m.workName = m.currentName
				m.workEmail = m.currentEmail
			}
			m.input = ""
			m.step = 2
		} else {
			m.displayMessage = "Invalid input. Please type 'work' or 'personal'."
			m.input = ""
		}
	case 2:
		if m.accountType == "personal" {
			m.workName = strings.TrimSpace(m.input)
		} else {
			m.personalName = strings.TrimSpace(m.input)
		}
		m.input = ""
		m.step = 3
	case 3:
		if m.accountType == "personal" {
			m.workEmail = strings.TrimSpace(m.input)
		} else {
			m.personalEmail = strings.TrimSpace(m.input)
		}
		m.input = ""
		m.step = 4
	case 4:
		switch m.input {
		case "1":
			if m.accountType == "personal" {
				m.displayMessage = "Already on personal account."
			} else {
				switchAccount(m.personalName, m.personalEmail)
				m.displayMessage = fmt.Sprintf("Switched to personal account: %s (%s)", m.personalName, m.personalEmail)
				m.accountType = "personal"
			}
		case "2":
			if m.accountType == "work" {
				m.displayMessage = "Already on work account."
			} else {
				switchAccount(m.workName, m.workEmail)
				m.displayMessage = fmt.Sprintf("Switched to work account: %s (%s)", m.workName, m.workEmail)
				m.accountType = "work"
			}
		case "q":
			os.Exit(0)
		}
		m.input = ""
	}
}

func (m model) View() string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	accountStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))

	header := headerStyle.Render("GitSwitch CLI")
	var body string

	switch m.step {
	case 1:
		body = fmt.Sprintf("Detected current GitHub account:\n\nUsername: %s\nEmail: %s\n\nWould you like to categorize this account as work or personal? (Type 'work' or 'personal')", accountStyle.Render(m.currentName), accountStyle.Render(m.currentEmail))
	case 2:
		body = "Enter username for the second account: " + m.input
	case 3:
		body = "Enter email for the second account: " + m.input
	case 4:
		currentAccount := fmt.Sprintf("Current GitHub account: %s (%s)", accountStyle.Render(m.accountType), m.getEmail())
		options := "Press 1 to switch to personal account\nPress 2 to switch to work account\nPress q to quit"
		body = fmt.Sprintf("%s\n\n%s\n\n%s", header, currentAccount, options)
	}

	if m.displayMessage != "" {
		body += "\n\n" + m.displayMessage
		m.displayMessage = ""
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(fmt.Sprintf("%s\n\n%s", header, body))
}

func (m model) getEmail() string {
	if m.accountType == "personal" {
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
