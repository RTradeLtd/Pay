package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/RTradeLtd/Temporal/config"
	"github.com/RTradeLtd/Temporal/database"
	log "github.com/sirupsen/logrus"

	"github.com/streadway/amqp"
)

func (qm *QueueManager) setupLogging() error {
	logFileName := fmt.Sprintf("/var/log/temporal/%s_serice.log", qm.QueueName)
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
	if err != nil {
		return err
	}
	logger := log.New()
	logger.Out = logFile
	qm.Logger = logger
	qm.Logger.Info("Logging initialized")
	return nil
}

func (qm *QueueManager) parseQueueName(queueName string) error {
	host, err := os.Hostname()
	if err != nil {
		return err
	}
	qm.QueueName = fmt.Sprintf("%s+%s", host, queueName)
	return nil
}

// Initialize is used to connect to the given queue, for publishing or consuming purposes
func Initialize(queueName, connectionURL string, publish, service bool) (*QueueManager, error) {
	conn, err := setupConnection(connectionURL)
	if err != nil {
		return nil, err
	}
	qm := QueueManager{Connection: conn}
	if err := qm.OpenChannel(); err != nil {
		return nil, err
	}

	qm.QueueName = queueName
	qm.Service = queueName

	if service {
		err = qm.setupLogging()
		if err != nil {
			return nil, err
		}
	}

	if err := qm.DeclareQueue(); err != nil {
		return nil, err
	}
	return &qm, nil
}

func setupConnection(connectionURL string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(connectionURL)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (qm *QueueManager) OpenChannel() error {
	ch, err := qm.Connection.Channel()
	if err != nil {
		return err
	}
	if qm.Logger != nil {
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
		}).Info("channel opened")
	}
	qm.Channel = ch
	return nil
}

// DeclareQueue is used to declare a queue for which messages will be sent to
func (qm *QueueManager) DeclareQueue() error {
	// we declare the queue as durable so that even if rabbitmq server stops
	// our messages won't be lost
	q, err := qm.Channel.QueueDeclare(
		qm.QueueName, // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return err
	}
	if qm.Logger != nil {
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
		}).Info("queue declared")
	}
	qm.Queue = &q
	return nil
}

// ConsumeMessage is used to consume messages that are sent to the queue
// Question, do we really want to ack messages that fail to be processed?
// Perhaps the error was temporary, and we allow it to be retried?
func (qm *QueueManager) ConsumeMessage(consumer, dbPass, dbURL, dbUser string, cfg *config.TemporalConfig) error {
	db, err := database.OpenDBConnection(database.DBOptions{
		User: dbUser, Password: dbPass, Address: dbURL})
	if err != nil {
		return err
	}

	// consider moving to true for auto-ack
	msgs, err := qm.Channel.Consume(
		qm.QueueName, // queue
		consumer,     // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return err
	}

	// check the queue name
	switch qm.Service {
	// only parse datbase file requests
	case PaymentCreationQueue:
		return nil
	default:
		log.Fatal("invalid queue name")
	}
	return nil
}

//PublishMessageWithExchange is used to publish a message to a given exchange
func (qm *QueueManager) PublishMessageWithExchange(body interface{}, exchangeName string) error {
	switch exchangeName {
	case PinExchange:
		break
	case PinRemovalExchange:
		break
	case IpfsKeyExchange:
		break
	default:
		return errors.New("invalid exchange name provided")
	}
	bodyMarshaled, err := json.Marshal(body)
	if err != nil {
		return err
	}
	err = qm.Channel.Publish(
		exchangeName, // exchange
		"",           // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         bodyMarshaled,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

// PublishMessage is used to produce messages that are sent to the queue, with a worker queue (one consumer)
func (qm *QueueManager) PublishMessage(body interface{}) error {
	// we use a persistent delivery mode to combine with the durable queue
	bodyMarshaled, err := json.Marshal(body)
	if err != nil {
		return err
	}
	err = qm.Channel.Publish(
		"",            // exchange
		qm.Queue.Name, // routing key
		false,         // mandatory
		false,         //immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         bodyMarshaled,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (qm *QueueManager) Close() error {
	return qm.Connection.Close()
}
