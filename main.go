package main

//import "oxypubsub/src/broker"

import (
	"fmt"
	broker "oxypubsub/src"
)

func main() {
	broker := broker.CreateBroker()

	broker.Start("10868")

	fmt.Println("Server started at 10868")
}
