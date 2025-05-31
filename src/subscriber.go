package broker

import "net"

type Subscriber struct {
	id         string
	connection net.Conn
	topics     []string
}
