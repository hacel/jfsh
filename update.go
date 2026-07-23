package main

import (
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hacel/jfsh/internal/jellyfin"
	"github.com/hacel/jfsh/internal/mpv"
)

type playbackStopped struct {
	err error
}

func (m *model) playItem() tea.Cmd {
	session := m.session
	item := m.items[m.currentItem]
	if jellyfin.IsEpisode(item) {
		return func() tea.Msg {
			// get all episodes of the series and find the index of selected episode
			items, err := session.Client.GetEpisodes(item)
			if err != nil {
				return err
			}
			idx := slices.IndexFunc(items, func(i jellyfin.Item) bool {
				return item.GetId() == i.GetId()
			})
			idx = max(0, idx) // sanity check
			if err := mpv.Play(session, items, idx); err != nil {
				return playbackStopped{err}
			}
			return playbackStopped{nil}
		}
	}
	return func() tea.Msg {
		if err := mpv.Play(session, []jellyfin.Item{item}, 0); err != nil {
			return playbackStopped{err}
		}
		return playbackStopped{nil}
	}
}

type toggleWatchedResult struct {
	err error
}

func (m *model) toggleWatchedStatus() tea.Cmd {
	m.loading = true
	session := m.session
	item := m.items[m.currentItem]
	if jellyfin.Watched(item) {
		return func() tea.Msg {
			if err := session.Client.MarkAsUnwatched(item); err != nil {
				return toggleWatchedResult{err}
			}
			return toggleWatchedResult{nil}
		}
	} else {
		return func() tea.Msg {
			if err := session.Client.MarkAsWatched(item); err != nil {
				return toggleWatchedResult{err}
			}
			return toggleWatchedResult{nil}
		}
	}
}

type fetchItemsResult struct {
	items []jellyfin.Item
	err   error
}

func (m *model) fetchItems() tea.Cmd {
	m.loading = true
	session := m.session
	if m.currentSeries != nil {
		return func() tea.Msg {
			items, err := session.Client.GetEpisodes(*m.currentSeries)
			if err != nil {
				return fetchItemsResult{nil, err}
			}
			return fetchItemsResult{items, nil}
		}
	}
	switch m.currentTab {
	case Resume:
		return func() tea.Msg {
			items, err := session.Client.GetResume(session.UserID)
			if err != nil {
				return fetchItemsResult{nil, err}
			}
			return fetchItemsResult{items, nil}
		}
	case NextUp:
		return func() tea.Msg {
			items, err := session.Client.GetNextUp()
			if err != nil {
				return fetchItemsResult{nil, err}
			}
			return fetchItemsResult{items, nil}
		}
	case RecentlyAdded:
		return func() tea.Msg {
			items, err := session.Client.GetRecentlyAdded()
			if err != nil {
				return fetchItemsResult{nil, err}
			}
			return fetchItemsResult{items, nil}
		}
	case Search:
		query := m.searchInput.Value()
		return func() tea.Msg {
			if query == "" {
				return fetchItemsResult{nil, nil}
			}
			items, err := session.Client.Search(query)
			if err != nil {
				return fetchItemsResult{nil, err}
			}
			return fetchItemsResult{items, nil}
		}
	default:
		panic("oops, selected tab is not in switch statement")
	}
}

func (m *model) applyFilter() {
	if !m.filterActive || m.filterInput.Value() == "" {
		m.items = m.allItems
		if m.currentItem >= len(m.items) {
			m.currentItem = 0
		}
		return
	}

	filterText := strings.ToLower(m.filterInput.Value())
	var filtered []jellyfin.Item

	for _, item := range m.allItems {
		title := strings.ToLower(jellyfin.GetItemTitle(item))
		desc := strings.ToLower(jellyfin.GetItemDescription(item))

		if strings.Contains(title, filterText) || strings.Contains(desc, filterText) {
			filtered = append(filtered, item)
		}
	}

	m.items = filtered
	if m.currentItem >= len(m.items) {
		m.currentItem = 0
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case error:
		m.err = msg
		return m, nil

	case playbackStopped:
		if msg.err != nil {
			m.err = msg.err
		}
		m.playing = nil
		m.updateKeys()
		return m, m.fetchItems()

	case toggleWatchedResult:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		}
		return m, m.fetchItems()

	case fetchItemsResult:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		}
		m.allItems = msg.items
		m.filterInput.SetValue("")
		m.filterActive = false
		m.applyFilter()
		m.updateKeys()
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		if key.Matches(msg, m.keyMap.ForceQuit) {
			return m, tea.Quit
		}

		if m.searchInput.Focused() {
			switch {
			case key.Matches(msg, m.keyMap.CancelWhileSearching):
				m.searchInput.Blur()
				m.updateKeys()
				return m, nil
			case key.Matches(msg, m.keyMap.AcceptWhileSearching):
				m.searchInput.Blur()
				m.updateKeys()
				return m, m.fetchItems()
			}
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}

		if m.filterInput.Focused() {
			switch {
			case key.Matches(msg, m.keyMap.CancelWhileFiltering):
				m.filterInput.Blur()
				m.filterInput.SetValue("")
				m.filterActive = false
				m.applyFilter()
				m.updateKeys()
				return m, nil
			case key.Matches(msg, m.keyMap.AcceptWhileFiltering):
				m.filterInput.Blur()
				m.updateKeys()
				return m, nil
			}
			var cmd tea.Cmd
			m.filterInput, cmd = m.filterInput.Update(msg)
			m.applyFilter()
			return m, cmd
		}

		switch {
		case key.Matches(msg, m.keyMap.CursorUp):
			if m.currentItem > 0 {
				m.currentItem--
			}
			m.updateKeys()
			return m, nil
		case key.Matches(msg, m.keyMap.CursorDown):
			if m.currentItem < len(m.items)-1 {
				m.currentItem++
			}
			m.updateKeys()
			return m, nil
		case key.Matches(msg, m.keyMap.PageUp):
			jump := m.height / 5
			m.currentItem = max(0, m.currentItem-jump)
			m.updateKeys()
			return m, nil
		case key.Matches(msg, m.keyMap.PageDown):
			jump := m.height / 5
			m.currentItem = min(len(m.items)-1, m.currentItem+jump)
			m.updateKeys()
			return m, nil
		case key.Matches(msg, m.keyMap.GoToEnd):
			m.currentItem = len(m.items) - 1
			m.updateKeys()
			return m, nil
		case key.Matches(msg, m.keyMap.GoToStart):
			m.currentItem = 0
			m.updateKeys()
			return m, nil

		case key.Matches(msg, m.keyMap.NextTab):
			if m.currentTab < Search {
				m.currentTab++
			} else {
				m.currentTab = 0
			}
			m.updateKeys()
			return m, m.fetchItems()
		case key.Matches(msg, m.keyMap.PrevTab):
			if m.currentTab > 0 {
				m.currentTab--
			} else {
				m.currentTab = Search
			}
			m.updateKeys()
			return m, m.fetchItems()

		case key.Matches(msg, m.keyMap.Search):
			m.searchInput.Focus()
			m.updateKeys()
			return m, nil
		case key.Matches(msg, m.keyMap.ClearSearch):
			m.searchInput.SetValue("")
			m.updateKeys()
			return m, m.fetchItems()

		case key.Matches(msg, m.keyMap.Filter):
			m.filterActive = true
			m.filterInput.Focus()
			m.updateKeys()
			return m, nil
		case key.Matches(msg, m.keyMap.ClearFilter):
			m.filterInput.SetValue("")
			m.filterActive = false
			m.applyFilter()
			m.updateKeys()
			return m, nil

		case key.Matches(msg, m.keyMap.Select):
			item := m.items[m.currentItem]
			if jellyfin.IsSeries(item) {
				m.currentSeries = &item
				m.updateKeys()
				return m, m.fetchItems()
			}
			m.playing = &item
			m.updateKeys()
			return m, m.playItem()

		case key.Matches(msg, m.keyMap.Back):
			m.currentSeries = nil
			m.updateKeys()
			return m, m.fetchItems()

		case key.Matches(msg, m.keyMap.ShowFullHelp):
			m.help.ShowAll = !m.help.ShowAll
			m.updateKeys()
			return m, nil

		case key.Matches(msg, m.keyMap.CloseFullHelp):
			m.help.ShowAll = !m.help.ShowAll
			m.updateKeys()
			return m, nil

		case key.Matches(msg, m.keyMap.ToggleWatched):
			return m, m.toggleWatchedStatus()

		case key.Matches(msg, m.keyMap.Refresh):
			return m, m.fetchItems()

		case key.Matches(msg, m.keyMap.Quit):
			return m, tea.Quit
		default:
			return m, nil
		}
	default:
		return m, nil
	}
}
