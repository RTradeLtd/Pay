package queue

import (
	log "github.com/sirupsen/logrus"

	"github.com/streadway/amqp"
)

var (
	// PaymentCreationQueue is a queue used to handle payment processing
	PaymentCreationQueue = "payment-creation-queue"
	// PaymentConfirmationQueue is a queue used to handle payment confirmations
	PaymentConfirmationQueue = "payment-confirmation-queue"
)

// QueueManager is a helper struct to interact with rabbitmq
type QueueManager struct {
	Connection   *amqp.Connection
	Channel      *amqp.Channel
	Queue        *amqp.Queue
	Logger       *log.Logger
	QueueName    string
	Service      string
	ExchangeName string
}

// PaymentCreation is for the payment creation queue
type PaymentCreation struct {
	TxHash     string `json:"tx_hash"`
	Blockchain string `json:"blockchain"`
	UserName   string `json:"user_name"`
}

// PaymentConfirmation is a message used to confirm a payment
type PaymentConfirmation struct {
	UserName      string `json:"user_name"`
	PaymentNumber int64  `json:"payment_number"`
}
