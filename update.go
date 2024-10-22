package main

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hacel/jfsh/jellyfin"
	"github.com/hacel/jfsh/mpv"
)

func (m model) fetchActiveTabItems() tea.Msg {
	switch m.tabs[m.activeTab] {
	case "Resume":
		items, err := m.client.GetResume()
		if err != nil {
			return err
		}
		return items
	case "Next Up":
		items, err := m.client.GetNextUp()
		if err != nil {
			return err
		}
		return items
	case "Latest":
		items, err := m.client.GetLatest()
		if err != nil {
			return err
		}
		return items
	default:
		panic("oops, selected tab is not in switch statement")
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case error:
		m.err = msg

	case []jellyfin.Item:
		// Cast to item to hand off to list.Model
		items := []list.Item{}
		for _, i := range msg {
			items = append(items, item(i))
		}
		return m, m.list.SetItems(items)

	case tea.WindowSizeMsg:
		m.list.SetSize(
			msg.Width-docStyle.GetHorizontalFrameSize(),
			msg.Height-docStyle.GetVerticalFrameSize()-tabStyle.GetVerticalFrameSize()-1, // 1 for \n
		)

	case playbackStopped:
		m.playing = nil
		return m, m.fetchActiveTabItems

	case tea.KeyMsg:
		if m.list.SettingFilter() {
			break
		}
		switch msg.String() {
		case "left", "h":
			if m.activeTab > 0 {
				m.activeTab--
			}
			m.list.ResetSelected()
			return m, m.fetchActiveTabItems
		case "right", "l":
			if m.activeTab < len(m.tabs)-1 {
				m.activeTab++
			}
			m.list.ResetSelected()
			return m, m.fetchActiveTabItems
		case "enter", "space":
			item, ok := m.list.SelectedItem().(item)
			if !ok {
				panic("failed casting list.Item to `item`")
			}
			m.playing = &item
			return m, func() tea.Msg {
				mpv.Play(m.client, jellyfin.Item(item))
				return playbackStopped{}
			}
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

type playbackStopped struct{}
