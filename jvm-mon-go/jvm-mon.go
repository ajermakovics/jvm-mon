package main

import (
	"encoding/json"
	"fmt"
	"github.com/GeertJohan/go.rice"
	"github.com/GeertJohan/go.rice/embedded"
	. "github.com/ajermakovics/jvm-mon-go/jvmmon"
	ui "github.com/gizak/termui"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

var pid, jar, port string
var jvms map[string]JVM
var server *Server
var version = "1.0-ea1"
var logFile *os.File

func init() {
	user := GetCurUser()
	var logErr error
	logPath := os.TempDir() + string(os.PathSeparator) + "jvm-mon_" + user + ".log"
	logFile, logErr := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_TRUNC, 0666)
	if logErr != nil {
		log.Fatalf("Error opening %v file: %v", logPath, logErr)
		panic(logErr)
	}
	log.SetOutput(logFile)
	println("jvm-mon v:", version, " user:", user, " log:", logPath)

	jvms = GetJVMs()
	log.Println("jvm-mon v", version)
	log.Println("Found JVMs: ", len(jvms))

	err := ui.Init()
	if err != nil {
		log.Fatal("Cannot initialize UI", err)
		panic(err)
	}

	server, err = NewServer()
	if err != nil {
		panic(err)
	}
	port = strconv.Itoa((*server).Port)
	go receiveMetrics()
	go checkConnections()
}

func main() {
	jar = loadJar()

	// Create UI
	jvmTable := NewNavTable(jvms, "JVMs (v" + version + ")", 9)
	memChart := NewMemChart()
	cpuChart := NewCpuChart()
	threadTable := NewThreadTable(9)

	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(6, 0, jvmTable),
			ui.NewCol(6, 0, memChart)),
		ui.NewRow(
			ui.NewCol(6, 0, threadTable),
			ui.NewCol(6, 0, cpuChart)))

	ui.Body.Align()
	ui.Render(ui.Body)

	ui.Handle("/sys/kbd/C-c", func(ui.Event) { ui.StopLoop() })
	ui.Handle("/sys/kbd/q", func(ui.Event) { ui.StopLoop() })
	ui.Handle("/nav-table/selected", monitor)

	ui.Loop()
	ui.Close()
	cleanUp()
}

func cleanUp() {
	os.Remove(jar) // from temp
}

func loadJar() string {
	for boxName, _ := range embedded.EmbeddedBoxes {
		log.Println("Embedded dir: ", boxName)
	}

	box := rice.MustFindBox(`build/libs`)
	jarFile, err := box.Open(`jvm-mon-go.jar`)
	if err != nil {
		panic(err)
	}
	stat, _ := jarFile.Stat()
	log.Println("Found embeded jar file: ", stat.Size())
	jarBytes, err := box.Bytes("jvm-mon-go.jar")
	if err != nil {
		panic(err)
	}
	tmpJarFile, err := ioutil.TempFile(os.TempDir(), "jvm-mon-go.jar")
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

		ui.SendCustomEvt("/metrics/mem", metrics)
		ui.SendCustomEvt("/metrics/cpu", metrics)
		ui.SendCustomEvt("/metrics/threads", metrics.Threads)
	}
}

func monitor(e ui.Event) {
	pid = e.Data.(string)
	jvm := jvms[pid]
	ui.SendCustomEvt("/metrics/mem/clear", pid)
	ui.SendCustomEvt("/metrics/cpu/clear", pid)
	go attachAgent(jvm, jar, port)
}

func attachAgent(jvm JVM, jar string, port string) {
	err := jvm.AttachAndLoadAgent(jar, port)
	if err != nil {
		log.Println("Cannot attach to pid ", pid)
	}
}

func findJar() string {
	//workDir, _ := os.Getwd()
	self, _ := os.Executable()
	selfDir := filepath.Dir(self)

	jar = filepath.Join(selfDir, "libs", "jvm-mon-go.jar") // during dev
	if _, err := os.Stat(jar); os.IsNotExist(err) {
		jar = filepath.Join(selfDir, "jvm-mon-go.jar")
	}
	if _, err := os.Stat(jar); os.IsNotExist(err) {
		log.Fatal("Agent jar not found: ", jar, "Error: ", err)
		panic(err)
	}
	log.Println("Agent jar file: ", jar)
	return jar
}
