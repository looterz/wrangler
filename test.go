package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var programQuit chan bool
var processMonitorQuit chan bool
var processPid int

func isProcessRunning() bool {
	cmd := exec.Command("tasklist")
	output, _ := cmd.Output()
	text := string(output)

	if len(text) > 0 {
		return strings.Contains(text, strconv.Itoa(processPid))
	}

	return false
}

func processMonitor() {
	for {
		select {
		case <-processMonitorQuit:
			return
		default:
			if isProcessRunning() {
				log.Println("Process is running", processPid)
			} else {
				log.Println("Proccess crash detected for", processPid)

				time.Sleep(time.Second * 5)

				processPid = startProcess()
				go processMonitor()
				return
			}
		}

		time.Sleep(time.Second)
	}
}

func startProcess() int {
	cmd := exec.Command("notepad.exe")
	if err := cmd.Start(); err != nil {
		log.Panic(err)
	}

	return cmd.Process.Pid
}

func simulateRestart() {
	time.Sleep(time.Second * 5)

	processMonitorQuit <- true

	fmt.Println("Restarting process")

	proc, err := os.FindProcess(processPid)
	if err != nil {
		fmt.Println(err)
	}

	err = proc.Kill()
	if err != nil {
		fmt.Println(err)
	}

	time.Sleep(time.Second * 5)

	processPid = startProcess()

	go processMonitor()
}

func main() {
	programQuit = make(chan bool)
	processMonitorQuit = make(chan bool)
	processPid = startProcess()

	go processMonitor()
	go simulateRestart()

	<-programQuit
}
