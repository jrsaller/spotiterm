package models

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zmb3/spotify/v2"

	"github.com/jrsaller/spotiterm/internal/helpers"
)

type searchmodel struct {
	player    *spotify.Client
	search    textinput.Model
	spinner   spinner.Model
	Tracks    []spotify.FullTrack
	searching bool
	cursor    int
	err       error
}

func initsearchmodel(sp_client *spotify.Client) searchmodel {
	searchInput := textinput.New()
	searchInput.Placeholder = "Search"
	searchInput.Focus()
	searchInput.CharLimit = 50
	searchInput.Width = 50

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(spotifyGreen))

	return searchmodel{
		player:  sp_client,
		search:  searchInput,
		spinner: s,
		err:     nil,
	}
}

func (m searchmodel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m searchmodel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.search.Value() == "" {
				return m, nil
			}
			if m.Tracks != nil {
				track := m.Tracks[m.cursor]
				// play the selected track
				m.player.PlayOpt(context.Background(), &spotify.PlayOptions{
					URIs: []spotify.URI{track.URI},
				})
			}

		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown:
			if m.cursor < len(m.Tracks)-1 {
				m.cursor++
			}
		}
		m.search, cmd = m.search.Update(msg)
		if m.search.Value() == "" {
			m.Tracks = nil
			return m, nil
		}
		m.searching = true
		res, err := m.player.Search(context.Background(), m.search.Value(), spotify.SearchTypeTrack)
		if err != nil {
			m.err = err
			return m, nil
		}
		if res.Tracks != nil {
			m.Tracks = res.Tracks.Tracks
		}
		m.searching = false
		cmds = append(cmds, cmd)
	case error:
		m.err = msg
		return m, nil
	}
	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m searchmodel) View() string {
	final := m.search.View() + "\n"
	if m.searching {
		final += "\n" + m.spinner.View() + "  Searching"
	} else {
		for i, track := range m.Tracks {
			t := fmt.Sprintf("%s - %s", track.Artists[0].Name, track.Name)
			if i == m.cursor {
				final += lipgloss.NewStyle().Foreground(lipgloss.Color(spotifyGreen)).Render("▶ ") + t + "\n"
			} else {
				final += "  " + t + "\n"
			}
		}
	}
	var artHeight = 15
	var artWidth = int(float64(artHeight) * 2.5)
	art := ""
	if len(m.Tracks) > 0 {
		art = helpers.GenerateASCIIArt(m.Tracks[m.cursor].Album.Images[0].URL, artWidth, artHeight)
	}
	currentPlay, err := m.player.PlayerState(context.Background())
	if err != nil {
		return "Error getting player state"
	}
	nowplaying := fmt.Sprintf("Now Playing: %s - %s \t %s", currentPlay.Item.Artists[0].Name, currentPlay.Item.Name, getTimeString(currentPlay))
	return lipgloss.JoinHorizontal(lipgloss.Center, final, art) + "\n" + nowplaying
}

func (m searchmodel) Reset() {
	
}
