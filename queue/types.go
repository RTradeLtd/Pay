package queue

// Queue is a typed string used to declare the various queue names
type Queue string

func (qt Queue) String() string {
	return string(qt)
}

var (
	// EmailSendQueue is a queue used to handle sending email messages
	EmailSendQueue Queue = "email-send-queue"
	// DashPaymentConfirmationQueue is a queue used to handle confirming dash payments
	DashPaymentConfirmationQueue Queue = "dash-payment-confirmation-queue"
	// EthPaymentConfirmationQueue is a queue used to handle ethereum based payment confirmations
	EthPaymentConfirmationQueue Queue = "eth-payment-confirmation-queue"
	// ErrReconnect is an error emitted when a protocol connection error occurs
	// It is used to signal reconnect of queue consumers and publishers
	ErrReconnect = "protocol connection error, reconnect"
)

// EthPaymentConfirmation is a message used to confirm an ethereum based payment
type EthPaymentConfirmation struct {
	UserName      string `json:"user_name"`
	PaymentNumber int64  `json:"payment_number"`
}

// DashPaymentConfirmation is a message used to signal processing of a dash payment
type DashPaymentConfirmation struct {
	UserName         string `json:"user_name"`
	PaymentForwardID string `json:"payment_forward_id"`
	PaymentNumber    int64  `json:"payment_number"`
}

// BchPaymentConfirmation is used to confirm a bitcoin cash based payment
type BchPaymentConfirmation struct {
	UserName         string `json:"user_name"`
	SenderAddress    string `json:"sender_address"`
	RecipientAddress string `json:"recipient_address"`
	PaymentNumber    int64  `json:"payment_number"`
}

// EmailSend is a helper struct used to contained formatted content ot send as an email
type EmailSend struct {
	Subject     string   `json:"subject"`
	Content     string   `json:"content"`
	ContentType string   `json:"content_type"`
	UserNames   []string `json:"user_names"`
	Emails      []string `json:"emails,omitempty"`
}
