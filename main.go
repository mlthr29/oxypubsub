package main

import (
	"log"
	"os"
	"os/signal"
	broker "oxypubsub/src"
	"syscall"
)

func main() {
	b := broker.CreateBroker()

	server := broker.CreateTCPServer(":8080", ":8081", b)

	log.Println("Server starting...")

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Server shutting down...")

	server.Stop()
}
