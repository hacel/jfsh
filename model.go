package main

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hacel/jfsh/internal/config"
	"github.com/hacel/jfsh/internal/jellyfin"
)

type tab int

const (
	Resume tab = iota
	NextUp
	RecentlyAdded
	Search
	ResumeTabName        = "Resume"
	NextUpTabName        = "Next Up"
	RecentlyAddedTabName = "Recently Added"
	SearchTabName        = "Search"
)

type model struct {
	keyMap KeyMap
	help   help.Model

	width  int
	height int

	session *config.Session

	currentTab  tab
	searchInput textinput.Model

	items       []jellyfin.Item
	allItems    []jellyfin.Item
	currentItem int

	filterActive bool
	filterInput  textinput.Model

	currentSeries *jellyfin.Item

	playing *jellyfin.Item

	err     error
	spinner spinner.Model
	loading bool
}

func initialModel(session *config.Session) model {
	searchInput := textinput.New()
	searchInput.Prompt = "Search: "
	searchInput.Width = 40

	filterInput := textinput.New()
	filterInput.Prompt = "Filter: "
	filterInput.Width = 40

	m := model{
		keyMap:      defaultKeyMap(),
		help:        help.New(),
		session:     session,
		searchInput: searchInput,
		filterInput: filterInput,
		spinner:     spinner.New(spinner.WithSpinner(spinner.Dot)),
		loading:     true,
	}
	m.updateKeys()
	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchItems(),
		m.spinner.Tick,
	)
}
