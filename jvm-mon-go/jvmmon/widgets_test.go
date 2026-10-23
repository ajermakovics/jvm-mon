package jvmmon

import "testing"

func TestCpuChartValuesIncludesGcPercent(t *testing.T) {
	got := cpuChartValues(Metrics{Load: 37.5, GcUsage: 425})
	want := [2]float64{37.5, 42.5}
	if got != want {
		t.Fatalf("cpuChartValues() = %v, want %v", got, want)
	}
}
