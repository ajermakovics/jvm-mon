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
	labels := []string{"PID", "Ver.", "User", "Main"}
	rows := [][]string{labels}
	for pid, jvm := range data {
		rows = append(rows, []string{pid, jvm.Version, jvm.User, jvm.ProcName})
	}

	table := widgets.NewTable()
	table.Rows = rows
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.Border = true
	table.Title = borderLabel
	table.TextAlignment = ui.AlignLeft
	table.ColumnWidths = []int{6, 5, 10, -1}
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

	eb.SubscribeAsync("attach-error", func(pid string) {
		rowIndex := findIndex(rows, pid)
		if rowIndex > -1 {
			table.RowStyles[rowIndex] = ui.NewStyle(ui.ColorRed)
		}
		ui.Render(table)
	}, false)

	return table
}

func findIndex(matrix [][]string, needle string) int {
	for i, row := range matrix {
		if len(row) > 0 && row[0] == needle {
			return i
		}
	}
	return -1
}

func NewThreadTable(rowCount int, eb EventBus.Bus) *widgets.Table {
	rows := [][]string{threadTableLabels()}

	table := widgets.NewTable()
	table.Rows = rows
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.Border = true
	table.Title = "Threads"
	table.RowSeparator = false
	table.ColumnWidths = []int{6, 15, 10, -1}

	eb.Subscribe("metrics.Threads", func(threads Threads) {
		threadArr := threads.Threads
		table.Title = "Threads (" + strconv.Itoa(threads.Count) + ")"

		rows := [][]string{threadTableLabels()}
		for idx, thread := range threadArr {
			rows = append(rows, thread.toRow())
			if idx == rowCount-1 {
				break
			}
		}
		table.Rows = rows
		ui.Render(table)
	})

	eb.Subscribe("jvm-selected", func(pid string) { // clear
		table.Rows = [][]string{threadTableLabels()}
		table.Title = "Threads"
		ui.Render(table)
	})

	return table
}

func threadTableLabels() []string {
	return []string{"Id", "State", "CpuTime", "Name"}
}

func (t *Thread) toRow() []string {
	cpuTime := strconv.FormatInt(t.CpuTime/1000, 10)
	if t.CpuTime == 0 {
		cpuTime = ""
	}
	return []string{strconv.FormatInt(t.Id, 10), t.State, cpuTime, t.Name}
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
