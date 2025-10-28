package main

import (
	"os"
	"os/signal"
	"syscall"
	"vibe-drop/internal/fileservice"
)

func main() {
	go fileservice.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fileservice.Stop()
}
