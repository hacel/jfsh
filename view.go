package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	inactiveTabColor = lipgloss.Color("#000B25")
	activeTabColor   = lipgloss.Color("#923FAD")
	docStyle         = lipgloss.NewStyle().Margin(2)
	tabStyle         = lipgloss.NewStyle().Margin(0, 1, 1, 1).Padding(0, 2)
)

func (m model) View() string {
	if m.playing != nil {
		return docStyle.Render(fmt.Sprintf("Now playing %q\nExit mpv to return to menu", m.playing.Title()))
	}

	doc := strings.Builder{}
	var tabs []string
	for i, name := range m.tabs {
		color := inactiveTabColor
		if i == m.activeTab {
			color = activeTabColor
		}
		tabs = append(tabs, tabStyle.Background(color).Render(name))
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	doc.WriteString(row)
	doc.WriteString("\n")
	doc.WriteString(m.list.View())
	return docStyle.Render(doc.String())
}
