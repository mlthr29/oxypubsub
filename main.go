package main

import (
	"os"
	"os/signal"
	broker "oxypubsub/src"
	"syscall"
)

func main() {
	b := broker.CreateBroker()

	server := broker.CreateTCPServer(":8080", b)

	go server.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	server.Stop()
}
