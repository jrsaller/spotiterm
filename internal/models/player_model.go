package models

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zmb3/spotify/v2"

	"github.com/jrsaller/spotiterm/internal/helpers"
)

type playermodel struct {
	buttons      []string
	cursor       int
	player       *spotify.Client
	help         help.Model
	keys         helpers.KeyMap
	progress_bar progress.Model
}

const REPEAT = "REP"
const SHUFFLE = "SHF"

func initialplayermodel(player *spotify.Client) playermodel {
	return playermodel{
		buttons:      []string{"◁◁", "||", "▷▷", REPEAT, SHUFFLE},
		cursor:       1,
		player:       player,
		help:         help.New(),
		keys:         helpers.Keys,
		progress_bar: progress.New(progress.WithSolidFill(spotifyGreen), progress.WithoutPercentage()),
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m playermodel) Init() tea.Cmd {
	return tickCmd()
}

func (m playermodel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
		m.progress_bar.Width = msg.Width / 3
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Right):
			m.cursor++
			if m.cursor >= len(m.buttons) {
				m.cursor = 0
			}
		case key.Matches(msg, m.keys.Left):
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.buttons) - 1
			}
		case key.Matches(msg, m.keys.Select):
			currentPlay, err := m.player.PlayerState(context.Background())
			if err != nil {
				log.Fatal(err)
			}
			switch m.cursor {
			case 0:
				m.player.Previous(context.Background())
				return m, tickCmd()
			case 1:
				if err != nil {
					log.Fatal(err)
				}
				if currentPlay.Playing {
					m.player.Pause(context.Background())
					return m, tickCmd()
				} else {
					m.player.Play(context.Background())
					return m, tickCmd()
				}
			case 2:
				m.player.Next(context.Background())
				return m, tickCmd()
			case 3:
				switch currentPlay.RepeatState {
				case "track":
					m.player.Repeat(context.Background(), "context")
				case "context":
					m.player.Repeat(context.Background(), "off")
				case "off":
					m.player.Repeat(context.Background(), "track")
				}
				return m, tickCmd()
			case 4:
				if err != nil {
					log.Fatal(err)
				}
				m.player.Shuffle(context.Background(), !currentPlay.ShuffleState)
			}
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
		return m, cmd
	case tickMsg:
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
	BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).
	Width(12).
	Align(lipgloss.Center)

var asciiBox = lipgloss.NewStyle().
	MarginRight(4).BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63"))

var activeBox = box.Background(lipgloss.Color(spotifyGreen)).Foreground(lipgloss.Color("white"))

var artHeight = 15
var artWidth = int(float64(artHeight) * 2.5)

var nowPlayingBox = lipgloss.NewStyle().Bold(true).
	PaddingTop(2).PaddingLeft(4).
	Height(artHeight).
	Width(65).
	AlignVertical(lipgloss.Center)

func getTimeString(currentPlay *spotify.PlayerState) string {
	now := time.Duration(currentPlay.Progress) * time.Millisecond
	total := time.Duration(currentPlay.Item.Duration) * time.Millisecond
	nowMinutes := int(now.Minutes())
	nowSeconds := int(now.Seconds()) % 60
	totalMinutes := int(total.Minutes())
	totalSeconds := int(total.Seconds()) % 60
	nowStr := fmt.Sprintf("%02d:%02d", nowMinutes, nowSeconds)
	totalStr := fmt.Sprintf("%02d:%02d", totalMinutes, totalSeconds)
	return fmt.Sprintf("  %s / %s", nowStr, totalStr)

}
func (m playermodel) View() string {
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
	asciialbum := helpers.GenerateASCIIArt(currentPlay.Item.Album.Images[0].URL, artWidth, artHeight)
	nowPlaying := nowPlayingBox.Render(fmt.Sprintf("Now Playing:\n\n%s\nby %s\n\nDevice: %s", currentPlay.Item.Name, currentPlay.Item.Artists[0].Name, currentPlay.Device.Name))
	if currentPlay.Playing {
		m.buttons[1] = "||"
	} else {
		m.buttons[1] = "▷"
	}
	renderItems := []string{}
	for _, button := range m.buttons {
		currentBox := box
		text := button

		if button == REPEAT {
			if currentPlay.RepeatState == "track" {
				text += " (T)"
				currentBox = activeBox
			} else if currentPlay.RepeatState == "context" {
				text += " (C)"
				currentBox = activeBox
			}
		}
		if button == SHUFFLE {
			if currentPlay.ShuffleState {
				currentBox = activeBox
			}
		}
		if button == m.buttons[m.cursor] {
			currentBox = currentBox.BorderForeground(lipgloss.Color(spotifyGreen))
		}

		renderItems = append(renderItems, currentBox.Render(text))
	}

	pl_position := float64(currentPlay.Progress) / float64(currentPlay.Item.Duration)

	return lipgloss.JoinHorizontal(lipgloss.Top, asciiBox.Render(asciialbum), nowPlaying) +
		"\n" + lipgloss.NewStyle().Padding(1, 0).Render(m.progress_bar.ViewAs(pl_position)+getTimeString(currentPlay)) +
		"\n" + lipgloss.JoinHorizontal(lipgloss.Center, renderItems...) +
		"\n" + m.help.View(m.keys)
}
