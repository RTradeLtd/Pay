package queue

import (
	log "github.com/sirupsen/logrus"

	"github.com/streadway/amqp"
)

var (
	// DashPaymentConfirmationQueue is a queue used to handle confirming dash payments
	DashPaymentConfirmationQueue = "dash-payment-confirmation-queue"
	// PaymentCreationQueue is a queue used to handle payment processing
	PaymentCreationQueue = "payment-creation-queue"
	// PaymentConfirmationQueue is a queue used to handle payment confirmations
	PaymentConfirmationQueue = "payment-confirmation-queue"
)

// Manager is a helper struct to interact with rabbitmq
type Manager struct {
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

// DashPaymentConfirmation is a message used to signal processing of a dash payment
type DashPaymentConfirmation struct {
	UserName         string `json:"user_name"`
	PaymentForwardID string `json:"payment_forward_id"`
	PaymentNumber    int64  `json:"payment_number"`
}
