package main

import (
	"encoding/json"
	"fmt"
	"log"
	"offline_parser/utils"
	"os"
	"strings"

	"github.com/streadway/amqp"
)

// tempo for holding data
var valz []string
var m = map[string]bool{}
var str strings.Builder

// init expected result slices of strings
func checkResult(k int) {
	tvalz := make([]string, 0)
	utils.ReadFile(fmt.Sprintf("../clean_csv/test%d.cdr", k), &tvalz)
	if len(tvalz) != len(valz) {
		log.Fatalf("ITest fail on test%d expected len: %d got %d", k, len(tvalz), len(valz))
		os.Exit(1)
	}

	for i := range tvalz {
		exp, got := tvalz[i], valz[i]
		if exp != got {
			log.Fatalf("ITest fail on test%d expected val: %s got %s", k, exp, got)
			os.Exit(1)
		}
	}
	log.Printf("Test number %d finished with success", k)
	valz, tvalz = valz[:0], tvalz[:0]
	valz, tvalz = nil, nil
}

func init() {
	utils.Gen(len(utils.S), 3, "", &m)
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	utils.HandleError(err, "Can't connect to rabbit")
	defer conn.Close()
	amqpChannel, err := conn.Channel()
	utils.HandleError(err, "Can't create AMQP channel")
	defer amqpChannel.Close()
	queue, err := amqpChannel.QueueDeclare(utils.TQueueName, true, false, false, false, nil)
	utils.HandleError(err, "Couldn't declare 'add' queue")
	err = amqpChannel.Qos(1, 0, false)
	utils.HandleError(err, "Couldn't configure 'qos'")
	messageChannel, err := amqpChannel.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	fmt.Println(queue, messageChannel)
	utils.HandleError(err, "Couldn't register consumer")

	stopChan := make(chan bool)
	// consuming messages asyncronously
	k := 1
	go func() {
		log.Printf("Consumer ready, PID: %d", os.Getpid())
		for d := range messageChannel {
			log.Printf("Received a message: %s", "bytes")
			task := &utils.ParseTask{}
			err := json.Unmarshal(d.Body, task)
			if err != nil {
				log.Printf("Error decoding JSON: %s", err)
			}
			log.Printf("Len of data %d", len(task.Data))
			valz = task.Data
			utils.ParseJob(&valz, &str, &m)
			utils.WriteJob(fmt.Sprintf("test%d", k), "../clean_csv", &valz)
			if err := d.Ack(false); err != nil {
				utils.HandleError(err, "Error acking message")
			} else {
				log.Printf("Acked message")
			}
			// check rez againt exp rez
			checkResult(k)
			k++
		}
	}()
	// test run succed
	os.Exit(0)
	// blocking, until received value
	<-stopChan
}
