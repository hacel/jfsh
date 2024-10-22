package config

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/hacel/jfsh/jellyfin"
	"github.com/spf13/viper"
)

const (
	host = iota
	username
	password
)

type model struct {
	client *jellyfin.Client

	unhidden bool
	inputs   []textinput.Model
	focused  int
	err      error
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.initClient, // this short circuits the entire form if it succeeds
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case unhideForm:
		m.unhidden = true
		return m, textinput.Blink
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.focused == len(m.inputs)-1 {
				return m, m.initClient
			}
			m.focused = (m.focused + 1) % len(m.inputs)
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyShiftTab, tea.KeyCtrlP, tea.KeyUp:
			m.focused--
			// Wrap around
			if m.focused < 0 {
				m.focused = len(m.inputs) - 1
			}
		case tea.KeyTab, tea.KeyCtrlN, tea.KeyDown:
			m.focused = (m.focused + 1) % len(m.inputs)
		}
		for i := range m.inputs {
			m.inputs[i].Blur()
		}
		m.inputs[m.focused].Focus()

	// We handle errors just like any other message
	case error:
		m.err = msg
		return m, nil
	}

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

var textStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#923FAD"))

func (m model) View() string {
	if !m.unhidden {
		return ""
	}

	// TODO: make styling nicer
	doc := strings.Builder{}
	doc.WriteString(" Jellyfin\n")
	doc.WriteString("\n")
	doc.WriteString(" " + textStyle.Render("Host") + "\n")
	doc.WriteString(" " + m.inputs[host].View() + "\n")
	if m.inputs[host].Err != nil {
		doc.WriteString(" " + m.inputs[host].Err.Error() + "\n")
	}
	doc.WriteString("\n")
	doc.WriteString(" " + textStyle.Render("Username") + "\n")
	doc.WriteString(" " + m.inputs[username].View() + "\n")
	if m.inputs[username].Err != nil {
		doc.WriteString(" " + m.inputs[username].Err.Error() + "\n")
	}
	doc.WriteString("\n")
	doc.WriteString(" " + textStyle.Render("Password") + "\n")
	doc.WriteString(" " + m.inputs[password].View() + "\n")
	if m.inputs[password].Err != nil {
		doc.WriteString(" " + m.inputs[password].Err.Error() + "\n")
	}
	doc.WriteString("\n")
	if m.err != nil {
		doc.WriteString(" " + m.err.Error() + "\n\n")
	}
	return doc.String()
}

var jfClient *jellyfin.Client

type unhideForm struct{}

func (m model) initClient() tea.Msg {
	for _, input := range m.inputs {
		if input.Err != nil || input.Value() == "" {
			return unhideForm{}
		}
	}
	host, username, password := m.inputs[host].Value(), m.inputs[username].Value(), m.inputs[password].Value()
	client, err := jellyfin.NewClient(
		host,
		username,
		password,
		viper.GetString("client_name"),
		viper.GetString("device"),
		viper.GetString("device_id"),
		viper.GetString("client_version"),
		viper.GetString("token"),
		viper.GetString("userId"),
	)
	if err != nil {
		return err
	}
	jfClient = client
	viper.Set("host", host)
	viper.Set("username", username)
	viper.Set("password", password)
	viper.Set("userId", client.UserId)
	viper.Set("token", client.Token)
	viper.WriteConfig()
	viper.SafeWriteConfig()
	return tea.Quit()
}

func Run(clientName, clientVersion, cfgPath string) *jellyfin.Client {
	viper.AddConfigPath(filepath.Join(xdg.ConfigHome, "jfsh"))
	viper.SetConfigName("jfsh")
	viper.SetConfigType("yaml")
	viper.SetConfigFile(cfgPath) // doesn't override if cfgPath is empty
	viper.ReadInConfig()
	viper.Set("client_name", clientName)
	deviceId := viper.GetString("device_id")
	if deviceId == "" {
		deviceId = uuid.NewString()
		viper.Set("device_id", deviceId)
	}
	device := viper.GetString("device")
	if device == "" {
		device, _ = os.Hostname()
		viper.Set("device", device)
	}
	viper.Set("client_version", clientVersion)

	form := make([]textinput.Model, 3)
	form[host] = textinput.New()
	form[host].Focus()
	form[host].SetValue(viper.GetString("host"))
	form[host].Validate = func(s string) error {
		_, err := url.Parse(s) // this never seems to err
		return err
	}

	form[username] = textinput.New()
	form[username].SetValue(viper.GetString("username"))

	form[password] = textinput.New()
	form[password].SetValue(viper.GetString("password"))

	m := model{inputs: form}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		panic(err)
	}
	return jfClient
}
