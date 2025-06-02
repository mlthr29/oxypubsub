package broker

import (
	"log"
	"sync"
)

type Broker struct {
	subscribers map[string][]chan string
	mutex       sync.RWMutex
}

func CreateBroker() *Broker {
	return &Broker{
		subscribers: make(map[string][]chan string, 0),
	}
}

func (broker *Broker) Subscribe(topic string) <-chan string {
	broker.mutex.Lock()
	defer broker.mutex.Unlock()

	ch := make(chan string, 10)
	broker.subscribers[topic] = append(broker.subscribers[topic], ch)

	log.Printf("Topic '%s' has a new subsciber!", topic)

	return ch
}

func (broker *Broker) Publish(topic, message string) {
	broker.mutex.RLock()
	defer broker.mutex.RUnlock()

	channels, exists := broker.subscribers[topic]
	if !exists {
		log.Printf("No subscribers for topic: '%s'", topic)
		return
	}

	for _, ch := range channels {
		go func(channel chan string) {
			select {
			case channel <- message:
			default:
				log.Println("Message was dropped because channel is full")
			}
		}(ch)
	}

	log.Println("Message successfully published to subscribers")
}

func (broker *Broker) GetTopics() []string {
	broker.mutex.RLock()
	defer broker.mutex.RUnlock()

	topics := make([]string, 0, len(broker.subscribers))
	for topic := range broker.subscribers {
		topics = append(topics, topic)
	}
	return topics
}
