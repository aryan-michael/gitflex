package main

import (
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	account string
}

func initialModel() model {
	currentAccount, _ := getCurrentAccount()
	return model{account: currentAccount}
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
		case "1":
			switchAccount("personal")
			m.account = "personal"
		case "2":
			switchAccount("work")
			m.account = "work"
		}
	}
	return m, nil
}

func (m model) View() string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	accountStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))

	header := headerStyle.Render("Welcome to GitSwitch")
	currentAccount := fmt.Sprintf("Current GitHub account: %s", accountStyle.Render(m.account))
	options := "Press 1 to switch to personal account\nPress 2 to switch to work account\nPress q to quit"

	return lipgloss.NewStyle().Padding(1, 2).Render(fmt.Sprintf("%s\n\n%s\n\n%s", header, currentAccount, options))
}

func switchAccount(account string) {
	var username, email string
	if account == "personal" {
		username = "MihirDharaiya"
		email = "mdharaiya800@gmail.com"
	} else {
		username = "projectcoop1907"
		email = "projectcoop1907@gmail.com"
	}
	exec.Command("git", "config", "--global", "user.name", username).Run()
	exec.Command("git", "config", "--global", "user.email", email).Run()
}

func getCurrentAccount() (string, error) {
	username, err := exec.Command("git", "config", "--global", "user.name").Output()
	if err != nil {
		return "", err
	}
	return string(username), nil
}
