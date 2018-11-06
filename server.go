package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

func startServer(level string, name string, maxPlayers string) (process *os.Process) {
	cmd := exec.Command(serverBin, fmt.Sprintf("%s?ServerName=%s?MaxPlayers=%s", level, name, maxPlayers), "-log")
	if err := cmd.Start(); err != nil {
		log.Panic(err)
	}

	log.Println("Server started with pid", cmd.Process.Pid)

	return cmd.Process
}

func updateServer(serverBranch string) {
	args := []string{"+login anonymous", "+app_update 412680"}
	if serverBranch != "" {
		args = append(args, fmt.Sprintf("-beta %s", serverBranch))
	}
	args = append(args, "validate")
	args = append(args, "+quit")

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
