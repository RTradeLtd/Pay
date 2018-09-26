package queue

import (
	"encoding/json"

	"github.com/RTradeLtd/Temporal_Payment-ETH/ethereum"
	"github.com/RTradeLtd/config"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

// ProcessEthereumBasedPayment is used to process ethereum based payment messages
func (qm *QueueManager) ProcessEthereumBasedPayment(msgs <-chan amqp.Delivery, db *gorm.DB, cfg *config.TemporalConfig) error {
	eClient, err := ethereum.NewClient(cfg, "infura")
	if err != nil {
		return err
	}
	for d := range msgs {
		pc := PaymentCreation{}
		if err := json.Unmarshal(d.Body, &pc); err != nil {
			qm.Logger.Error("failed to unmarshal message")
			d.Ack(false)
			continue
		}
		switch pc.Type {
		case "eth":
			if err := eClient.ProcessEthPaymentTx(pc.TxHash); err != nil {
				qm.Logger.Error("failed to wait for confirmations")
				d.Ack(false)
				continue
			}
		}
	}
	return nil
}

func (qm *QueueManager) processRTCPayment() error {
	return nil
}
