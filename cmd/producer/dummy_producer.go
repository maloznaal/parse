package main

import (
	"encoding/json"
	"fmt"
	"log"
	"offline_parser/utils"

	"github.com/streadway/amqp"
)

func handleError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

// use read from utils and print
var valz []string

func main() {
	// configure & connect dummy rabbit on default cred-s
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	handleError(err, "Can't connect to AMQP")
	defer conn.Close()

	amqpChannel, err := conn.Channel()
	handleError(err, "Can't create a amqpChannel")

	defer amqpChannel.Close()

	queue, err := amqpChannel.QueueDeclare(utils.TQueueName, true, false, false, false, nil)
	handleError(err, "Could not declare `add` queue")

	stopChan := make(chan bool)
	// producing k files into test Queue
	k := 1
	for k <= 5 {
		// read file and publish to rabbit mq
		utils.ReadFile(fmt.Sprintf("../tests/test%d.cdr", k), &valz)
		task := utils.ParseTask{Data: valz, FileName: fmt.Sprintf("out%d", k)}
		body, err := json.Marshal(task)
		if err != nil {
			handleError(err, "Error encoding JSON")
		}
		err = amqpChannel.Publish("", queue.Name, false, false, amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         body,
		})
		if err != nil {
			log.Fatalf("Error publishing message: %s", err)
		}
		log.Printf("Producing task with len %d", len(task.Data))
		// manualy free memory -> cause not writing to disk
		valz = nil
		k++
	}
	// blocks
	<-stopChan
}
