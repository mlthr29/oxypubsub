package main

import broker "oxypubsub/src"

func main() {
	broker := broker.CreateBroker()

	broker.Subscribe("test-topic")

	broker.Publish("test-topic", "first message published")
}
