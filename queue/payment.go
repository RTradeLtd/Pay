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
		// TODO: ADD CALL TO SEE IF THIS PAYMENT IS ALREADY BEING PROCESSED
		// THIS WILL REQUIRE PUTTING A FLAG IN THE DB TO INDICATE PROCESSING
		switch payment.Type {
		case "eth":
			// process ethereum payments
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
			}).Info("processing ethereum based payment")
			// process the payment (waiting for confirmations, status checks, etc...)
			if err := service.Client.ProcessEthPaymentTx(pc.TxHash); err != nil {
				qm.Logger.WithFields(log.Fields{
					"service": qm.Service,
					"error":   err.Error(),
				}).Error("failed to wait for confirmations")
				d.Ack(false)
				continue
			}
			// payment confirmed
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
			}).Info("ethereum based payment confirmed")
		case "rtc":
			// process rtc payments
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
			}).Info("processing rtc based payment")
			// process the payment (waiting for confirmations, status checks, etc...)
			if err := service.Client.ProcessRtcPaymentTx(pc.TxHash); err != nil {
				qm.Logger.WithFields(log.Fields{
					"service": qm.Service,
					"error":   err.Error(),
				}).Error("failed to wait for confirmations")
				d.Ack(false)
				continue
			}
			// payment confirmed
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
			}).Info("rtc based payment confirmed")
		default:
			// indicates an unsupported payment type
			err = errors.New("unsupported payment type")
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("unsupported payment type")
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
		if _, err = service.UM.AddCreditsForUser(pc.UserName, payment.USDValue); err != nil {
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
