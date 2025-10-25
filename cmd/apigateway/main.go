package main

import (
	"os"
	"os/signal"
	"syscall"
	"vibe-drop/internal/apigateway"
)

func main() {
	go apigateway.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	apigateway.Stop()
}
