package jvmmon

type Thread struct {
	Id      int64
	Name    string
	State   string
	CpuTime int64
}

type Threads struct {
	Count   int
	Threads []Thread
}

type Metrics struct {
	Used    float64
	Max     float64
	Load    float64
	GcUsage float64
	Threads Threads
}
