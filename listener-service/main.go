package main

import (
	"fmt"
	"listener/event"
	"log"
	"math"
	"os"
	"time"

	// AMQP is a messaging protocol used for communication between client applications and messaging brokers. It defines a set of rules and formats that ensure interoperability between different components of a messaging system.
	// https://www.rabbitmq.com/tutorials/amqp-concepts.html
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// try to connect to rabbitmq
	rabbitConn, err := connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rabbitConn.Close()

	// start listening for messages
	log.Println("Listening for and consuming RabbitMQ messages...")

	// create a consumer that consumes message from the queue
	consumer, err := event.NewConsumer(rabbitConn)
	if err != nil {
		log.Println(err)
	}

	// watch the queue and comsume events
	err = consumer.Listen([]string{"log.INFO", "log.WARNING", "log.ERROR"})
	if err != nil {
		log.Println(err)
	}
}

func connect() (*amqp.Connection, error) {
	var counts int64
	var backoff = 1 * time.Second
	var connection *amqp.Connection

	// don't continue until rabbit is ready
	for {
		// rabbitmq matches our service in docker compose
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			fmt.Println("RabbitMQ not yet ready...")
			counts++
		} else {
			log.Println("Connected to RabbitMQ")
			connection = c
			break
		}

		// If we didn't connect :

		// if we didn't connect after 5 tries
		if counts > 5 {
			fmt.Println(err)
			return nil, err
		}

		// if we haven't tried atleast 5 times ,then call backoff
		backoff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("backing off....")
		// suspends the execution of the loop for the calculated duration.
		time.Sleep(backoff)
		continue
	}

	return connection, nil
}

// AMQP stands for Advanced Message Queuing Protocol. It is an open standard application layer protocol for message-oriented middleware. AMQP defines a set of rules for how messages are formatted, delivered, and routed between applications. It is supported by a wide range of messaging brokers, including RabbitMQ.

// Imagine you have two applications, A and B. You want A to be able to send messages to B, but you don't want A to have to wait for B to be ready to receive the message. You also don't want A to have to know anything about B's specific location or address.AMQP can help you solve this problem. AMQP provides a way for A to send messages to B without having to know anything about B's specific location or address. AMQP does this by using a messaging broker. The messaging broker is like a post office. It receives messages from A and stores them until B is ready to receive them.When B is ready to receive messages, it connects to the messaging broker and asks for messages. The messaging broker then delivers the messages to B.AMQP is a very flexible protocol. It can be used to send messages between applications in a variety of ways. For example, you can use AMQP to send messages between applications that are running on the same computer, or you can use AMQP to send messages between applications that are running on different computers in different parts of the world.

// Here are some of the key features of AMQP:

// Message orientation AMQP messages are sent and received in a message-oriented way, meaning that they are not tied to a specific connection or session. This makes AMQP more scalable and reliable than other messaging protocols, such as TCP/IP sockets.

// Queuing AMQP messages can be stored in queues, which allows applications to send messages without having to wait for a recipient to be available. This can improve performance and reliability, especially in high-volume messaging applications.

// Routing AMQP messages can be routed to specific recipients based on a variety of criteria, such as the message content or the recipient's address. This allows applications to decouple from each other and communicate in a loosely coupled way.

// Security AMQP supports a variety of security features, such as authentication, authorization, and encryption. This helps to protect messages from unauthorized access and tampering.
