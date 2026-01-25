package main

import (
	"fmt"
	"os"
	"time"
	tea "github.com/charmbracelet/bubbletea"
	cpu "github.com/shirou/gopsutil/v4/cpu"
	progress_bar "github.com/charmbracelet/bubbles/progress"
	lip "github.com/charmbracelet/lipgloss"
)

//  CPU
// custom message to carry cpu data
type cpuMsg [] float64

//Function to calculate cpu usage and return as a message when done
// 1sec  = 1000 Milliseconds
func tickCpu() tea.Msg {
	percent_usage , _ := cpu.Percent(650*time.Millisecond,false)
	return cpuMsg(percent_usage)
}

// PROGRESS BAR for current CPU usage
const (
	padding  = 0
	maxWidth = 250
)

var helpStyle = lip.NewStyle().Foreground(lip.Color("#626262")).Render




// MODEL
type model struct{
	cpu_usage float64
	
	progress progress_bar.Model
}

// Init
// tea.Cmd is just defined type for function- i.e. Init should return a fucntion that it will run at the start of TUI
func (m model) Init() tea.Cmd{

	return tickCpu
}

// UPDATE
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
    
	case tea.KeyMsg:
		if msg.String() == "q" || msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

	case cpuMsg:
		if len(msg) > 0 {
			m.cpu_usage = msg[0]
			// Update the progress bar to the new percentage
			cmd := m.progress.SetPercent(m.cpu_usage / 100.0)
			
			// Batch: Run the bar animation AND the next CPU tick
			return m, tea.Batch(tickCpu, cmd)
		}
		return m, tickCpu

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - padding*2 - 4
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
	
	pad := lip.NewStyle().Padding(2).Render
	
	// rendering the progress bar view
	return pad(fmt.Sprintf(
		"CPU Usage: %.2f%%\n\n%s\n\nPress 'q' to quit.", 
		m.cpu_usage, 
		m.progress.View(), // <-- This prints the bar
	))
}

// MAIN
func main(){
	
	prog := progress_bar.New(progress_bar.WithScaledGradient("#FF7CCB", "#FDFF8C"))

	intialModel := model{
		cpu_usage: 0,
		progress: prog,
	}

	p := tea.NewProgram(intialModel)

	if _,err := p.Run(); err != nil {
	fmt.Printf("There's an error: %v" ,err)
	os.Exit(0)

	}
}
