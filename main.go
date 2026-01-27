package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
	
	bar "github.com/NimbleMarkets/ntcharts/barchart"
	progress_bar "github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	lip "github.com/charmbracelet/lipgloss"
	cpu "github.com/shirou/gopsutil/v4/cpu"
)

//	CPU

type tickMsg struct {
    totalUsage float64
    perCoreUsage []float64
}

// Function to calculate cpu usage and return as a message when done
// 1sec  = 1000 Milliseconds
func tickCpu() tea.Msg {
	// Get per-core usage. cpu.Percent sleeps for the duration!
	perCore, _ := cpu.Percent(650*time.Millisecond, true)

	// Calculate total average manually to avoid sleeping twice
	var sum float64
	for _, v := range perCore {
		sum += v
	}
	avg := sum / float64(len(perCore))
	return tickMsg{
		totalUsage: avg,
		perCoreUsage: perCore,
	}
}

var (
	// The Outer Box (Round Border)
	wrapperStyle = lip.NewStyle().
			Border(lip.RoundedBorder()).
			BorderForeground(lip.Color("62"))

	// The Left Inner Pane (Just Padding)
	leftPaneStyle = lip.NewStyle().
			Padding(1, 2)

	// The Right Inner Pane (Padding + Left Border for the separator)
	rightPaneStyle = lip.NewStyle().
			Padding(1, 2).
			Border(lip.NormalBorder(), false, false, false, true). // Left Border Only
			BorderForeground(lip.Color("62"))

	// Header Text Style
	headerStyle = lip.NewStyle().
			Bold(true).
			Foreground(lip.Color("#FAFAFA")).
			MarginBottom(1)
)

// PROGRESS BAR for current CPU usage
const (
	
	maxWidth = 75
)

var cpu_style = lip.NewStyle().PaddingBottom(1).PaddingTop(1).Border(lip.RoundedBorder()).PaddingLeft(1).PaddingRight(1)

var cpu_text_style = lip.NewStyle().Bold(true)

var helpStyle = lip.NewStyle().Foreground(lip.Color("#626262")).Render

// MODEL
type model struct {
	cpu_usage float64
	cpu_cores int
	core_usages []float64
	cpu_info []cpu.InfoStat
	progress progress_bar.Model
	barchart bar.Model
	
}

// Init
// tea.Cmd is just defined type for function- i.e. Init should return a fucntion that it will run at the start of TUI
func (m model) Init() tea.Cmd {
	
	return tickCpu
}

// UPDATE
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if msg.String() == "q" || msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	
	case tickMsg:
		m.cpu_usage = msg.totalUsage
		m.core_usages = msg.perCoreUsage

		ProgCmd := m.progress.SetPercent(m.cpu_usage / 100.0)

		m.barchart = bar.New(40,10)
		// rebuild this data for every tick
		var barChartData []bar.BarData
		for i,u := range m.core_usages{

			// Colors based on load
			color :="10" // green
			if u>50{
				color = "11" //Yellow
			}
			if u>80{
				color = "9" //Red

			}

			barChartData = append(barChartData,bar.BarData{
				Label: strconv.Itoa(i+1), // For core numbers
				Values: []bar.BarValue{
					{Name:"Usage",
					Value: u,
					Style:lip.NewStyle().Foreground(lip.Color(color)),
				},
				},
			},

			
		)
		}

		// update new data in model barchart
		m.barchart.PushAll(barChartData)
		m.barchart.Draw()
		return m,tea.Batch(tickCpu,ProgCmd)

	

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width
		if m.progress.Width > maxWidth {
			m.progress.Width = maxWidth
		}
		return m, nil

	// CRITICAL ADDITION: This allows the progress bar to animate
	case progress_bar.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress_bar.Model)
		return m, cmd
	}

	return m, nil
}

// VIEW
func (m model) View() string {

	// Build Left Content (Text + Progress)
	leftStr := lip.JoinVertical(
		lip.Left,
		headerStyle.Render(fmt.Sprintf("CPU LOAD (%.2f%%)", m.cpu_usage)),
		m.progress.View(),
		"\nPress 'q' to quit.",
	)

	// Build Right Content (Header + Chart)
	rightStr := lip.JoinVertical(
		lip.Left,
		headerStyle.Render("CORES (1-12)"), // Added Header Here
		m.barchart.View(),
	)

	//  Render Panes
	leftView := leftPaneStyle.Render(leftStr)
	rightView := rightPaneStyle.Render(rightStr)

	// 4Join and Wrap
	// join them horizontally, then wrap the result in the rounded border
	joined := lip.JoinHorizontal(lip.Top, leftView, rightView)
	
	return wrapperStyle.Render(joined)
}

// MAIN
func main() {

	

	prog := progress_bar.New(progress_bar.WithScaledGradient("#FF7CCB", "#FDFF8C"),progress_bar.WithWidth(1))

	bc := bar.New(40, 10) 

	intialModel := model{
		cpu_usage: 0,
		barchart: bc,
		progress:  prog,
	}
	
	p := tea.NewProgram(intialModel)

	if _, err := p.Run(); err != nil {
		fmt.Printf("There's an error: %v", err)
		os.Exit(1)

	}
}
