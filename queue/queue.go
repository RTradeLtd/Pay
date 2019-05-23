package queue

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io/ioutil"
	"sync"

	"github.com/RTradeLtd/gorm"
	"go.uber.org/zap"

	"github.com/RTradeLtd/config/v2"
	"github.com/streadway/amqp"
)

// Manager is a helper struct to interact with rabbitmq
type Manager struct {
	connection   *amqp.Connection
	channel      *amqp.Channel
	queue        *amqp.Queue
	l            *zap.SugaredLogger
	db           *gorm.DB
	cfg          *config.TemporalConfig
	ErrCh        chan *amqp.Error
	QueueName    Queue
	ExchangeName string
}

// New is used to instantiate a new connection to rabbitmq as a publisher or consumer
func New(queue Queue, cfg *config.TemporalConfig, logger *zap.SugaredLogger, publish bool) (*Manager, error) {
	conn, err := setupConnection(cfg.RabbitMQ.URL, cfg)
	if err != nil {
		return nil, err
	}
	var queueType string
	if publish {
		queueType = "publish"
	} else {
		queueType = "consumer"
	}
	// create base queue manager
	qm := Manager{connection: conn, QueueName: queue, l: logger.Named(queue.String() + "." + queueType)}
	// open a channel
	if err := qm.openChannel(); err != nil {
		return nil, err
	}
	// if we aren't publishing, and are consuming
	// setup a queue to receive messages on
	if !publish {
		if err = qm.declareQueue(); err != nil {
			return nil, err
		}
	}
	// register err channel notifier
	qm.RegisterConnectionClosure()
	return &qm, nil
}

func setupConnection(connectionURL string, cfg *config.TemporalConfig) (*amqp.Connection, error) {
	var (
		conn *amqp.Connection
		err  error
	)
	if cfg.RabbitMQ.TLSConfig.CACertFile == "" {
		conn, err = amqp.Dial(connectionURL)
	} else {
		// see https://godoc.org/github.com/streadway/amqp#DialTLS for more information
		tlsConfig := new(tls.Config)
		tlsConfig.RootCAs = x509.NewCertPool()
		ca, err := ioutil.ReadFile(cfg.RabbitMQ.TLSConfig.CACertFile)
		if err != nil {
			return nil, err
		}
		if ok := tlsConfig.RootCAs.AppendCertsFromPEM(ca); !ok {
			return nil, errors.New("failed to successfully append cert file")
		}
		cert, err := tls.LoadX509KeyPair(cfg.RabbitMQ.TLSConfig.CertFile, cfg.RabbitMQ.TLSConfig.KeyFile)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
		conn, err = amqp.DialTLS(connectionURL, tlsConfig)
	}
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// OpenChannel is used to open a channel to the rabbitmq server
func (qm *Manager) openChannel() error {
	ch, err := qm.connection.Channel()
	if err != nil {
		return err
	}
	qm.l.Info("channel opened")
	qm.channel = ch
	return qm.channel.Qos(10, 0, false)
}

// DeclareQueue is used to declare a queue for which messages will be sent to
func (qm *Manager) declareQueue() error {
	// we declare the queue as durable so that even if rabbitmq server stops
	// our messages won't be lost
	q, err := qm.channel.QueueDeclare(
		qm.QueueName.String(), // name
		true,                  // durable
		false,                 // delete when unused
		false,                 // exclusive
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		return err
	}
	qm.l.Info("queue declared")
	qm.queue = &q
	return nil
}

// ConsumeMessages is used to consume messages that are sent to the queue
// Question, do we really want to ack messages that fail to be processed?
// Perhaps the error was temporary, and we allow it to be retried?
func (qm *Manager) ConsumeMessages(ctx context.Context, wg *sync.WaitGroup, db *gorm.DB, cfg *config.TemporalConfig) error {
	// embed database into queue manager
	qm.db = db
	// embed config into queue manager
	qm.cfg = cfg

	// we do not auto-ack, as if a consumer dies we don't want the message to be lost
	// not specifying the consumer name uses an automatically generated id
	msgs, err := qm.channel.Consume(
		qm.QueueName.String(), // queue
		"",                    // consumer
		false,                 // auto-ack
		false,                 // exclusive
		false,                 // no-local
		false,                 // no-wait
		nil,                   // args
	)
	if err != nil {
		return err
	}

	// check the queue name
	switch qm.QueueName {
	case DashPaymentConfirmationQueue:
		return qm.ProcessDASHPayment(ctx, wg, msgs)
	case EthPaymentConfirmationQueue:
		return qm.ProcessETHPayment(ctx, wg, msgs)
	case BitcoinCashPaymentConfirmationQueue:
		return qm.ProcessBCHPayment(ctx, wg, msgs)
	default:
		return errors.New("invalid queue name")
	}
}

// PublishMessage is used to produce messages that are sent to the queue, with a worker queue (one consumer)
func (qm *Manager) PublishMessage(body interface{}) error {
	bodyMarshaled, err := json.Marshal(body)
	if err != nil {
		return err
	}
	if err = qm.channel.Publish(
		"",                    // exchange - this is left empty, and becomes the default exchange
		qm.QueueName.String(), // routing key
		false,                 // mandatory
		false,                 // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent, // messages will persist through crashes, etc..
			ContentType:  "text/plain",
			Body:         bodyMarshaled,
		},
	); err != nil {
		return err
	}
	return nil
}

// RegisterConnectionClosure is used to register a channel which we may receive
// connection level errors. This covers all channel, and connection errors.
func (qm *Manager) RegisterConnectionClosure() {
	qm.ErrCh = qm.connection.NotifyClose(make(chan *amqp.Error))
}

// Close is used to close our queue resources
func (qm *Manager) Close() error {
	// closing the connection also closes the channel
	return qm.connection.Close()
}
