package jvmmon

import (
	"image"
	"testing"

	"github.com/asaskevich/EventBus"
	ui "github.com/gizak/termui/v3"
)

func TestCpuChartValuesIncludesGcPercent(t *testing.T) {
	got := cpuChartValues(Metrics{Load: 37.5, GcUsage: 425})
	want := [2]float64{37.5, 42.5}
	if got != want {
		t.Fatalf("cpuChartValues() = %v, want %v", got, want)
	}
}

func TestCpuChartUpdateAndReset(t *testing.T) {
	chart := NewCpuChart(EventBus.New())
	chart.SetRect(0, 0, 40, 10)

	chart.Update(Metrics{Load: 10, GcUsage: 50})
	if got := chart.Data[0][len(chart.Data[0])-1]; got != 10 {
		t.Fatalf("last CPU value = %v, want 10", got)
	}
	if got := chart.Data[1][len(chart.Data[1])-1]; got != 5 {
		t.Fatalf("last GC value = %v, want 5", got)
	}

	chart.Reset()
	if len(chart.Data) != 2 || len(chart.Data[0]) != 2 || len(chart.Data[1]) != 2 {
		t.Fatalf("Reset() data = %v, want two series of two zero points", chart.Data)
	}
	if chart.hasData {
		t.Fatal("Reset() should clear hasData")
	}
}

func TestCpuChartUpdateTrimsToWidth(t *testing.T) {
	chart := NewCpuChart(EventBus.New())
	chart.SetRect(0, 0, 20, 10)
	for i := 0; i < 100; i++ {
		chart.Update(Metrics{Load: float64(i)})
	}
	if got, max := len(chart.Data[0]), 20/2; got > max {
		t.Fatalf("CPU series length = %d, want <= %d", got, max)
	}
	if len(chart.Data[0]) != len(chart.Data[1]) {
		t.Fatalf("series lengths differ: %d vs %d", len(chart.Data[0]), len(chart.Data[1]))
	}
}

func TestCpuChartDrawsColoredTitle(t *testing.T) {
	chart := NewCpuChart(EventBus.New())
	chart.SetRect(0, 0, 40, 10)
	chart.Update(Metrics{Load: 37.5, GcUsage: 12})

	buf := ui.NewBuffer(chart.GetRect())
	chart.Draw(buf)

	var title string
	for x := 2; x < 38; x++ {
		title += string(buf.GetCell(image.Pt(x, 0)).Rune)
	}
	if want := "CPU: 37.5%, GC: 1.2%"; title[:len(want)] != want {
		t.Fatalf("title = %q, want prefix %q", title, want)
	}

	if got := buf.GetCell(image.Pt(2, 0)).Style.Fg; got != cpuColor {
		t.Errorf("CPU title color = %v, want %v", got, cpuColor)
	}
	gcStart := 2 + len("CPU: 37.5%, ")
	if got := buf.GetCell(image.Pt(gcStart, 0)).Style.Fg; got != gcColor {
		t.Errorf("GC title color = %v, want %v", got, gcColor)
	}
}

func TestCpuChartTitleClippedToWidth(t *testing.T) {
	chart := NewCpuChart(EventBus.New())
	chart.SetRect(0, 0, 12, 6)
	chart.Update(Metrics{Load: 37.5, GcUsage: 12})

	buf := ui.NewBuffer(chart.GetRect())
	chart.Draw(buf)

	// the right border column must not be overwritten by the title
	if got := buf.GetCell(image.Pt(11, 0)).Rune; got != ui.TOP_RIGHT {
		t.Fatalf("top-right border = %q, want %q", got, ui.TOP_RIGHT)
	}
}
