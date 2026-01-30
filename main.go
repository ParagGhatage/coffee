package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	bar "github.com/NimbleMarkets/ntcharts/barchart"
	progress_bar "github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	lip "github.com/charmbracelet/lipgloss"
	cpu "github.com/shirou/gopsutil/v4/cpu"
	mem "github.com/shirou/gopsutil/v4/mem"
)

//	CPU

type tickMsg struct {
    totalUsage float64
    perCoreUsage []float64
	totalMem float64
	usedMem float64
	freeMem float64
}

// Function to calculate cpu usage and return as a message when done
// 1sec  = 1000 Milliseconds
func tickCpu() tea.Msg {
	// Get per-core usage. cpu.Percent sleeps for the duration!
	perCore, _ := cpu.Percent(650*time.Millisecond, true)

	memStats,_ := mem.VirtualMemory()

	MemTotal := convertBytesToGB(memStats.Total)
	MemUsed := convertBytesToGB(memStats.Used)
	MemFree := convertBytesToGB(memStats.Available)


	// Calculate total average manually to avoid sleeping twice
	var sum float64
	for _, v := range perCore {
		sum += v
	}
	avg := sum / float64(len(perCore))
	return tickMsg{
		totalUsage: avg,
		perCoreUsage: perCore,

		totalMem: MemTotal,
		usedMem: MemUsed,
		freeMem: MemFree,
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

// MEMORY
func convertBytesToGB(B uint64) float64 {

	// deviding by 10^9 to convert it to GB
	gb := float64(B)*math.Pow10((-9))
	return(gb)
}

// MODEL
type model struct {
	cpu_usage float64
	cpu_cores int
	core_usages []float64
	cpu_info []cpu.InfoStat

	mem_total float64
	mem_free float64
	mem_used float64

	progress progress_bar.Model
	barchart bar.Model
	mem_progress progress_bar.Model
	
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

		m.mem_used = msg.totalUsage
		MemProgCmd := m.mem_progress.SetPercent(m.mem_used/m.mem_total)

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
		return m,tea.Batch(tickCpu,ProgCmd,MemProgCmd)

	

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width
		if m.progress.Width > maxWidth {
			m.progress.Width = maxWidth
		}
		// IMPORTANT: Update the memory progress width too!
        m.mem_progress.Width = m.progress.Width
		return m, nil

	// CRITICAL ADDITION: This allows the progress bar to animate
	case progress_bar.FrameMsg:
		newCpuModel, cpuCmd := m.progress.Update(msg)
        newMemModel, memCmd := m.mem_progress.Update(msg)

        m.progress = newCpuModel.(progress_bar.Model)
        m.mem_progress = newMemModel.(progress_bar.Model)

        return m, tea.Batch(cpuCmd, memCmd)
	}

	return m, nil
}

// VIEW
func (m model) View() string {

    // 1. CPU Pane
    leftStr := lip.JoinVertical(
        lip.Left,
        headerStyle.Render(fmt.Sprintf("CPU (%.2f%%)", m.cpu_usage)),
        m.progress.View(),
    )
    leftView := leftPaneStyle.Render(leftStr)

    // 2. Memory Pane (Create a formatted pane for this too)
    // We display Used / Total GB
    memLabel := fmt.Sprintf("MEM (%.2f/%.2f GB)", m.mem_used, m.mem_total)
    
    midStr := lip.JoinVertical(
        lip.Left,
        headerStyle.Render(memLabel),
        m.mem_progress.View(),
    )
    // Reuse the left style for consistency
    midView := leftPaneStyle.Render(midStr)

    // 3. Cores Pane
    rightStr := lip.JoinVertical(
        lip.Left,
        headerStyle.Render("CPU CORES"),
        m.barchart.View(),
    )
    rightView := rightPaneStyle.Render(rightStr)

    // Join all three
    joined := lip.JoinHorizontal(lip.Top, leftView, midView, rightView)

    return wrapperStyle.Render(joined)
}

// MAIN
func main() {

	

	prog := progress_bar.New(progress_bar.WithScaledGradient("#FF7CCB", "#FDFF8C"),progress_bar.WithWidth(1))

	bc := bar.New(40, 10) 

	prog_mem := progress_bar.New(progress_bar.WithScaledGradient("#c169fc", "#ff5132"),progress_bar.WithWidth(1))

	m1,_ := mem.VirtualMemory() //to calculate and render a number representing total memory


	intialModel := model{
		cpu_usage: 0,
		barchart: bc,
		progress:  prog,
		mem_progress:prog_mem,

		mem_total: convertBytesToGB(m1.Total),
		mem_used: 0,
	}

	
	
	p := tea.NewProgram(intialModel)

	if _, err := p.Run(); err != nil {
		fmt.Printf("There's an error: %v", err)
		os.Exit(1)

	}
}
