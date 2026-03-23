package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/connoryoung/awair-downloader/internal/awair"
	"github.com/connoryoung/awair-downloader/internal/domain"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true)
	labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Width(10)
	valueStyle = lipgloss.NewStyle().Width(12)
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

func indexLabel(i int) string {
	switch {
	case i <= 0:
		return "Excellent"
	case i == 1:
		return "Good"
	case i == 2:
		return "Fair"
	case i == 3:
		return "Poor"
	default:
		return "Very Poor"
	}
}

func indexLabelStyle(i int) lipgloss.Style {
	switch {
	case i <= 0:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	case i == 1:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("22"))
	case i == 2:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	case i == 3:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	}
}

type tickMsg struct{}
type fetchedMsg struct {
	reading *domain.Reading
	err     error
}

type Model struct {
	client   *awair.Client
	device   domain.Device
	interval time.Duration
	reading  *domain.Reading
	err      error
	loading  bool
}

func New(client *awair.Client, device domain.Device, interval time.Duration) Model {
	return Model{client: client, device: device, interval: interval, loading: true}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.fetch(), m.tick())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tickMsg:
		m.loading = true
		return m, tea.Batch(m.fetch(), m.tick())
	case fetchedMsg:
		m.loading = false
		m.reading = msg.reading
		m.err = msg.err
	}
	return m, nil
}

func (m Model) View() string {
	title := titleStyle.Render("Awair — " + m.device.Name)
	if m.loading && m.reading == nil {
		return title + "\n\nFetching…\n"
	}

	var updated string
	if m.reading != nil {
		updated = mutedStyle.Render("  updated " + m.reading.Timestamp.Local().Format("3:04:05 PM MST"))
	}

	header := title + updated + "\n\n"

	if m.err != nil && m.reading == nil {
		return header + errorStyle.Render("Error: "+m.err.Error()) + "\n"
	}

	r := m.reading
	rows := []struct {
		label string
		value string
		index int
	}{
		{"Score", fmt.Sprintf("%.0f", r.Score), 0},
		{"Temp", fmt.Sprintf("%.1f °F", r.Temp.Value*9/5+32), r.Temp.Index},
		{"Humidity", fmt.Sprintf("%.1f %%", r.Humidity.Value), r.Humidity.Index},
		{"CO2", fmt.Sprintf("%.0f ppm", r.CO2.Value), r.CO2.Index},
		{"VOC", fmt.Sprintf("%.0f ppb", r.VOC.Value), r.VOC.Index},
		{"PM2.5", fmt.Sprintf("%.1f µg/m³", r.PM25.Value), r.PM25.Index},
	}

	out := header
	for i, row := range rows {
		line := labelStyle.Render(row.label) + valueStyle.Render(row.value)
		if i > 0 {
			line += indexLabelStyle(row.index).Render(indexLabel(row.index))
		}
		out += line + "\n"
	}

	out += "\n" + mutedStyle.Render("Polling every "+m.interval.String()+" · q to quit")
	if m.err != nil {
		out += "\n" + errorStyle.Render("Last fetch error: "+m.err.Error())
	}
	return out
}

func (m Model) fetch() tea.Cmd {
	return func() tea.Msg {
		r, err := m.client.Latest(m.device.DeviceType, m.device.DeviceID)
		return fetchedMsg{reading: r, err: err}
	}
}

func (m Model) tick() tea.Cmd {
	return tea.Tick(m.interval, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}
