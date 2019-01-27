package queue

import (
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/gorm"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

// Queue is a typed string used to declare the various queue names
type Queue string

func (qt Queue) String() string {
	return string(qt)
}

var (
	// DashPaymentConfirmationQueue is a queue used to handle confirming dash payments
	DashPaymentConfirmationQueue Queue = "dash-payment-confirmation-queue"
	// PaymentCreationQueue is a queue used to handle payment processing
	PaymentCreationQueue Queue = "payment-creation-queue"
	// PaymentConfirmationQueue is a queue used to handle payment confirmations
	PaymentConfirmationQueue Queue = "payment-confirmation-queue"
	// ErrReconnect is an error emitted when a protocol connection error occurs
	// It is used to signal reconnect of queue consumers and publishers
	ErrReconnect = "protocol connection error, reconnect"
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
