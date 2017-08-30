package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"time"
)

var (
	uri          = string("amqp://guest:guest@10.1.10.200:5672/test")
	exchange     = flag.String("exchange", "signal-exchange", "Durable, non-auto-deleted AMQP exchange name")
	exchangeType = flag.String("exchange-type", "direct", "Exchange type - direct|fanout|topic|x-custom")
	queue        = flag.String("queue", "signal-queue", "Ephemeral AMQP queue name")
	bindingKey   = flag.String("key", "signal-key", "AMQP binding key")
	consumerTag  = flag.String("consumer-tag", "signal-consumer", "AMQP consumer")
	lifetime     = flag.Duration("lifetime", 0*time.Second, "lifetime of process before shutdown (0s=infinite)")
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func createQueue(channel *amqp.Channel, queueName string) (*amqp.Channel, error) {
	log.Printf("declared Exchange, declaring Queue %q", queueName)
	queue, err := channel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when usused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Declare: %s", err)
	}
	log.Printf("declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, bindingKey)

	if err = channel.QueueBind(
		queue.Name,  // name of the queue
		*bindingKey, // bindingKey
		*exchange,   // sourceExchange
		false,       // noWait
		nil,         // arguments
	); err != nil {
		return nil, fmt.Errorf("Queue Bind: %s", err)
	}
	return channel, nil
}

func publish(c *amqp.Channel, body string) error {
	log.Printf("publishing %dB body (%s)", len(body), body)
	if c == nil {
		return fmt.Errorf("connection to rabbitmq might not be ready yet")
	}
	if err := c.Publish(
		*exchange,   // publish to an exchange
		*bindingKey, // routing to 0 or more queues
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "application/json",
			ContentEncoding: "",
			Body:            []byte(body),
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		},
	); err != nil {
		return fmt.Errorf("Exchange Publish: %s", err)
	}
	return nil
}

func main() {
	conn, err := amqp.Dial(uri)
	failOnError(err, "Failed to connect to RabbitMQ")

	out, err := json.Marshal(conn)
	body := "hello"
	fmt.Println(string(out))
	ch, err := conn.Channel()
	in, err := json.Marshal(ch)
	fmt.Println(string(in))

	failOnError(err, "Failed to open a channel")
	q, err := createQueue(ch, "hello")
	failOnError(err, "CQ")
	go func() {
		for {
			publish(q, "yo")
		}
	}()

	op, err := json.Marshal(q)
	fmt.Println(string(op))

	msgs, err := q.Consume(
		"hello", // queue
		"",      // consumer
		true,    // auto-ack
		false,   // exclusive
		false,   // no-local
		false,   // no-wait
		nil,     // args
	)
	failOnError(err, "Failed to register a consumer")
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf(" [x] Sent %s", body)
	failOnError(err, "Failed to publish a message")
	//
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
	defer ch.Close()
	defer q.Close()
	defer conn.Close()
}
