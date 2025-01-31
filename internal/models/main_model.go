package models

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/zmb3/spotify/v2"

	"github.com/jrsaller/spotiterm/internal/helpers"
)

var spotifyGreen = "#1ED760"

type sessionView int

const (
	PLAYER sessionView = iota
	SEARCH
)

type model struct {
	help help.Model
	keys helpers.KeyMap

	sessionView sessionView
	playermodel playermodel
	searchmodel searchmodel
}

func initialModel(sp_client *spotify.Client) model {
	return model{
		help: help.New(),
		keys: helpers.Keys,

		sessionView: PLAYER,
		playermodel: initialplayermodel(sp_client),
		searchmodel: initsearchmodel(sp_client),
	}
}

type tickMsg time.Time



func (m model) Init() tea.Cmd {
	return m.playermodel.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Quit) && m.sessionView == PLAYER {
			return m, tea.Quit
		}
		if key.Matches(msg, m.keys.Quit) && m.sessionView == SEARCH {
			m.sessionView = PLAYER
			m.searchmodel.search.Reset()
			m.searchmodel.Tracks = nil
			m.searchmodel.cursor = 0
			m.searchmodel.searching = false
			m.searchmodel.err = nil
			return m, m.playermodel.Init()
		}
		if key.Matches(msg, m.keys.Search) && m.sessionView == PLAYER {
			m.sessionView = SEARCH
			return m, m.searchmodel.Init()
		}
	case tea.WindowSizeMsg:
		tea.Printf("width: %d, height: %d\n", msg.Width, msg.Height)
	}
	switch m.sessionView {
	case PLAYER:
		newmodel, newcmd := m.playermodel.Update(msg)
		nm, ok := newmodel.(playermodel)
		if !ok {
			panic("model is not playermodel")
		}
		m.playermodel = nm
		cmd = newcmd
	case SEARCH:
		newmodel, newcmd := m.searchmodel.Update(msg)
		nm, ok := newmodel.(searchmodel)
		if !ok {
			panic("model is not searchmodel")
		}
		m.searchmodel = nm
		cmd = newcmd
	}
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	switch m.sessionView {
	case PLAYER:
		return m.playermodel.View()
	case SEARCH:
		return m.searchmodel.View()
	}
	return ""
}

func StartTea(sp_client *spotify.Client) {
	p := tea.NewProgram(initialModel(sp_client), tea.WithAltScreen())
	// p := tea.NewProgram(initialModel(sp_client))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
