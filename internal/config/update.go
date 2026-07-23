package config

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) initSession() tea.Cmd {
	currentSettings := m.settings
	currentSettings.Host = m.inputs[hostInput].Value()
	currentSettings.Username = m.inputs[usernameInput].Value()
	currentState, clientVersion, device := m.state, m.clientVersion, m.device
	password := m.inputs[passwordInput].Value()
	return func() tea.Msg {
		session, err := createSession(currentSettings, currentState, clientVersion, device, password)
		if err != nil {
			return err
		}
		return session
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case error:
		m.err = msg
		return m, nil

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		return m, nil

	case *Session:
		host, username := m.inputs[hostInput].Value(), m.inputs[usernameInput].Value()
		if host != m.settings.Host || username != m.settings.Username {
			m.settings.Host = host
			m.settings.Username = username
			if err := writeSettings(m.configPath, m.settings); err != nil {
				m.err = err
				return m, nil
			}
		}

		// Persist authentication separately so declarative settings remain read-only.
		m.state.Token = msg.Token
		m.state.UserID = msg.UserID
		if err := writeState(m.statePath, m.state); err != nil {
			m.err = err
			return m, nil
		}
		m.session = msg
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.Type {

		case tea.KeyEnter:
			if m.currentInput == len(m.inputs)-1 {
				valid := true
				if m.inputs[hostInput].Err != nil || m.inputs[hostInput].Value() == "" {
					valid = false
				}
				if m.inputs[usernameInput].Err != nil || m.inputs[usernameInput].Value() == "" {
					valid = false
				}
				if m.inputs[passwordInput].Err != nil {
					valid = false
				}
				if valid {
					return m, m.initSession()
				}
			}
			m.currentInput = (m.currentInput + 1) % len(m.inputs)

		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyShiftTab, tea.KeyCtrlP, tea.KeyUp:
			m.currentInput--
			if m.currentInput < 0 {
				m.currentInput = len(m.inputs) - 1
			}

		case tea.KeyTab, tea.KeyCtrlN, tea.KeyDown:
			m.currentInput = (m.currentInput + 1) % len(m.inputs)
		}

		for i := range m.inputs {
			m.inputs[i].Blur()
		}
		m.inputs[m.currentInput].Focus()
	}

	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}
