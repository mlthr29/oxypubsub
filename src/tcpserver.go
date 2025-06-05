package broker

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

type TCPServer struct {
	subAddress  string
	pubAddress  string
	broker      *Broker
	subListener net.Listener
	pubListener net.Listener
	wg          sync.WaitGroup
	quit        chan struct{}
}

func CreateTCPServer(subAddress, pubAddress string, broker *Broker) *TCPServer {
	return &TCPServer{
		subAddress: subAddress,
		pubAddress: pubAddress,
		broker:     broker,
		quit:       make(chan struct{}),
	}
}

func (s *TCPServer) Start() error {
	var err error

	s.subListener, err = net.Listen("tcp", s.subAddress)
	if err != nil {
		return fmt.Errorf("Failed to start the Sub listener: %v", err)
	}

	s.pubListener, err = net.Listen("tcp", s.pubAddress)
	if err != nil {
		return fmt.Errorf("Failed to start the Pub listener: %v", err)
	}

	s.wg.Add(2)

	go s.accept(s.subListener, s.handleSubscribers)
	go s.accept(s.pubListener, s.handlePublishers)

	return nil
}

func (s *TCPServer) Stop() {
	close(s.quit)
	if s.subListener != nil {
		s.subListener.Close()
	}

	if s.pubListener != nil {
		s.pubListener.Close()
	}

	s.wg.Wait()
}

func (s *TCPServer) accept(listener net.Listener, handler func(net.Conn)) {
	defer s.wg.Done()

	for {
		conn, err := listener.Accept()

		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				fmt.Errorf("Error accepting connection: %w", err)
				continue
			}
		}

		s.wg.Add(1)

		go func() {
			defer s.wg.Done()
			handler(conn)
		}()
	}
}

func (s *TCPServer) handleSubscribers(conn net.Conn) {
	defer conn.Close()

	msgCh := s.broker.Subscribe()
	defer s.broker.Unsubscribe(msgCh)

	fmt.Fprintf(conn, "WELCOME: Connected to Oxy Pub/Sub Subscriber Server\n")

	disconnected := make(chan struct{})

	go func() {
		buf := make([]byte, 1)
		for {
			_, err := conn.Read(buf)
			if err != nil {
				close(disconnected)
				return
			}
		}
	}()

	for {
		select {
		case msg, ok := <-msgCh:
			if !ok {
				return
			}
			if _, err := fmt.Fprintf(conn, "MESSAGE: %s\n", msg); err != nil {
				return
			}

		case <-disconnected:
			return
		case <-s.quit:
			return
		}
	}
}

func (s *TCPServer) handlePublishers(conn net.Conn) {
	defer conn.Close()

	pubCh := s.broker.RegisterPublisher()
	defer s.broker.DeregisterPublisher(pubCh)

	fmt.Fprintf(conn, "WELCOME: Connected to Oxy Pub/Sub Publisher Server\n")
	fmt.Fprintf(conn, "Commands: PUB <message>, QUIT\n")

	done := make(chan struct{})

	go func() {
		for {
			select {
			case notification, ok := <-pubCh:
				if !ok {
					return
				}
				fmt.Fprintf(conn, "INFO: %s\n", notification)
			case <-done:
				return
			}
		}
	}()

	defer close(done)

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		command := strings.ToUpper(parts[0])

		switch command {
		case "PUB":
			if len(parts) < 2 {
				fmt.Fprintf(conn, "ERROR: PUB requires a message\n")
				continue
			}

			s.broker.Publish(parts[1])
		case "QUIT":
			return
		default:
			fmt.Fprintf(conn, "ERROR: Command unknown")
		}
	}
}
