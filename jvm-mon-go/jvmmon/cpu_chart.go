package jvmmon

import (
	"fmt"
	"image"

	"github.com/asaskevich/EventBus"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

const (
	cpuColor = ui.ColorYellow
	gcColor  = ui.ColorRed
)

// CpuChart is a line chart of CPU and GC load. Its title is colored to match
// the plotted lines, so it doubles as the legend.
type CpuChart struct {
	*widgets.Plot
	cpu, gc float64
	hasData bool
}

type titleSegment struct {
	text  string
	style ui.Style
}

// cpuChartValues returns CPU and GC load as percentages. The agent sends GC
// usage in tenths of a percent.
func cpuChartValues(metrics Metrics) [2]float64 {
	return [2]float64{metrics.Load, metrics.GcUsage / 10}
}

func NewCpuChart(eb EventBus.Bus) *CpuChart {
	chart := &CpuChart{Plot: widgets.NewPlot()}
	chart.AxesColor = ui.ColorWhite
	chart.PlotType = widgets.LineChart
	chart.LineColors = []ui.Color{cpuColor, gcColor}
	chart.Reset()

	eb.Subscribe("metrics", func(metrics Metrics) {
		uiMu.Lock()
		defer uiMu.Unlock()
		chart.Update(metrics)
		ui.Render(chart)
	})

	eb.Subscribe("jvm-selected", func(pid string) {
		uiMu.Lock()
		defer uiMu.Unlock()
		chart.Reset()
		ui.Render(chart)
	})

	return chart
}

// Update appends a metrics sample to the chart. It does not render.
func (c *CpuChart) Update(metrics Metrics) {
	maxX := c.Bounds().Max.X
	values := cpuChartValues(metrics)
	for i, value := range values {
		data := append(c.Data[i], value)
		if len(data) > maxX/2 {
			data = data[1:]
		}
		c.Data[i] = data
	}
	c.cpu, c.gc = values[0], values[1]
	c.hasData = true
}

// Reset clears the chart back to its initial empty state. It does not render.
func (c *CpuChart) Reset() {
	c.Data = [][]float64{{0, 0}, {0, 0}}
	c.cpu, c.gc = 0, 0
	c.hasData = false
}

func (c *CpuChart) titleSegments() []titleSegment {
	cpuStyle := ui.NewStyle(cpuColor)
	gcStyle := ui.NewStyle(gcColor)
	sepStyle := ui.NewStyle(ui.ColorWhite)

	if !c.hasData {
		return []titleSegment{
			{"CPU %", cpuStyle},
			{" / ", sepStyle},
			{"GC %", gcStyle},
		}
	}
	return []titleSegment{
		{fmt.Sprintf("CPU: %.1f%%", c.cpu), cpuStyle},
		{", ", sepStyle},
		{fmt.Sprintf("GC: %.1f%%", c.gc), gcStyle},
	}
}

// Draw implements ui.Drawable. The embedded Plot draws the border and lines
// with an empty title, then the colored title segments are drawn on top.
func (c *CpuChart) Draw(buf *ui.Buffer) {
	c.Plot.Draw(buf)

	x := c.Min.X + 2 // same offset as ui.Block titles
	limit := c.Max.X - 2
	for _, seg := range c.titleSegments() {
		for _, r := range seg.text {
			if x >= limit {
				return
			}
			buf.SetCell(ui.Cell{Rune: r, Style: seg.style}, image.Pt(x, c.Min.Y))
			x++
		}
	}
}
