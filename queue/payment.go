package queue

import (
	"encoding/json"

	"github.com/RTradeLtd/Temporal_Payment-ETH/service"
	"github.com/RTradeLtd/config"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// ProcessEthereumBasedPayment is used to process ethereum based payment messages
func (qm *QueueManager) ProcessEthereumBasedPayment(msgs <-chan amqp.Delivery, db *gorm.DB, cfg *config.TemporalConfig) error {
	service, err := service.GeneratePaymentService(cfg, "infura")
	if err != nil {
		return err
	}
	qm.Logger.WithFields(log.Fields{
		"service": qm.Service,
	}).Info("processing ethereum payment message")
	for d := range msgs {
		qm.Logger.WithFields(log.Fields{
			"service": qm.Service,
		}).Info("new message received")
		pc := PaymentCreation{}
		if err := json.Unmarshal(d.Body, &pc); err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to unmarshal message")
			d.Ack(false)
			continue
		}
		if _, err := service.PM.FindPaymentByTxHash(pc.TxHash); err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to find payment by tx hash")
		}
		switch pc.Type {
		case "eth":
			if err := service.Client.ProcessEthPaymentTx(pc.TxHash); err != nil {
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
