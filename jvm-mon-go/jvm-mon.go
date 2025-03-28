package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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

func init() {
	if len(os.Args) > 1 && os.Args[1] == "-v" {
		println("jvm-mon v:", version)
		os.Exit(0)
	}

	user := GetCurUser()
	var logErr error
	logPath := os.TempDir() + string(os.PathSeparator) + "jvm-mon_" + user + ".log"
	logFile, logErr := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_TRUNC, 0666)
	if logErr != nil {
		log.Fatalf("Error opening %v file: %v", logPath, logErr)
		panic(logErr)
	}
	log.SetOutput(logFile)
	log.Println("jvm-mon v:", version, "user:", user)
	println("jvm-mon v:", version, " user:", user, " log:", logPath)

	jvms = GetJVMs()
	log.Println("Found JVMs: ", len(jvms))

	eb = EventBus.New()

	var serverErr error
	server, serverErr = NewServer(eb)
	if serverErr != nil {
		panic(serverErr)
	}
	port = strconv.Itoa((*server).Port)

	go receiveMetrics()
	go checkConnections()
}

func main() {
	jar = loadJar()

	err := ui.Init()
	if err != nil {
		log.Fatal("Cannot initialize UI", err)
		panic(err)
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

	ui.Render(grid)

	eb.SubscribeAsync("jvm-selected", monitor, false)

	uiEvents := ui.PollEvents()
	for {
		select {
		case e := <-uiEvents:
			if e.Type == ui.KeyboardEvent {
				eb.Publish("keyboard-events", e.ID)
			}
			switch e.ID {
			case "q", "<C-c>", "<Escape>": // exit
				cleanUp()
				return
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				grid.SetRect(0, 0, payload.Width, payload.Height)
				ui.Clear()
				ui.Render(grid)
			}
		}
	}
}

func cleanUp() {
	os.Remove(jar) // from temp
	ui.Close()
}

func loadJar() string {
	log.Println("Found embedded jar file: ", len(jarBytes))
	tmpJarFile, err := os.CreateTemp(os.TempDir(), "jvm-mon-go.jar")
	if _, err = tmpJarFile.Write(jarBytes); err != nil {
		fmt.Println("Failed to write to temporary file", err)
	}
	var tmpJarPath = tmpJarFile.Name()
	log.Println("Created temp file ", tmpJarPath)

	if err := tmpJarFile.Close(); err != nil {
		fmt.Println(err)
	}

	err = os.Chmod(tmpJarPath, 0644)
	if err != nil {
		log.Println("Cannot chmod ", tmpJarPath, " ", err)
	}

	return tmpJarPath
}

func checkConnections() {
	for {
		addr := <-(*server).Connections
		log.Println("JVM Connected ", addr)
	}
}

func receiveMetrics() {
	for {
		msg := <-(*server).Messages
		var metrics Metrics
		msgBytes := []byte(msg)
		err := json.Unmarshal(msgBytes, &metrics)
		if err != nil {
			log.Fatal("Cannot unmarshal: ", msg, "err: ", err)
			continue
		}

		eb.Publish("metrics", metrics)
		eb.Publish("metrics.Threads", metrics.Threads)
	}
}

func monitor(pid string) {
	log.Println("Monitoring pid: ", pid)
	jvm := jvms[pid]
	go attachAgent(jvm, jar, port)
}

func attachAgent(jvm JVM, jar string, port string) {
	err := jvm.AttachAndLoadAgent(jar, port)
	if err != nil {
		log.Println("Cannot attach to pid ", jvm.Pid)
		eb.Publish("attach-error", jvm.Pid)
	}
}
