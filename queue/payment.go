package queue

import (
	"encoding/json"
	"errors"

	"github.com/RTradeLtd/Temporal_Payment-ETH/service"
	"github.com/RTradeLtd/config"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// ProcessEthereumBasedPayment is used to process ethereum based payment messages
func (qm *QueueManager) ProcessEthereumBasedPayment(msgs <-chan amqp.Delivery, db *gorm.DB, cfg *config.TemporalConfig) error {
	// instantiate our payment service
	service, err := service.GeneratePaymentService(cfg, "infura")
	if err != nil {
		return err
	}
	qm.Logger.WithFields(log.Fields{
		"service": qm.Service,
	}).Info("processing ethereum based payments")
	// start processing messages
	for d := range msgs {
		// received new message
		qm.Logger.WithFields(log.Fields{
			"service": qm.Service,
		}).Info("new message received")
		// attempt to unmarshal
		pc := PaymentCreation{}
		if err := json.Unmarshal(d.Body, &pc); err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to unmarshal message")
			d.Ack(false)
			continue
		}
		// search for payment in DB, verifyying validity
		payment, err := service.PM.FindPaymentByTxHash(pc.TxHash)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to find payment by tx hash")
			d.Ack(false)
			continue
		}
		// verify the blockchain
		if pc.Blockchain != "ethereum" && pc.Blockchain != payment.Blockchain {
			err = errors.New("mismatching payment data between queue, and db")
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("mismatching payment data between queue, and db")
			d.Ack(false)
			continue
		}
		// verify username
		if payment.UserName != pc.UserName {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
			}).Warn("suspicious message username does not match what is in database")
			d.Ack(false)
			continue
		}
		// check if the payment is already confirmed
		if payment.Confirmed {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
			}).Warn("payment already confirmed")
			d.Ack(false)
			continue
		}
		if err = service.Client.ProcessPaymentTx(pc.TxHash); err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to process payment tx")
			d.Ack(false)
			continue
		}
		// mark payment as confirmed in the db
		if _, err = service.PM.ConfirmPayment(pc.TxHash); err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to confirm payment")
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
		// start processing new payments
		qm.Logger.WithFields(log.Fields{
			"service": qm.Service,
		}).Info("credits added for user")
		d.Ack(false)
		continue
	}
	return nil
}
