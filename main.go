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
	actionList     list.Model
	displayMessage string
}

const configFile = "gitswitch_config.json"

var (
	titleStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5F87")).Background(lipgloss.Color("#282A36")).Padding(1, 4).MarginBottom(1).Align(lipgloss.Center).Render
	headerStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#50FA7B")).Render
	accountStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#BD93F9")).Render
	placeholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD")).Render
)

func (a Account) Title() string       { return a.Alias }
func (a Account) Description() string { return fmt.Sprintf("%s (%s)", a.Name, a.Email) }
func (a Account) FilterValue() string { return a.Alias }

type Action struct {
	name        string
	description string
}

func (a Action) Title() string       { return a.name }
func (a Action) Description() string { return a.description }
func (a Action) FilterValue() string { return a.name }

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

	// Set up the account list
	m.accountList = list.New(m.getAccountItems(), list.NewDefaultDelegate(), 20, 10)
	m.accountList.Title = "Select an account to switch to"
	m.accountList.SetShowPagination(false)
	m.accountList.SetShowHelp(false)
	m.accountList.SetShowStatusBar(false)

	// Set up the action list
	actions := []list.Item{
		Action{name: "List", description: "List all accounts"},
		Action{name: "Add", description: "Add a new account"},
		Action{name: "Switch", description: "Switch to another account"},
		Action{name: "Delete", description: "Delete an account"}, // New delete option
	}
	m.actionList = list.New(actions, list.NewDefaultDelegate(), 20, 10)
	m.actionList.Title = "Select an action"
	m.actionList.SetShowPagination(false)
	m.actionList.SetShowHelp(false)
	m.actionList.SetShowStatusBar(false)

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
			return m.handleEnter()
		}
	}

	switch m.step {
	case 2:
		m.actionList, cmd = m.actionList.Update(msg)
	case 4:
		m.accountList, cmd = m.accountList.Update(msg)
	default:
		m.textInput, cmd = m.textInput.Update(msg)
	}
	return m, cmd
}

func (m model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case 1:
		m.step = 2
	case 2:
		selectedAction := m.actionList.SelectedItem().(Action)
		switch selectedAction.name {
		case "List":
			m.displayAccounts()
			m.step = 1
		case "Add":
			m.step = 3
			m.inputField = "name"
			m.textInput.Placeholder = placeholderStyle("Enter account name")
			m.textInput.SetValue("")
		case "Switch":
			m.step = 4
		case "Delete":
			m.step = 5
		}
	case 3:
		// Adding an account
		if m.inputField == "name" {
			m.currentAccount.Name = m.textInput.Value()
			m.inputField = "email"
			m.textInput.Placeholder = placeholderStyle("Enter account email")
			m.textInput.SetValue("")
		} else if m.inputField == "email" {
			m.currentAccount.Email = m.textInput.Value()
			m.inputField = "alias"
			m.textInput.Placeholder = placeholderStyle("Enter account alias")
			m.textInput.SetValue("")
		} else {
			m.currentAccount.Alias = m.textInput.Value()
			m.accounts = append(m.accounts, m.currentAccount)
			m.saveConfig()
			m.displayMessage = fmt.Sprintf("Added new account: %s (%s)", accountStyle(m.currentAccount.Name), accountStyle(m.currentAccount.Email))
			m.step = 1
		}
	case 4:
		// Switching account
		selectedAccount := m.accountList.SelectedItem().(Account)
		switchAccount(selectedAccount.Name, selectedAccount.Email)
		m.currentAccount = selectedAccount
		m.displayMessage = fmt.Sprintf("Switched to account: %s (%s)", accountStyle(selectedAccount.Name), accountStyle(selectedAccount.Email))
		return m, tea.Quit
	case 5:
		// Deleting an account
		selectedAccount := m.accountList.SelectedItem().(Account)
		if selectedAccount.Alias == "Default" {
			m.displayMessage = "You cannot delete the Default account."
			m.step = 1
		} else {
			m.deleteAccount(selectedAccount)
			m.displayMessage = fmt.Sprintf("Deleted account: %s (%s)", accountStyle(selectedAccount.Name), accountStyle(selectedAccount.Email))
			m.step = 1
		}
	}
	return m, nil
}

func (m model) View() string {
	header := titleStyle("GitSwitch")
	var body string

	switch m.step {
	case 1:
		body = fmt.Sprintf("Current GitHub account: %s\n\nUsername: %s\nEmail: %s\n\nPress Enter to continue.",
			accountStyle(m.currentAccount.Alias),
			accountStyle(m.currentAccount.Name),
			accountStyle(m.currentAccount.Email))
		if m.displayMessage != "" {
			body += "\n\n" + m.displayMessage
		}
	case 2:
		body = m.actionList.View()
	case 3:
		body = fmt.Sprintf("%s\n%s", placeholderStyle(m.textInput.Placeholder), m.textInput.View())
	case 4:
		body = m.accountList.View()
	case 5:
		// Display account details after switching and before closing
		body = fmt.Sprintf("Switched to account: %s\n\nUsername: %s\nEmail: %s\n\nPress Enter to exit.",
			accountStyle(m.currentAccount.Alias),
			accountStyle(m.currentAccount.Name),
			accountStyle(m.currentAccount.Email))
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
		m.displayMessage += fmt.Sprintf("- %s: %s (%s)\n", accountStyle(account.Alias), accountStyle(account.Name), accountStyle(account.Email))
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
func (m *model) deleteAccount(account Account) {
	for i, acc := range m.accounts {
		if acc == account {
			m.accounts = append(m.accounts[:i], m.accounts[i+1:]...)
			break
		}
	}
	m.saveConfig()
}
