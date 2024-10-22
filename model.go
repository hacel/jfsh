package main

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hacel/jfsh/jellyfin"
)

type model struct {
	err error

	client *jellyfin.Client

	tabs      []string
	activeTab int

	list list.Model

	playing *item
}

func initialModel(client *jellyfin.Client) model {
	m := model{
		client: client,
		tabs:   []string{"Resume", "Next Up", "Latest"},
		list:   list.New(nil, list.NewDefaultDelegate(), 0, 0),
	}
	m.list.SetShowTitle(false)
	return m
}

func (m model) Init() tea.Cmd {
	return m.fetchActiveTabItems
}
