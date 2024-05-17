package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"lightblocks/common"
	"lightblocks/mymap"

	"github.com/streadway/amqp"
)

func processCommand(command common.Command, myorderedmap *mymap.OrderedSet, file *os.File) error {
	switch command.Command {
	case "addItem":
		// we don't run insertion in a separate go routine as we need to
		// maintain the insert order here. We could pass the values to channel
		// and run InsertElement operations in a separate go routine (maintaining the
		// insert order) but it wouldn't give us much speedup
		myorderedmap.InsertElement(command.Values[0], command.Values[1])
	case "deleteItem":
		go myorderedmap.RemoveElement(command.Values[0])
	case "getItem":
		go func(key string) {
			value := myorderedmap.GetElement(key)
			// I'm not sure I'm required to use lock here but file writes are not
			// inherently thread safe (from golang standard and internall from linux syscall)
			// so we may want to use lock to avoid let's say data corruptions when writing.
			// from different go routines in golang.
			if _, err := fmt.Fprintf(file, "%s\n", value); err != nil {
				fmt.Errorf("Couldn't get item from the map: %s", err.Error())
			}
		}(command.Values[0])
	case "getAllItems":
		go func() {
			values := myorderedmap.GetAllElements()
			for _, value := range values {
				if _, err := fmt.Fprintf(file, "%s\n", value); err != nil {
					fmt.Errorf("Couldn't get elements from the map: %s", err.Error())
				}
			}
		}()
	}

	return nil
}

// Server and client implementation is pretty straightforward. They simply parse command line
// parameters and connect to queue for receiving and sending commands correspondingly. We could also
// define interface and mock (probably with mockify package) those interfaces and have functional tests
// but as it's not too complicated is shouldn't be necessary.
func main() {
	if len(os.Args) < 3 {
		log.Fatalf("Usage: %s <RabbitMQ URL> <file path>", os.Args[0])
	}
	rabbitMQURL := os.Args[1]
	filePath := os.Args[2]

	// Validate RabbitMQ URL
	_, err := url.Parse(rabbitMQURL)
	if err != nil {
		log.Fatalf("queue url err: %s", err)
	}

	// Validate file path
	file, err := os.Create(filePath)
	defer file.Close()
	if err != nil {
		log.Fatalf("creating file failed: %s", err)
	}

	conn, err := amqp.Dial(rabbitMQURL)
	defer conn.Close()
	if err != nil {
		log.Fatalf("connecting queue failed: %s", err)
	}

	ch, err := conn.Channel()
	defer ch.Close()
	if err != nil {
		log.Fatalf("getting channel failed: %s", err)
	}

	c, err := ch.QueueDeclare(
		"commands", // name
		false,      // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		log.Fatalf("declaring channel failed: %s", err)
	}

	// Consume messages from the queue
	commands, err := ch.Consume(
		c.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("creating channel consumer failed: %s", err)
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// create internal data structure to track command data
	myOrderedMap := mymap.NewOrderedSet()

	for {
		select {
		case c := <-commands:
			var command common.Command
			err := json.Unmarshal(c.Body, &command)
			if err != nil {
				fmt.Errorf("Error unmarshalling command: %s", err.Error())
			}

			// run command in parallel
			processCommand(command, myOrderedMap, file)
		case <-shutdown:
			fmt.Println("\nshutting down server")
			return
		}
	}
}
