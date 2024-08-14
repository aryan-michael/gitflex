package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Account struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Alias string `json:"alias"`
}

type model struct {
	accounts       []Account
	currentAccount Account
	step           int
	textInput      textinput.Model
	inputField     string
	accountList    list.Model
	displayMessage string
}

const configFile = "gitswitch_config.json"

func (a Account) Title() string       { return a.Alias }
func (a Account) Description() string { return fmt.Sprintf("%s (%s)", a.Name, a.Email) }
func (a Account) FilterValue() string { return a.Alias }

func initialModel() model {
	ti := textinput.New()
	ti.Focus()

	m := model{
		step:      1,
		textInput: ti,
	}

	// Load existing configuration
	if err := m.loadConfig(); err != nil {
		// If config doesn't exist, set up initial configuration
		currentName, _ := getGitConfig("user.name")
		currentEmail, _ := getGitConfig("user.email")
		m.accounts = append(m.accounts, Account{
			Name:  strings.TrimSpace(currentName),
			Email: strings.TrimSpace(currentEmail),
			Alias: "Default",
		})
		m.saveConfig()
	}

	// Detect current account
	currentName, _ := getGitConfig("user.name")
	currentEmail, _ := getGitConfig("user.email")
	currentName = strings.TrimSpace(currentName)
	currentEmail = strings.TrimSpace(currentEmail)

	for _, account := range m.accounts {
		if account.Name == currentName && account.Email == currentEmail {
			m.currentAccount = account
			break
		}
	}

	if m.currentAccount.Name == "" {
		m.currentAccount = Account{Name: currentName, Email: currentEmail, Alias: "Unknown"}
	}

	m.accountList = list.New(m.getAccountItems(), list.NewDefaultDelegate(), 0, 0)
	m.accountList.Title = "Select an account to switch to"

	return m
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			m.handleEnter()
		}
	}

	if m.step == 4 {
		m.accountList, cmd = m.accountList.Update(msg)
	} else {
		m.textInput, cmd = m.textInput.Update(msg)
	}
	return m, cmd
}

func (m *model) handleEnter() {
	switch m.step {
	case 1:
		m.step = 2
		m.inputField = "action"
		m.textInput.Placeholder = "Enter 'list' to see accounts, 'add' to add a new account, or 'switch' to switch accounts"
		m.textInput.SetValue("")
	case 2:
		action := strings.ToLower(m.textInput.Value())
		switch action {
		case "list":
			m.displayAccounts()
			m.step = 1
		case "add":
			m.step = 3
			m.inputField = "name"
			m.textInput.Placeholder = "Enter account name"
			m.textInput.SetValue("")
		case "switch":
			m.step = 4
		default:
			m.displayMessage = "Invalid action. Please enter 'list', 'add', or 'switch'."
			m.step = 1
		}
	case 3:
		if m.inputField == "name" {
			m.currentAccount.Name = m.textInput.Value()
			m.inputField = "email"
			m.textInput.Placeholder = "Enter account email"
			m.textInput.SetValue("")
		} else if m.inputField == "email" {
			m.currentAccount.Email = m.textInput.Value()
			m.inputField = "alias"
			m.textInput.Placeholder = "Enter account alias"
			m.textInput.SetValue("")
		} else {
			m.currentAccount.Alias = m.textInput.Value()
			m.accounts = append(m.accounts, m.currentAccount)
			m.saveConfig()
			m.displayMessage = fmt.Sprintf("Added new account: %s (%s)", m.currentAccount.Name, m.currentAccount.Email)
			m.step = 1
		}
	case 4:
		selectedAccount := m.accountList.SelectedItem().(Account)
		switchAccount(selectedAccount.Name, selectedAccount.Email)
		m.currentAccount = selectedAccount
		m.displayMessage = fmt.Sprintf("Switched to account: %s (%s)", selectedAccount.Name, selectedAccount.Email)
		m.step = 1
	}
}

func (m model) View() string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	accountStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))

	header := headerStyle.Render("GitSwitch CLI")
	var body string

	switch m.step {
	case 1:
		body = fmt.Sprintf("Current GitHub account: %s\n\nUsername: %s\nEmail: %s\n\nPress Enter to continue.",
			accountStyle.Render(m.currentAccount.Alias),
			accountStyle.Render(m.currentAccount.Name),
			accountStyle.Render(m.currentAccount.Email))
		if m.displayMessage != "" {
			body += "\n\n" + m.displayMessage
		}
	case 2, 3:
		body = fmt.Sprintf("%s\n%s", m.textInput.Placeholder, m.textInput.View())
	case 4:
		body = m.accountList.View()
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(fmt.Sprintf("%s\n\n%s", header, body))
}

func (m *model) loadConfig() error {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &m.accounts)
}

func (m *model) saveConfig() error {
	data, err := json.MarshalIndent(m.accounts, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configFile, data, 0644)
}

func (m *model) displayAccounts() {
	m.displayMessage = "Saved accounts:\n"
	for _, account := range m.accounts {
		m.displayMessage += fmt.Sprintf("- %s: %s (%s)\n", account.Alias, account.Name, account.Email)
	}
}

func (m *model) getAccountItems() []list.Item {
	items := make([]list.Item, len(m.accounts))
	for i, account := range m.accounts {
		items[i] = account
	}
	return items
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
