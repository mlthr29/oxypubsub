package broker_test

import (
	"bufio"
	"net"
	"strings"
	"testing"
	"time"

	broker "oxypubsub/src"
)

func TestCreateTCPServer(t *testing.T) {
	b := broker.CreateBroker()
	server := broker.CreateTCPServer(":0", ":0", b)

	if server == nil {
		t.Fatal("Expected: new server.Actual: nil.")
	}
}

func TestTCPServerStartStop(t *testing.T) {
	b := broker.CreateBroker()
	server := broker.CreateTCPServer(":0", ":0", b)

	err := server.Start()
	if err != nil {
		t.Fatalf("Expected: server started. Actual: %v", err)
	}

	server.Stop()
}

func TestSubscriberConnection(t *testing.T) {
	b := broker.CreateBroker()
	server := broker.CreateTCPServer(":8082", ":8083", b)

	err := server.Start()
	if err != nil {
		t.Fatalf("Expected: server started. Actual: %v", err)
	}
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	conn, err := net.Dial("tcp", ":8082")
	if err != nil {
		t.Fatalf("Failed to connect to subscriber port: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	welcome, _, err := reader.ReadLine()
	if err != nil {
		t.Fatalf("Failed to read welcome message: %v", err)
	}

	if !strings.Contains(string(welcome), "WELCOME") {
		t.Fatalf("Expected: greeting message. Actual: %s", string(welcome))
	}
}

func TestPublisherConnection(t *testing.T) {
	b := broker.CreateBroker()
	server := broker.CreateTCPServer(":8084", ":8085", b)

	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	conn, err := net.Dial("tcp", ":8085")
	if err != nil {
		t.Fatalf("Failed to connect to publisher port: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	welcome, _, err := reader.ReadLine()
	if err != nil {
		t.Fatalf("Failed to read welcome message: %v", err)
	}

	if !strings.Contains(string(welcome), "WELCOME") {
		t.Fatalf("Expected: greeting message. Actual: %s", string(welcome))
	}
}

func TestPublishSubscribeFlow(t *testing.T) {
	b := broker.CreateBroker()
	server := broker.CreateTCPServer(":8086", ":8087", b)

	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	subConn, err := net.Dial("tcp", ":8086")
	if err != nil {
		t.Fatalf("Failed to connect subscriber: %v", err)
	}
	defer subConn.Close()

	pubConn, err := net.Dial("tcp", ":8087")
	if err != nil {
		t.Fatalf("Failed to connect publisher: %v", err)
	}
	defer pubConn.Close()

	subReader := bufio.NewReader(subConn)
	pubReader := bufio.NewReader(pubConn)

	subReader.ReadLine()
	pubReader.ReadLine()
	pubReader.ReadLine()

	testMessage := "Hello World"
	pubConn.Write([]byte("PUB " + testMessage + "\n"))

	response, _, err := subReader.ReadLine()
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	if !strings.Contains(string(response), testMessage) {
		t.Fatalf("Expected: message %q. Actual: %s", testMessage, string(response))
	}
}

func TestPublisherQuitCommand(t *testing.T) {
	b := broker.CreateBroker()
	server := broker.CreateTCPServer(":8088", ":8089", b)

	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	conn, err := net.Dial("tcp", ":8089")
	if err != nil {
		t.Fatalf("Failed to connect to publisher port: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	reader.ReadLine()
	reader.ReadLine()

	conn.Write([]byte("QUIT\n"))

	conn.SetReadDeadline(time.Now().Add(time.Second))
	_, err = reader.ReadByte()
	if err == nil {
		t.Fatal("Expected: connection closed. Actual: connection open.")
	}
}

func TestSubscriberConnectionNotifications(t *testing.T) {
	b := broker.CreateBroker()
	server := broker.CreateTCPServer(":8090", ":8091", b)

	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	pubConn, err := net.Dial("tcp", ":8091")
	if err != nil {
		t.Fatalf("Failed to connect as publisher: %v", err)
	}
	defer pubConn.Close()

	pubReader := bufio.NewReader(pubConn)
	pubReader.ReadLine()
	pubReader.ReadLine()

	subConn, err := net.Dial("tcp", ":8090")
	if err != nil {
		t.Fatalf("Failed to connect as subscriber: %v", err)
	}

	notification, _, err := pubReader.ReadLine()
	if err != nil {
		t.Fatalf("Failed to read subscriber join notification: %v", err)
	}

	if !strings.Contains(string(notification), "NEW SUBSCRIBER JOINED") {
		t.Fatalf("Expected: new subscriber notification. Actual: %s", string(notification))
	}

	time.Sleep(3 * time.Second)

	subConn.Close()

	notification, _, err = pubReader.ReadLine()
	if err != nil {
		t.Fatalf("Failed to read subscriber disconnect notification: %v", err)
	}

	if !strings.Contains(string(notification), "NO SUBSCRIBERS AVAILABLE") {
		t.Fatalf("Expected: no subscribers notification. Actual: %s", string(notification))
	}
}
