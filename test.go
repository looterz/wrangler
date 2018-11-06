package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

var programQuit chan bool
var processMonitorQuit chan bool
var processPid int

func isProcessRunning() bool {
	process, err := os.FindProcess(processPid)
	if err != nil {
		fmt.Println(err)
	}

	return process != nil
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
				log.Println("Proccess is not running")
			}
		}

		time.Sleep(time.Second)
	}
}

func startProcess() int {
	cmd := exec.Command("calc.exe")
	if err := cmd.Start(); err != nil {
		log.Panic(err)
	}

	return cmd.Process.Pid
}

func main() {
	programQuit = make(chan bool)
	processMonitorQuit = make(chan bool)
	processPid = startProcess()

	go processMonitor()

	<-programQuit
}
