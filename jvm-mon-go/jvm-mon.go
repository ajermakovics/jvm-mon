package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	. "github.com/ajermakovics/jvm-mon-go/jvmmon"
	"github.com/asaskevich/EventBus"
	ui "github.com/gizak/termui/v3"

	_ "embed"
)

var jar, port string
var jvms map[string]JVM
var server *Server
var version = "1.3"
var eb EventBus.Bus

//go:embed build/libs/jvm-mon-go.jar
var jarBytes []byte

// openLog opens the log in the user's private cache dir. If unavailable, it
// falls back to a uniquely named temp file (CreateTemp never opens an
// existing file, so pre-planted symlinks are not followed).
func openLog() (*os.File, error) {
	if dir, err := os.UserCacheDir(); err == nil {
		dir = filepath.Join(dir, "jvm-mon")
		if os.MkdirAll(dir, 0700) == nil {
			path := filepath.Join(dir, "jvm-mon.log")
			if f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600); err == nil {
				return f, nil
			}
		}
	}
	return os.CreateTemp("", "jvm-mon-*.log")
}

func setup() {
	showVersion := flag.Bool("v", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println("jvm-mon v:", version)
		os.Exit(0)
	}

	user := GetCurUser()
	logFile, err := openLog()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening log file: %v\n", err)
		os.Exit(1)
	}
	logPath := logFile.Name()
	log.SetOutput(logFile)
	log.Println("jvm-mon v:", version, "user:", user)
	println("jvm-mon v:", version, " user:", user, " log:", logPath)

	jvms = GetJVMs()
	log.Println("Found JVMs: ", len(jvms))

	eb = EventBus.New()

	var serverErr error
	server, serverErr = NewServer(eb)
	if serverErr != nil {
		fmt.Fprintln(os.Stderr, "Cannot start server:", serverErr)
		os.Exit(1)
	}
	port = strconv.Itoa(server.Port)

	go receiveMetrics()
	go checkConnections()
}

func main() {
	setup()

	var err error
	jar, err = loadJar()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot extract agent jar:", err)
		os.Exit(1)
	}
	defer os.Remove(jar)

	if err = ui.Init(); err != nil {
		_ = os.Remove(jar) // os.Exit skips deferred calls
		fmt.Fprintln(os.Stderr, "Cannot initialize UI:", err)
		os.Exit(1)
	}
	defer ui.Close()

	// Create UI
	jvmTable := NewNavTable(jvms, "JVMs (v"+version+")", 9, eb)
	memChart := NewMemChart(eb)
	cpuChart := NewCpuChart(eb)
	threadTable := NewThreadTable(14, eb)

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	half := 1.0 / 2
	grid.Set(
		ui.NewRow(half,
			ui.NewCol(half, jvmTable),
			ui.NewCol(half, cpuChart)),
		ui.NewRow(half,
			ui.NewCol(half, threadTable),
			ui.NewCol(half, memChart)))

	Render(grid)

	eb.SubscribeAsync("jvm-selected", monitor, false)

	uiEvents := ui.PollEvents()
	for {
		select {
		case e := <-uiEvents:
			if e.Type == ui.KeyboardEvent {
				eb.Publish("keyboard-events", e.ID)
			}
			switch e.ID {
			case "q", "<C-c>", "<Escape>": // exit (deferred cleanup runs)
				logErr(server.Close())
				return
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				WithUI(func() {
					grid.SetRect(0, 0, payload.Width, payload.Height)
					ui.Clear()
					ui.Render(grid)
				})
			}
		}
	}
}

func logErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

func loadJar() (string, error) {
	log.Println("Found embedded jar file: ", len(jarBytes))
	tmpJarFile, err := os.CreateTemp("", "jvm-mon-go-*.jar")
	if err != nil {
		return "", err
	}
	tmpJarPath := tmpJarFile.Name()
	if _, err = tmpJarFile.Write(jarBytes); err != nil {
		_ = tmpJarFile.Close()
		_ = os.Remove(tmpJarPath)
		return "", err
	}
	if err = tmpJarFile.Close(); err != nil {
		_ = os.Remove(tmpJarPath)
		return "", err
	}
	log.Println("Created temp file ", tmpJarPath)

	// target JVM may run as another user (root mode) and must be able to read the jar
	if err = os.Chmod(tmpJarPath, 0644); err != nil {
		log.Println("Cannot chmod ", tmpJarPath, " ", err)
	}

	return tmpJarPath, nil
}

func checkConnections() {
	for {
		addr := <-server.Connections
		log.Println("JVM Connected ", addr)
	}
}

func receiveMetrics() {
	for {
		msg := <-server.Messages
		var metrics Metrics
		if err := json.Unmarshal([]byte(msg), &metrics); err != nil {
			log.Println("Cannot unmarshal:", msg, "err:", err)
			continue
		}

		eb.Publish("metrics", metrics)
		eb.Publish("metrics.Threads", metrics.Threads)
	}
}

func monitor(pid string) {
	log.Println("Monitoring pid: ", pid)
	jvm, ok := jvms[pid]
	if !ok {
		log.Println("Unknown pid:", pid)
		return
	}
	go attachAgent(jvm, jar, port)
}

func attachAgent(jvm JVM, jar string, port string) {
	err := jvm.AttachAndLoadAgent(jar, port)
	if err != nil {
		log.Println("Cannot attach to pid ", jvm.Pid, err)
		eb.Publish("attach-error", jvm.Pid)
	}
}
