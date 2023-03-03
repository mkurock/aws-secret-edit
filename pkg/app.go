package pkg

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AWSListResult struct {
	SecretList []awsResult
}

type awsResult struct {
	ARN           string   `json:"ARN"`
	Name          string   `json:"Name"`
	VersionId     string   `json:"VersionId"`
	SecretString  string   `json:"SecretString"`
	CreatedDate   string   `json:"CreatedDate"`
	VersionStages []string `json:"VersionStages"`
}

type model struct {
	selectedSecret    string
	secretList        []awsResult
	list              list.Model
	phase             string
	message           string
	secretValueBuffer string
}

func (m model) Init() tea.Cmd {
	if m.selectedSecret != "" {
		return openEditor(m.selectedSecret, true)
	} else {
		return getSecrets
	}
}

func initialModel(l list.Model, secretName string) model {
	return model{
		selectedSecret: secretName,
		secretList:     nil,
		list:           l,
		phase:          "list",
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
		case "q":
			return m, tea.Quit
		case "enter":
			if m.phase == "error" {
				return m, openEditor(m.selectedSecret, false)
			} else if m.phase == "confirmation" {
				return m, updateSecretCmd(m.selectedSecret, m.secretValueBuffer)
			} else {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					m.selectedSecret = string(i)
					return m, openEditor(m.selectedSecret, true)
				}
			}
		}
	case updateListCmd:
		m.secretList = msg.list
		newModel := []list.Item{}
		for s := range m.secretList {
			newModel = append(newModel, item(m.secretList[s].Name))
		}
		c := m.list.SetItems(newModel)
		return m, c
	case editorClosed:
		if msg.err != nil {
			return m, tea.Quit
		}
		return m, checkSecretValid(m.selectedSecret)
	case editorResult:
		if msg.error {
			m.phase = "error"
			m.message = msg.msg
		} else {
			m.phase = "confirmation"
			m.message = msg.msg
      m.secretValueBuffer = msg.value
		}
		return m, nil
	case secretUpdated:
		if msg.err == nil {
			return m, tea.Quit
		} else {
			m.phase = "error"
			m.message = fmt.Sprintf("Secret updated failed\n%v", msg.err)
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func Run(secretName string) {
	l := list.New([]list.Item{}, itemDelegate{}, 20, 15)
	l.Title = "Select a secret you want to edit"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
	p := tea.NewProgram(initialModel(l, secretName))
	if err := p.Start(); err != nil {
		fmt.Printf("could not start program: %v", err)
	}
}
