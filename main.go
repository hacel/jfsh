package main

import (
	"io"
	"log"
	"log/slog"
	"path/filepath"
	"runtime/debug"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hacel/jfsh/internal/config"

	"github.com/adrg/xdg"
	"github.com/spf13/pflag"
)

var (
	version = "unknown"
	commit  = ""
	date    = ""
)

func main() {
	// try to set version from build info
	if version == "unknown" {
		if info, ok := debug.ReadBuildInfo(); ok {
			version = info.Main.Version
		}
	}

	cfgPath := pflag.StringP("config", "c", filepath.Join(xdg.ConfigHome, "jfsh", "jfsh.yaml"), "config file path")
	statePath := pflag.String("state", filepath.Join(xdg.StateHome, "jfsh", "state.yaml"), "state file path")
	debugPath := pflag.StringP("debug", "d", "", "debug log file path (enables debug logging)")
	printVersion := pflag.BoolP("version", "v", false, "show version")
	help := pflag.BoolP("help", "h", false, "show help")
	pflag.Parse()

	if *help {
		println("Usage:  jfsh [OPTIONS]")
		println()
		println("Options:")
		pflag.PrintDefaults()
		return
	}

	if *printVersion {
		println("version", version)
		if commit != "" {
			println("commit", commit)
		}
		if date != "" {
			println("date", date)
		}
		return
	}

	if *debugPath != "" {
		f, err := tea.LogToFile(*debugPath, "")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		slog.SetDefault(
			slog.New(
				slog.NewJSONHandler(log.Writer(), &slog.HandlerOptions{Level: slog.LevelDebug, ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if a.Key == slog.TimeKey {
						return slog.String(a.Key, a.Value.Time().Format(time.DateTime))
					}
					return a
				}})))
		slog.Info("enabled debug logging")
	} else {
		log.SetOutput(io.Discard)
	}

	// First, load configuration and initialize the authenticated application session.
	session := config.Run(version, *cfgPath, *statePath)
	if session == nil {
		// err handling should happen inside the config model, this means the user quit
		return
	}

	// now we can run the main bubbletea model
	p := tea.NewProgram(initialModel(session), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
