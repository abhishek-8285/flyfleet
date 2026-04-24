package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	// Find all processes named fleet-core or main (if built by air/go run)
	procDirs, _ := ioutil.ReadDir("/proc")
	for _, dir := range procDirs {
		if !dir.IsDir() {
			continue
		}
		pid := dir.Name()
		cmdlinePath := filepath.Join("/proc", pid, "cmdline")
		content, err := ioutil.ReadFile(cmdlinePath)
		if err != nil {
			continue
		}
		cmdline := string(content)
		if strings.Contains(cmdline, "fleet-core") || strings.Contains(cmdline, "main") {
			if pid == fmt.Sprintf("%d", os.Getpid()) {
				continue
			}
			fmt.Printf("Found process: %s (PID: %s). Killing...\n", cmdline, pid)
			exec.Command("kill", "-9", pid).Run()
		}
	}

	// Build
	fmt.Println("Building...")
	buildCmd := exec.Command("go", "build", "-o", "fleet-core-new", "./cmd/main.go")
	buildCmd.Dir = "/home/abhishek/Desktop/flyfleet/services/fleet-core"
	if out, err := buildCmd.CombinedOutput(); err != nil {
		fmt.Printf("Build failed: %s\n", out)
		return
	}

	// Start
	fmt.Println("Starting...")
	startCmd := exec.Command("./fleet-core-new")
	startCmd.Dir = "/home/abhishek/Desktop/flyfleet/services/fleet-core"
	logFile, _ := os.OpenFile("/home/abhishek/Desktop/flyfleet/services/fleet-core/fleet-manual-new.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	startCmd.Stdout = logFile
	startCmd.Stderr = logFile
	if err := startCmd.Start(); err != nil {
		fmt.Printf("Start failed: %v\n", err)
	} else {
		fmt.Printf("Started new process PID: %d\n", startCmd.Process.Pid)
	}
}
