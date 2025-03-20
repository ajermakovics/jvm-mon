package jvmmon

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	ui "github.com/gizak/termui/v3" // <- ui shortcut, optional
	"github.com/gizak/termui/v3/widgets"
	"strconv"
)

// "log"

func NewNavTable(data map[string]JVM, borderLabel string, rowCount int, eb EventBus.Bus) *widgets.Table {
	labels := []string{"PID", "Ver.", "Main"}
	rows := [][]string{labels}
	for key, jvm := range data {
		rows = append(rows, []string{key, jvm.Version, jvm.ProcName})
	}
	for i := len(rows); i < rowCount+1; i++ {
		rows = append(rows, []string{"", "", ""})
	}

	table := widgets.NewTable()
	table.Rows = rows
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.Border = true
	table.Title = borderLabel
	table.TextAlignment = ui.AlignLeft
	table.ColumnWidths = []int{7, 5, -1}
	table.RowSeparator = false

	if len(data) == 0 {
		return table
	}
	selected := 1
	table.RowStyles[selected] = ui.NewStyle(ui.ColorYellow)

	eb.SubscribeAsync("keyboard-events", func(e string) {
		if e == "<Up>" {
			if selected > 1 {
				table.RowStyles[selected] = ui.NewStyle(ui.ColorWhite)
				selected -= 1
				table.RowStyles[selected] = ui.NewStyle(ui.ColorYellow)
				ui.Render(table)
			}
		}

		if e == "<Down>" {
			if selected < len(data) {
				table.RowStyles[selected] = ui.NewStyle(ui.ColorWhite)
				selected += 1
				table.RowStyles[selected] = ui.NewStyle(ui.ColorYellow)
				ui.Render(table)
			}
		}

		if e == "<Enter>" {
			pid := rows[selected][0]
			eb.Publish("jvm-selected", pid)
		}
	}, false)

	return table
}

func NewThreadTable(rowCount int, eb EventBus.Bus) *widgets.Table {
	labels := []string{"Id", "Name", "State", "CpuTime"}
	rows := [][]string{labels}

	for i := 0; i < rowCount; i++ {
		rows = append(rows, []string{"  ", "                 ", "        ", "    "})
	}

	table := widgets.NewTable()
	table.Rows = rows
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.Border = true
	table.SetRect(0, 0, 20, 20)
	table.Title = "Threads"
	table.RowSeparator = false

	eb.Subscribe("metrics.Threads", func(threads Threads) {
		threadArr := threads.Threads
		table.Title = "Threads (" + strconv.Itoa(threads.Count) + ")"

		rows := [][]string{threads.labels()}
		for idx, thread := range threadArr {
			rows = append(rows, thread.toRow())
			if idx == rowCount-1 {
				break
			}
		}
		table.Rows = rows
		ui.Render(table)
	})

	return table
}

func (t *Threads) labels() []string {
	return []string{"Id", "Name", "State", "CpuTime"}
}

func (t *Thread) toRow() []string {
	cpuTime := strconv.FormatInt(t.CpuTime/1000, 10)
	if t.CpuTime == 0 {
		cpuTime = ""
	}
	return []string{strconv.FormatInt(t.Id, 10), t.Name, t.State, cpuTime}
}

func NewMemChart(eb EventBus.Bus) *widgets.SparklineGroup {
	chart := widgets.NewSparkline()
	chart.Data = []float64{}
	chart.LineColor = ui.ColorGreen
	chart.TitleStyle.Fg = ui.ColorWhite

	slg := widgets.NewSparklineGroup(chart)
	slg.Title = "Memory"

	eb.Subscribe("metrics", func(metrics Metrics) {
		maxX := slg.Bounds().Max.X
		data := append(chart.Data, metrics.Used)
		if len(data) > maxX/2 {
			data = data[1:]
		}
		chart.Data = data
		slg.Title = fmt.Sprintf("Used: %d, Max: %d MB", int(metrics.Used), int(metrics.Max))
		chart.MaxVal = metrics.Max
		ui.Render(slg)
	})

	eb.Subscribe("jvm-selected", func(pid string) {
		chart.Data = []float64{}
		slg.Title = fmt.Sprintf("Memory")
		ui.Render(slg)
	})

	return slg
}

func NewCpuChart(eb EventBus.Bus) *widgets.Plot {
	chart := widgets.NewPlot()
	chart.Data = make([][]float64, 1)
	chart.Data[0] = []float64{0, 0}
	chart.LineColors[0] = ui.ColorYellow
	chart.TitleStyle.Fg = ui.ColorWhite
	chart.AxesColor = ui.ColorWhite
	chart.PlotType = widgets.LineChart
	//chart.MaxVal = 100.0

	eb.Subscribe("metrics", func(metrics Metrics) {
		maxX := chart.Bounds().Max.X
		data := append(chart.Data[0], metrics.Load)
		if len(data) > maxX/2 {
			data = data[1:]
		}
		chart.Data[0] = data

		chart.Title = fmt.Sprintf("CPU: %d ", int(metrics.Load)) + "%"
		ui.Render(chart)
	})

	eb.Subscribe("jvm-selected", func(pid string) {
		chart.Data[0] = []float64{0, 0}
		chart.Title = fmt.Sprintf("CPU %")
		ui.Render(chart)
	})

	return chart
}
