package main

import (
	"io"
	"log"
	"log/slog"
	"path/filepath"
	"runtime/debug"

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
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Info("enabled debug logging")
	} else {
		log.SetOutput(io.Discard)
	}

	// first off, run a side bubbletea model that takes care of configuration and initializing the api client
	client := config.Run(version, *cfgPath)
	if client == nil {
		// err handling should happen inside the config model, this means the user quit
		return
	}

	// now we can run the main bubbletea model
	p := tea.NewProgram(initialModel(client), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
