package jvmmon

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	ui "github.com/gizak/termui/v3" // <- ui shortcut, optional
	"github.com/gizak/termui/v3/widgets"
	"sort"
	"strconv"
)

func NewNavTable(data map[string]JVM, borderLabel string, rowCount int, eb EventBus.Bus) *widgets.Table {
	labels := []string{"PID", "Ver.", "User", "Main"}
	rows := [][]string{labels}
	pids := make([]string, 0, len(data))
	for pid := range data {
		pids = append(pids, pid)
	}
	sort.Slice(pids, func(i, k int) bool {
		a, errA := strconv.Atoi(pids[i])
		b, errB := strconv.Atoi(pids[k])
		if errA != nil || errB != nil {
			return pids[i] < pids[k]
		}
		return a < b
	})
	for _, pid := range pids {
		jvm := data[pid]
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
		uiMu.Lock()
		defer uiMu.Unlock()
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
			go eb.Publish("jvm-selected", pid) // outside UI lock: handlers take it too
		}
	}, false)

	eb.SubscribeAsync("attach-error", func(pid string) {
		uiMu.Lock()
		defer uiMu.Unlock()
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
		uiMu.Lock()
		defer uiMu.Unlock()
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
		uiMu.Lock()
		defer uiMu.Unlock()
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
		uiMu.Lock()
		defer uiMu.Unlock()
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
		uiMu.Lock()
		defer uiMu.Unlock()
		chart.Data = []float64{}
		slg.Title = "Memory"
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
		uiMu.Lock()
		defer uiMu.Unlock()
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
		uiMu.Lock()
		defer uiMu.Unlock()
		chart.Data[0] = []float64{0, 0}
		chart.Title = "CPU %"
		ui.Render(chart)
	})

	return chart
}
