package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// Try to find the process on 8084
	cmd := exec.Command("lsof", "-t", "-i:8084")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error finding process: %v\n", err)
	} else {
		pids := strings.Fields(string(output))
		for _, pid := range pids {
			fmt.Printf("Killing PID: %s\n", pid)
			killCmd := exec.Command("kill", "-9", pid)
			killCmd.Run()
		}
	}

	// Try to build and start
	fmt.Println("Building backend...")
	buildCmd := exec.Command("go", "build", "-o", "fleet-core", "./cmd/main.go")
	buildCmd.Dir = "/home/abhishek/Desktop/flyfleet/services/fleet-core"
	if out, err := buildCmd.CombinedOutput(); err != nil {
		fmt.Printf("Build error: %v\nOutput: %s\n", err, out)
		return
	}

	fmt.Println("Starting backend...")
	// We use a separate process to avoid being killed when this script exits
	startCmd := exec.Command("./fleet-core")
	startCmd.Dir = "/home/abhishek/Desktop/flyfleet/services/fleet-core"
	
	logFile, _ := os.OpenFile("/home/abhishek/Desktop/flyfleet/services/fleet-core/fleet-manual.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	startCmd.Stdout = logFile
	startCmd.Stderr = logFile
	
	if err := startCmd.Start(); err != nil {
		fmt.Printf("Start error: %v\n", err)
		return
	}
	
	fmt.Printf("Backend started with PID: %d\n", startCmd.Process.Pid)
}
