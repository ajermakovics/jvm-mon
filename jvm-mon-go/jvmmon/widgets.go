package jvmmon

import (
	"fmt"
	ui "github.com/gizak/termui" // <- ui shortcut, optional
	"strconv"
)

// "log"

func NewNavTable(data map[string]JVM, borderLabel string, rowCount int) *ui.Table {
	labels := []string{"PID", "Ver.", "Main"}
	rows := [][]string{labels}
	for key, jvm := range data {
		rows = append(rows, []string{key, jvm.Version, jvm.ProcName})
	}
	for i := len(rows); i < rowCount + 1; i++ {
		rows = append(rows, []string{"", "", ""})
	}

	table := ui.NewTable()
	table.Rows = rows
	table.FgColor = ui.ColorWhite
	table.BgColor = ui.ColorDefault
	table.Separator = false
	table.Analysis()
	table.SetSize()
	table.Y = 0
	table.X = 0
	table.Border = true
	table.BorderLabel = borderLabel

	if len(data) == 0 {
		return table
	}
	selected := 1
	table.BgColors[selected] = ui.ColorBlue

	ui.Handle("/sys/kbd/<up>", func(e ui.Event) { // <up>, <down>, <enter>, <escape> C-8 (delete)
		if selected > 1 {
			table.BgColors[selected] = ui.ColorDefault
			selected -= 1
			table.BgColors[selected] = ui.ColorBlue
			ui.Render(table)
		}
	})

	ui.Handle("/sys/kbd/<down>", func(e ui.Event) {
		if selected < len(data) {
			table.BgColors[selected] = ui.ColorDefault
			selected += 1
			table.BgColors[selected] = ui.ColorBlue
			ui.Render(table)
		}
	})

	ui.Handle("/sys/kbd/<enter>", func(e ui.Event) {
		key := rows[selected][0]
		ui.SendCustomEvt("/nav-table/selected", key)
	})

	return table
}

func NewThreadTable(rowCount int) *ui.Table {
	labels := []string{"Id", "Name", "State", "CpuTime"}
	rows := [][]string{labels}

	for i := 0; i < rowCount; i++ {
		rows = append(rows, []string{"  ", "                 ", "        ", "    "})
	}

	table := ui.NewTable()
	table.Rows = rows
	table.FgColor = ui.ColorWhite
	table.BgColor = ui.ColorDefault
	table.Separator = false
	table.Analysis()
	table.SetSize()
	table.Y = 0
	table.X = 0
	table.Border = true
	table.BorderLabel = "Threads"

	ui.Handle("/metrics/threads", func(e ui.Event) {
		threads := e.Data.(Threads)
		threadArr := threads.Threads
		table.BorderLabel = "Threads (" + strconv.Itoa(threads.Count) + ")"

		rows := [][]string{threads.labels()}
		for idx, thread := range threadArr {
			rows = append(rows, thread.toRow())
			if idx == rowCount-1 {
				break
			}
		}
		table.Rows = rows
		//table.SetSize()
		ui.Render(table)
	})

	return table
}

func (t *Threads) labels() []string {
	return []string{"Id", "Name", "State", "CpuTime"}
}

func (t *Thread) toRow() []string {
	cpuTime := strconv.FormatInt(t.CpuTime / 1000, 10)
	if t.CpuTime == 0 {
		cpuTime = ""
	}
	return []string{strconv.FormatInt(t.Id, 10), t.Name, t.State, cpuTime}
}

func NewMemChart() *ui.LineChart {
	chart := ui.NewLineChart()
	chart.BorderLabel = "Memory"
	chart.Data = []float64{}
	chart.Width = 50
	chart.Height = 12
	chart.DotStyle = '+'
	chart.X = 0
	chart.Y = 0
	chart.AxesColor = ui.ColorWhite
	chart.LineColor = ui.ColorGreen | ui.AttrBold

	ui.Handle("/metrics/mem", func(e ui.Event) {
		metrics := e.Data.(Metrics)
		chart.Data = append(chart.Data, metrics.Used)
		chart.BorderLabel = fmt.Sprintf("Used: %d, Max: %d MB", int(metrics.Used), int(metrics.Max))
		ui.Render(chart)
	})

	ui.Handle("/metrics/mem/clear", func(e ui.Event) {
		chart.Data = []float64{}
		chart.BorderLabel = fmt.Sprintf("Memory")
		ui.Render(chart)
	})

	return chart
}

func NewCpuChart() *ui.LineChart {
	chart := ui.NewLineChart()
	chart.BorderLabel = "CPU %"
	chart.Data = []float64{}
	chart.Width = 50
	chart.Height = 12
	chart.X = 0
	chart.Y = 0
	chart.AxesColor = ui.ColorWhite
	chart.LineColor = ui.ColorYellow

	ui.Handle("/metrics/cpu", func(e ui.Event) {
		metrics := e.Data.(Metrics)
		chart.Data = append(chart.Data, metrics.Load)
		chart.BorderLabel = fmt.Sprintf("CPU: %d ", int(metrics.Load)) + "%"
		ui.Render(chart)
	})

	ui.Handle("/metrics/cpu/clear", func(e ui.Event) {
		chart.Data = []float64{}
		chart.BorderLabel = "CPU %"
		ui.Render(chart)
	})

	return chart
}
