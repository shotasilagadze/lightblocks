package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"lightblocks/common"

	"github.com/streadway/amqp"
)

func validateFilePath(filePath string) error {
	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Check if the file is readable
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("file is not readable: %s", filePath)
	}
	file.Close()
	return nil
}

func main() {
	// Parse command-line argument for file path
	filePath := flag.String("file", "", "Path to the JSON file")
	flag.Parse()

	if *filePath == "" {
		log.Fatalf("File path is required. Use -file flag to specify the path.")
	}

	// Validate the file path and permissions
	if err := validateFilePath(*filePath); err != nil {
		log.Fatalf("%s: %s", err, "Invalid file path or permissions")
	}

	// Read the JSON file
	file, err := ioutil.ReadFile(*filePath)
	if err != nil {
		log.Fatalf("%s: %s", err, "Failed to read JSON file")

	}

	// Parse the JSON array
	var commands []common.Command
	if err = json.Unmarshal(file, &commands); err != nil {
		log.Fatalf("%s: %s", err, "Failed to unmarshal JSON")
	}

	// Connect to RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("%s: %s", err, "Failed to connect to RabbitMQ")
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("%s: %s", err, "Failed to open a channel")
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"commands", // name
		false,      // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		log.Fatalf("%s: %s", err, "Failed to declare a queue")
	}

	// Iterate over the array and send each JSON object as a message
	for _, command := range commands {
		body, err := json.Marshal(command)
		if err != nil {
			log.Fatalf("%s: %s", err, "Failed to unmarshall json")
		}

		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        body,
			})
		if err != nil {
			log.Fatalf("%s: %s", err, "Failed to publish a message")
		}
	}

	fmt.Println("Successfully sent all JSON messages to RabbitMQ")
}
