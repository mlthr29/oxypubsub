package broker

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type TCPServer struct {
	address  string
	broker   *Broker
	listener net.Listener
	clients  map[string]*TCPClient
	mutex    sync.RWMutex
	quit     chan struct{}
}

type TCPClient struct {
	id            string
	conn          net.Conn
	broker        *Broker
	subscriptions map[string]<-chan string
	quit          chan struct{}
	mutex         sync.RWMutex
}

func CreateTCPServer(address string, broker *Broker) *TCPServer {
	return &TCPServer{
		address: address,
		broker:  broker,
		clients: make(map[string]*TCPClient),
		quit:    make(chan struct{}),
	}
}

func (s *TCPServer) Start() error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("Failed to start the TCP server: %v", err)
	}

	s.listener = listener

	for {
		select {
		case <-s.quit:
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-s.quit:
					return nil
				default:
					continue
				}
			}

			go s.handleConnection(conn)
		}
	}
}

func (s *TCPServer) Stop() {
	close(s.quit)
	if s.listener != nil {
		s.listener.Close()
	}

	s.mutex.Lock()
	for _, client := range s.clients {
		client.close()
	}
	s.mutex.Unlock()
}

func (s *TCPServer) handleConnection(conn net.Conn) {
	clientID := fmt.Sprintf("tcp-%s-%d", conn.RemoteAddr().String(), time.Now().UnixNano())

	client := &TCPClient{
		id:            clientID,
		conn:          conn,
		broker:        s.broker,
		subscriptions: make(map[string]<-chan string),
		quit:          make(chan struct{}),
	}

	s.mutex.Lock()
	s.clients[clientID] = client
	s.mutex.Unlock()

	fmt.Fprintf(conn, "WELCOME: Connected to Oxy Pub/Sub broker server\n")
	fmt.Fprintf(conn, "Commands: SUB <topic>, PUB <topic> <message>, LIST, QUIT\n")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		client.handleCommands()
	}()

	go func() {
		defer wg.Done()
		client.handleSubscriptions()
	}()

	wg.Wait()

	s.mutex.Lock()
	delete(s.clients, clientID)
	s.mutex.Unlock()

	conn.Close()
}

func (c *TCPClient) handleCommands() {
	defer c.close()

	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		command := strings.ToUpper(parts[0])

		switch command {
		case "SUB":
			if len(parts) < 2 {
				fmt.Fprintf(c.conn, "ERROR: SUB requires topic\n")
				continue
			}
			topic := parts[1]
			c.subscribe(topic)

		case "PUB":
			if len(parts) < 3 {
				fmt.Fprintf(c.conn, "ERROR: PUB requires topic and message\n")
				continue
			}
			topic := parts[1]
			message := strings.Join(parts[2:], " ")

			go func() {
				c.publish(topic, message)
			}()

		case "LIST":
			c.listTopics()

		case "QUIT":
			fmt.Fprintf(c.conn, "BYE: Disconnecting\n")
			return

		default:
			fmt.Fprintf(c.conn, "ERROR: Unknown command: %s\n", command)
		}
	}
}

func (c *TCPClient) handleSubscriptions() {
	defer c.close()

	for {
		select {
		case <-c.quit:
			return
		default:
			c.mutex.RLock()
			for topic, ch := range c.subscriptions {
				select {
				case msg := <-ch:
					fmt.Fprintf(c.conn, "MESSAGE %s: %s\n", topic, msg)
				default:
				}
			}
			c.mutex.RUnlock()

			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (c *TCPClient) subscribe(topic string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exists := c.subscriptions[topic]; exists {
		fmt.Fprintf(c.conn, "ERROR: Already subscribed to %s\n", topic)
		return
	}

	ch := c.broker.Subscribe(topic)
	c.subscriptions[topic] = ch

	fmt.Fprintf(c.conn, "SUBSCRIBED: %s\n", topic)
}

func (c *TCPClient) publish(topic, message string) {
	c.broker.Publish(topic, message)
}

func (c *TCPClient) listTopics() {
	topics := c.broker.GetTopics()

	if len(topics) == 0 {
		fmt.Fprintf(c.conn, "TOPICS: No active topics\n")
		return
	}

	fmt.Fprintf(c.conn, "TOPICS: ")
	for i, topic := range topics {
		if i > 0 {
			fmt.Fprintf(c.conn, ", ")
		}
		fmt.Fprintf(c.conn, "%s", topic)
	}
	fmt.Fprintf(c.conn, "\n")
}

func (c *TCPClient) close() {
	select {
	case <-c.quit:
		return
	default:
		close(c.quit)
	}
}
