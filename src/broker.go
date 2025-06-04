package broker

import (
	"sync"
)

type Broker struct {
	subscribers map[chan string]struct{}
	publishers  map[chan string]struct{}
	mutex       sync.RWMutex
}

func CreateBroker() *Broker {
	return &Broker{
		subscribers: make(map[chan string]struct{}),
		publishers:  make(map[chan string]struct{}),
	}
}

func (broker *Broker) Subscribe() chan string {
	broker.mutex.Lock()
	defer broker.mutex.Unlock()

	ch := make(chan string, 10)
	broker.subscribers[ch] = struct{}{}

	broker.notifyPublishers("A NEW SUBSCRIBER JOINED!")

	return ch
}

func (broker *Broker) Unsubscribe(ch chan string) {
	broker.mutex.Lock()
	defer broker.mutex.Unlock()

	if _, ok := broker.subscribers[ch]; ok {
		delete(broker.subscribers, ch)
		close(ch)

		if len(broker.subscribers) == 0 {
			broker.notifyPublishers("NO SUBSCRIBERS AVAILABLE!")
		}
	}
}

func (broker *Broker) RegisterPublisher() chan string {
	broker.mutex.Lock()
	defer broker.mutex.Unlock()

	ch := make(chan string, 10)
	broker.publishers[ch] = struct{}{}

	return ch
}

func (broker *Broker) DeregisterPublisher(ch chan string) {
	broker.mutex.Lock()
	defer broker.mutex.Unlock()

	if _, ok := broker.publishers[ch]; ok {
		delete(broker.publishers, ch)
		close(ch)
	}
}

func (broker *Broker) Publish(message string) {
	broker.mutex.RLock()
	defer broker.mutex.RUnlock()

	for ch := range broker.subscribers {
		select {
		case ch <- message:
		default:
		}
	}
}

func (broker *Broker) notifyPublishers(message string) {
	for ch := range broker.publishers {
		select {
		case ch <- message:
		default:
		}
	}
}
