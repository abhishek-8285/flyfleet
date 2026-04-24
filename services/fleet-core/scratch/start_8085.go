package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// Build
	fmt.Println("Building...")
	buildCmd := exec.Command("go", "build", "-o", "fleet-core-8085", "./cmd/main.go")
	buildCmd.Dir = "/home/abhishek/Desktop/flyfleet/services/fleet-core"
	if out, err := buildCmd.CombinedOutput(); err != nil {
		fmt.Printf("Build failed: %s\n", out)
		return
	}

	// Start
	fmt.Println("Starting on 8085...")
	startCmd := exec.Command("./fleet-core-8085")
	startCmd.Dir = "/home/abhishek/Desktop/flyfleet/services/fleet-core"
	logFile, _ := os.OpenFile("/home/abhishek/Desktop/flyfleet/services/fleet-core/fleet-8085.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	startCmd.Stdout = logFile
	startCmd.Stderr = logFile
	if err := startCmd.Start(); err != nil {
		fmt.Printf("Start failed: %v\n", err)
	} else {
		fmt.Printf("Started new process on 8085, PID: %d\n", startCmd.Process.Pid)
	}
}
