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
		payment, err := service.PM.FindPaymentByTxHash(pc.TxHash)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to find payment by tx hash")
			d.Ack(false)
			continue
		}
		if payment.UserName != pc.UserName {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
			}).Warn("suspicious message username does not match what is in database")
			d.Ack(false)
			continue
		}
		if payment.Confirmed {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
			}).Warn("payment already confirmed")
			d.Ack(false)
			continue
		}
		switch pc.Type {
		case "eth":
			if err := service.Client.ProcessEthPaymentTx(pc.TxHash); err != nil {
				qm.Logger.Error("failed to wait for confirmations")
				d.Ack(false)
				continue
			}
		}
		if _, err := service.PM.ConfirmPayment(pc.TxHash); err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to confirm payment")
			d.Ack(false)
			continue
		}
		qm.Logger.WithFields(log.Fields{
			"service": qm.Service,
		}).Info("payment confirmed")
		if _, err = service.UM.AddCreditsForUser(pc.UserName, payment.USDValue); err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to add credits for user")
		}
		qm.Logger.WithFields(log.Fields{
			"service": qm.Service,
		}).Info("credits added for user")
		continue
	}
	return nil
}

func (qm *QueueManager) processRTCPayment() error {
	return nil
}
