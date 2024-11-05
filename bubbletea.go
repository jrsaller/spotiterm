package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zmb3/spotify/v2"
)

var spotifyGreen = "#1ED760"

type model struct {
	choices  []string
	cursor   int
	player   *spotify.Client
	help     help.Model
	keys     keyMap
	progress_bar progress.Model
}


func initialModel(player *spotify.Client) model {
	return model{
		choices:  []string{"◁◁", "||", "▷▷"},
		cursor:   1,
		player:   player,
		help:     help.New(),
		keys:     keys,
		progress_bar: progress.New(progress.WithSolidFill(spotifyGreen), progress.WithoutPercentage()),
	}
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
		m.progress_bar.Width = msg.Width / 3
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Right):
			m.cursor++
			if m.cursor >= len(m.choices) {
				m.cursor = 0
			}
		case key.Matches(msg, m.keys.Left):
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.choices) - 1
			}
		case key.Matches(msg, m.keys.Select):
			switch m.cursor {
			case 0:
				m.player.Previous(context.Background())
				return m, tickCmd()
			case 1:
				playstatus, err := m.player.PlayerCurrentlyPlaying(context.Background())
				if err != nil {
					log.Fatal(err)
				}
				if playstatus.Playing {
					m.player.Pause(context.Background())
					return m, tickCmd()
				} else {
					m.player.Play(context.Background())
					return m, tickCmd()
				}
			case 2:
				m.player.Next(context.Background())
				return m, tickCmd()
			}
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	case tickMsg:
		if m.progress_bar.Percent() == 1.0 {
			return m, tea.Quit
		}
		
		return m, tickCmd()

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := m.progress_bar.Update(msg)
		m.progress_bar = progressModel.(progress.Model)
		return m, cmd
	}

	return m, nil
}

// styles
var box = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63")).
	Width(8).AlignHorizontal(lipgloss.Center)

var selectedBox = box.BorderForeground(lipgloss.Color(spotifyGreen))

var asciiBox = lipgloss.NewStyle().
				MarginRight(4).
				Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63"))

var artHeight = 15
var artWidth = int(float64(artHeight) * 2.5)
				
var nowPlayingBox = lipgloss.NewStyle().Bold(true).
				PaddingTop(2).PaddingLeft(4).
				Height(artHeight).
				Width(65).
				AlignVertical(lipgloss.Center)



func (m model) View() string {
	currentPlay, err := m.player.PlayerState(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("Current play:\n%+v\n", currentPlay)
	if currentPlay.Device.ID == "" {
		return "No device connected"
	} else if currentPlay.Item == nil {
		return "Podcast control is not supported at this time"
	}
	asciialbum := GenerateASCIIArt(currentPlay.Item.Album.Images[0].URL, artWidth, artHeight)
	nowPlaying := nowPlayingBox.Render(fmt.Sprintf("Now Playing:\n\n%s\nby %s\n\nDevice: %s", currentPlay.Item.Name, currentPlay.Item.Artists[0].Name, currentPlay.Device.Name))
	if currentPlay.Playing {
		m.choices[1] = "||"
	} else {
		m.choices[1] = "▷"
	}
	renderItems := []string{}
	for _, choice := range m.choices {
		if choice == m.choices[m.cursor] {
			renderItems = append(renderItems, selectedBox.Render(choice))
		} else {
			renderItems = append(renderItems, box.Render(choice))
		}
	}

	now := float64(currentPlay.Progress)
	total := float64(currentPlay.Item.Duration)
	pl_position := now / total

	return lipgloss.JoinHorizontal(lipgloss.Top, asciiBox.Render(asciialbum), nowPlaying) +
		"\n" + m.progress_bar.ViewAs(pl_position) +
		"\n" + lipgloss.JoinHorizontal(lipgloss.Center, renderItems...) +
		"\n" + m.help.View(m.keys)
}

func DisplayContent(sp_client *spotify.Client) {
	p := tea.NewProgram(initialModel(sp_client), tea.WithAltScreen())
	// p := tea.NewProgram(initialModel(sp_client))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
