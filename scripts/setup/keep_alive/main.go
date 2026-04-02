package main

import (
	"log"
	"os"
	"os/exec"
	"time"
)

func main() {
	for {
		cmd := exec.Command("./tmp/fleet-backend")
		
		// Set environment explicitly
		cmd.Env = []string{
			"PORT=8081",
			"GRPC_PORT=50051",
			"DATABASE_URL=host=/data/data/com.termux/files/usr/tmp user=u0_a412 dbname=fleet_db sslmode=disable",
			"NATS_URL=nats://localhost:4222",
			"REDIS_ADDR=localhost:6380",
			"INVOICE_STORAGE_PATH=./tmp/invoices",
			"HOME=" + os.Getenv("HOME"),
			"PATH=" + os.Getenv("PATH"),
			"USER=" + os.Getenv("USER"),
		}

		// Redirect output to files
		logFile, _ := os.OpenFile("tmp/logs/backend_persist.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		cmd.Stdout = logFile
		cmd.Stderr = logFile

		log.Println("Starting fleet-backend with explicit ENV...")
		err := cmd.Run()
		if err != nil {
			log.Printf("Backend exited with error: %v. Restarting in 2s...", err)
		} else {
			log.Println("Backend exited cleanly. Restarting in 2s...")
		}
		logFile.Close()
		time.Sleep(2 * time.Second)
	}
}
