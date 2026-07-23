// Package config loads application settings and creates an authenticated Jellyfin session.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/hacel/jfsh/internal/jellyfin"
	"gopkg.in/yaml.v3"
)

type settings struct {
	Host         string   `yaml:"host,omitempty"`
	Username     string   `yaml:"username,omitempty"`
	Device       string   `yaml:"device,omitempty"`
	SkipSegments []string `yaml:"skip_segments,omitempty"`
	PasswordFile string   `yaml:"password_file,omitempty"`
}

type state struct {
	DeviceID string `yaml:"device_id"`
	Token    string `yaml:"token,omitempty"`
	UserID   string `yaml:"user_id,omitempty"`
}

// Session contains the authenticated Jellyfin client and application configuration used with it.
type Session struct {
	Client       *jellyfin.Client
	Host         string
	UserID       string
	Token        string
	SkipSegments []string
}

func readYAML(path string, value any) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(data, value); err != nil {
		return fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return nil
}

func marshalYAML(value any) ([]byte, error) {
	data, err := yaml.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("failed to encode YAML: %w", err)
	}
	return data, nil
}

func writeSettings(path string, value settings) error {
	if info, err := os.Lstat(path); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm()&0o200 == 0 {
			return fmt.Errorf("%s is read-only; update its declarative configuration instead", path)
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	data, err := marshalYAML(value)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write settings: %w", err)
	}
	return nil
}

func writeState(path string, value state) error {
	if _, err := os.Stat(filepath.Dir(path)); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return fmt.Errorf("failed to create state directory: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to inspect state directory: %w", err)
	}
	data, err := marshalYAML(value)
	if err != nil {
		return err
	}

	// Write and rename in the same directory so interrupted writes cannot corrupt state.
	file, err := os.CreateTemp(filepath.Dir(path), ".state-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary state file: %w", err)
	}
	defer os.Remove(file.Name())
	if err := file.Chmod(0o600); err != nil {
		file.Close()
		return fmt.Errorf("failed to secure temporary state file: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		return fmt.Errorf("failed to write state: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close state: %w", err)
	}
	if err := os.Rename(file.Name(), path); err != nil {
		return fmt.Errorf("failed to replace state: %w", err)
	}
	return nil
}

func readPassword(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read password file: %w", err)
	}
	password := strings.TrimSuffix(string(data), "\n")
	return strings.TrimSuffix(password, "\r"), nil
}

func createSession(currentSettings settings, currentState state, clientVersion, device, password string) (*Session, error) {
	host, err := jellyfin.NormalizeHost(currentSettings.Host)
	if err != nil {
		return nil, err
	}
	if currentState.Token == "" || currentState.UserID == "" {
		currentState.Token, currentState.UserID, err = jellyfin.Authenticate(
			host,
			currentSettings.Username,
			password,
			device,
			currentState.DeviceID,
			clientVersion,
		)
		if err != nil {
			return nil, err
		}
	}
	client, err := jellyfin.NewClient(host, device, currentState.DeviceID, clientVersion, currentState.Token)
	if err != nil {
		return nil, err
	}
	return &Session{
		Client:       client,
		Host:         host,
		UserID:       currentState.UserID,
		Token:        currentState.Token,
		SkipSegments: currentSettings.SkipSegments,
	}, nil
}

func Run(clientVersion, configPath, statePath string) *Session {
	var currentSettings settings
	if err := readYAML(configPath, &currentSettings); err != nil {
		panic(err)
	}
	var currentState state
	if err := readYAML(statePath, &currentState); err != nil {
		panic(err)
	}

	// Resolve generated defaults in memory and persist only mutable state.
	device := currentSettings.Device
	if device == "" {
		device, _ = os.Hostname()
	}
	if currentState.DeviceID == "" {
		currentState.DeviceID = uuid.NewString()
		if err := writeState(statePath, currentState); err != nil {
			panic(err)
		}
	}

	// Read external credentials only when no reusable authentication state exists.
	password := ""
	var initialErr error
	if currentState.Token == "" || currentState.UserID == "" {
		if currentSettings.PasswordFile != "" {
			password, initialErr = readPassword(currentSettings.PasswordFile)
		}
	}
	if currentSettings.Host != "" && currentSettings.Username != "" && initialErr == nil {
		session, err := createSession(currentSettings, currentState, clientVersion, device, password)
		if err == nil {
			if session.Token != currentState.Token || session.UserID != currentState.UserID {
				currentState.Token = session.Token
				currentState.UserID = session.UserID
				if err := writeState(statePath, currentState); err != nil {
					panic(err)
				}
			}
			return session
		}
		initialErr = err
		slog.Error("failed to create session", "err", err)
	}

	// Fall back to the interactive form for missing or invalid credentials.
	m, err := tea.NewProgram(initialModel(currentSettings, currentState, clientVersion, configPath, statePath, device, password, initialErr), tea.WithAltScreen()).Run()
	if err != nil {
		panic(err)
	}
	result := m.(model)
	return result.session
}
