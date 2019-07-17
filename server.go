package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var programQuit chan bool
var processMonitorQuit chan bool

func processMonitor() {
	for {
		select {
		case <-processMonitorQuit:
			return
		default:
			if !isProcessRunning() {
				log.Println("Proccess crash detected, restarting")

				time.Sleep(time.Second * 60)

				if err := killServer(); err != nil {
					log.Println(err)
				}

				time.Sleep(time.Second * 10)

				startServer()

				go processMonitor()
				return
			}
		}

		time.Sleep(time.Second)
	}
}

func isProcessRunning() bool {
	cmd := exec.Command("tasklist")
	output, _ := cmd.Output()
	text := string(output)

	if len(text) > 0 {
		return strings.Contains(text, processName)
	}

	return false
}

func killServer() error {
	kill := exec.Command("TASKKILL", "/T", "/F", "/PID", strconv.Itoa(serverProcess.Pid))
	kill.Stderr = os.Stderr
	kill.Stdout = os.Stdout
	return kill.Run()
}

func startServer() {
	cmd := exec.Command(serverBin, fmt.Sprintf("%s?ServerName=%s?MaxPlayers=%s?Game=%s", serverMap, serverName, serverMaxPlayers, serverGame), fmt.Sprintf("-SteamServerName=%s", serverName), "-log")
	if err := cmd.Start(); err != nil {
		log.Panic(err)
	}

	log.Println("Server started with pid", cmd.Process.Pid)

	serverProcess = cmd.Process
}

func updateServerConfig(IniURI string) {
	log.Println("Downloading game ini from", IniURI)

	adminFilePath := fmt.Sprintf("%s\\TheIsle\\Saved\\Config\\WindowsServer\\Game.ini", Config.Server)

	if err := DownloadFile(adminFilePath, IniURI); err != nil {
		log.Println(err)
	} else {
		log.Println("Updated server admins file")
	}
}

func updateServer(IniURI string) {
	if IniURI != "" {
		updateServerConfig(IniURI)
	} else {
		updateServerConfig("https://s3-us-west-2.amazonaws.com/isle-static/Game.ini")
	}

	if Config.UseS3Bucket {
		updateServerS3()
		return
	}

	args := []string{"+login", "anonymous", "+app_update", "412680"}
	if serverBranch != "" && serverBranch != "live" {
		args = append(args, "-beta")
		args = append(args, serverBranch)
	}
	args = append(args, "validate")
	args = append(args, "+quit")

	log.Println("Updating server with args:\n", steamcmdBin, args)

	cmd := exec.Command(steamcmdBin, args...)

	var stdBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdBuffer)

	cmd.Stdout = mw
	cmd.Stderr = mw

	if err := cmd.Run(); err != nil {
		log.Panic(err)
	}

	log.Println(stdBuffer.String())
}

// DownloadFile downloads a file using http
func DownloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
