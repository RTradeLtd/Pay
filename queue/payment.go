package queue

import (
	"encoding/json"

	"github.com/RTradeLtd/Pay/service"
	"github.com/RTradeLtd/config"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// ProcessPaymentConfirmation is a queue process used to validate eth/rtc payments
func (qm *QueueManager) ProcessPaymentConfirmation(msgs <-chan amqp.Delivery, db *gorm.DB, cfg *config.TemporalConfig) error {
	service, err := service.GeneratePaymentService(cfg)
	if err != nil {
		return err
	}
	qm.Logger.WithFields(log.Fields{
		"service": qm.Service,
	}).Info("processing payment confirmations")
	for d := range msgs {
		qm.Logger.WithFields(log.Fields{
			"service": qm.Service,
		}).Info("new message received")
		pc := PaymentConfirmation{}
		if err := json.Unmarshal(d.Body, &pc); err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to unmarshal message")
			d.Ack(false)
			continue
		}
		payment, err := service.PM.FindPaymentByNumber(pc.UserName, pc.PaymentNumber)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to find payment")
			d.Ack(false)
			continue
		}
		if payment.Blockchain == "ethereum" {
			if err = service.Client.ProcessPaymentTx(payment.TxHash); err != nil {
				qm.Logger.WithFields(log.Fields{
					"service": qm.Service,
					"error":   err.Error(),
				}).Error("failedto validate payment")
				d.Ack(false)
				continue
			}
		}
		if payment.Blockchain == "dash" {
			if err := service.Dash.ProcessTransaction(payment.TxHash); err != nil {
				qm.Logger.WithFields(log.Fields{
					"service": qm.Service,
					"error":   err.Error(),
				}).Error("failedto validate payment")
				d.Ack(false)
				continue
			}
		}
		if _, err = service.PM.ConfirmPayment(payment.TxHash); err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to confirm payment in db")
			d.Ack(false)
			continue
		}
		// grant credits to the user
		if _, err = service.UM.AddCredits(pc.UserName, payment.USDValue); err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to add credits for user")
			d.Ack(false)
			continue
		}
		qm.Logger.WithFields(log.Fields{
			"service": qm.Service,
		}).Info("payment confirmed and credits granted")
		d.Ack(false)
	}
	return nil
}
