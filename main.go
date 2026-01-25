package main

import (
	"fmt"
	"os"
	"time"
	tea "github.com/charmbracelet/bubbletea"
	cpu "github.com/shirou/gopsutil/v4/cpu"
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

// MODEL
type model struct{
	cpu_usage float64
}

// Init
// tea.Cmd is just defined type for function- i.e. Init should return a fucntion that it will run at the start of TUI
func (m model) Init() tea.Cmd{

	return tickCpu
}

// UPDATE
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd){

	switch msg :=msg.(type){
	case tea.KeyMsg:
		if msg.String() =="q" || msg.Type == tea.KeyCtrlC{
			return m,tea.Quit
		}

	case cpuMsg:
		if len(msg)>0{
			m.cpu_usage =msg[0]
		}
		return m,tickCpu

	}
	return m,nil
}

// VIEW
func (m model) View() string {
	
	return fmt.Sprintf("Current CPU usage:%2f\n\nPress '+' to increment.\nPress 'q' to quit.\n",m.cpu_usage )
}

// MAIN
func main(){
	p := tea.NewProgram(model{})

	if _,err := p.Run(); err != nil {
	fmt.Printf("There's an error: %v" ,err)
	os.Exit(0)

	}
}
