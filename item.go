package main

import (
	"fmt"
	"strings"

	"github.com/hacel/jfsh/jellyfin"
	"github.com/sj14/jellyfin-go/api"
)

// Implements bubbles/list.Item interface
type item jellyfin.Item

func (i item) Title() string {
	str := &strings.Builder{}
	switch *i.Type {
	case api.BASEITEMKIND_MOVIE:
		fmt.Fprintf(str, "%s (%d)", *i.Name.Get(), *i.ProductionYear.Get())
		if i.UserData.IsSet() && i.UserData.Get().PlayedPercentage.IsSet() {
			fmt.Fprintf(str, " [%.f%%]", *i.UserData.Get().PlayedPercentage.Get())
		}
	case api.BASEITEMKIND_EPISODE:
		fmt.Fprintf(str, "%s S%.2dE%.2d", *i.SeriesName.Get(), *i.ParentIndexNumber.Get(), *i.IndexNumber.Get())
		if i.UserData.IsSet() && i.UserData.Get().PlayedPercentage.IsSet() {
			fmt.Fprintf(str, " [%.f%%]", *i.UserData.Get().PlayedPercentage.Get())
		}
	}
	return str.String()
}

func (i item) Description() string {
	str := &strings.Builder{}
	switch *i.Type {
	case api.BASEITEMKIND_MOVIE:
		fmt.Fprintf(str, "%s", *i.Name.Get())
	case api.BASEITEMKIND_EPISODE:
		fmt.Fprintf(str, "%s", *i.Name.Get())
	}
	return str.String()
}

func (i item) FilterValue() string { return *i.Name.Get() }
