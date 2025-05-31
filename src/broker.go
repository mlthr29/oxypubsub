package broker

import (
	"fmt"
	"log"
	"net"
)

type Broker struct {
	topics []string
}

func CreateBroker() *Broker {
	return &Broker{
		topics: make([]string, 0),
	}
}

func (broker *Broker) Start(port string) {
	listener, error := net.Listen("tcp", ":"+port)

	if error != nil {
		log.Fatal("Failed to start broker:", error)
	}

	fmt.Println("listner started")

	defer listener.Close()

	for {
		connection, error := listener.Accept()
		if error != nil {
			log.Println("Failed to accept connection:", error)
			continue
		}

		go broker.establishConnection(connection)
	}
}

func (*Broker) establishConnection(connection net.Conn) {
	fmt.Println("Connection established")
}
