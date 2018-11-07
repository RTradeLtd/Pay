package queue

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
	log "github.com/sirupsen/logrus"

	"github.com/streadway/amqp"
)

func (qm *Manager) setupLogging() error {
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

func (qm *Manager) parseQueueName(queueName string) error {
	host, err := os.Hostname()
	if err != nil {
		return err
	}
	qm.QueueName = fmt.Sprintf("%s+%s", host, queueName)
	return nil
}

// Initialize is used to connect to the given queue, for publishing or consuming purposes
func Initialize(queueName, connectionURL string, publish, service bool) (*Manager, error) {
	conn, err := setupConnection(connectionURL)
	if err != nil {
		return nil, err
	}
	qm := Manager{Connection: conn}
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

func (qm *Manager) OpenChannel() error {
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
func (qm *Manager) DeclareQueue() error {
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
func (qm *Manager) ConsumeMessage(consumer, dbPass, dbURL, dbUser string, cfg *config.TemporalConfig) error {
	fmt.Println("connecting to database")
	db, err := database.OpenDBConnection(database.DBOptions{
		User: dbUser, Password: dbPass, Address: dbURL, Port: "5432"})
	if err != nil {
		return err
	}
	fmt.Println("consuming messages")
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
	case PaymentConfirmationQueue:
		return qm.ProcessPaymentConfirmation(msgs, db, cfg)
	case DashPaymentConfirmationQueue:
		return qm.ProcessDashPaymentConfirmation(msgs, db, cfg)
	default:
		log.Fatal("invalid queue name")
	}
	return nil
}

// PublishMessage is used to produce messages that are sent to the queue, with a worker queue (one consumer)
func (qm *Manager) PublishMessage(body interface{}) error {
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

// Close is used to close a rabbitmq connection
func (qm *Manager) Close() error {
	return qm.Connection.Close()
}
