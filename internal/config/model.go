package config

import (
	"errors"
	"net/url"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// form fields
const (
	hostInput = iota
	usernameInput
	passwordInput
)

type model struct {
	session *Session
	err     error

	height int
	width  int

	inputs       []textinput.Model
	currentInput int

	settings      settings
	state         state
	clientVersion string
	configPath    string
	statePath     string
	device        string
}

func initialModel(currentSettings settings, currentState state, clientVersion, configPath, statePath, device, password string, initialErr error) model {
	form := make([]textinput.Model, 3)

	form[hostInput] = textinput.New()
	form[hostInput].Focus()
	form[hostInput].Prompt = ""
	form[hostInput].SetValue(currentSettings.Host)
	form[hostInput].Validate = func(s string) error {
		u, err := url.Parse(s)
		if err != nil {
			return errors.New("invalid format")
		}
		if u.Scheme == "" {
			return errors.New("must include scheme (http:// or https://)")
		}
		if u.Host == "" {
			return errors.New("URL must include host")
		}
		return nil
	}

	form[usernameInput] = textinput.New()
	form[usernameInput].Prompt = ""
	form[usernameInput].SetValue(currentSettings.Username)

	form[passwordInput] = textinput.New()
	form[passwordInput].Prompt = ""
	form[passwordInput].EchoMode = textinput.EchoPassword
	form[passwordInput].SetValue(password)

	return model{
		err:           initialErr,
		inputs:        form,
		settings:      currentSettings,
		state:         currentState,
		clientVersion: clientVersion,
		configPath:    configPath,
		statePath:     statePath,
		device:        device,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}
