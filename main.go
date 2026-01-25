package main

import (
	"fmt"
	"os"
	tea "github.com/charmbracelet/bubbletea"
	cpu "github.com/shirou/gopsutil/v4/cpu"
)

//  CPU
// custom message to carry cpu data
type cpuMsg [] float64


// MODEL
type model struct{
	cpu_usage float64
}

// Init
func (m model) Init() tea.Cmd{
	return nil
}

// UPDATE
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd){

	switch msg :=msg.(type){
	case tea.KeyMsg:
		if msg.String() =="q" || msg.Type == tea.KeyCtrlC{
			return m,tea.Quit
		}

		
	}
	return m,nil
}

func (m model) View() string {
	percent , _ := cpu.Percent(0,false)
	return fmt.Sprintf("The cpu USAGE is:%2f\n\nPress '+' to increment.\nPress 'q' to quit.\n",percent[0] )
}

func main(){
	p := tea.NewProgram(model{})

	if _,err := p.Run(); err != nil {
	fmt.Printf("There's an error: %v" ,err)
	os.Exit(0)

	}
}
