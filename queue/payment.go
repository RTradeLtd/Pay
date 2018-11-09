package queue

import (
	"encoding/json"
	"fmt"

	"github.com/RTradeLtd/Pay/dash"
	"github.com/RTradeLtd/Pay/service"
	"github.com/RTradeLtd/config"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// ProcessPaymentConfirmation is a queue process used to validate eth/rtc payments
func (qm *Manager) ProcessPaymentConfirmation(msgs <-chan amqp.Delivery, db *gorm.DB, cfg *config.TemporalConfig) error {
	service, err := service.GeneratePaymentService(cfg, &service.Opts{EthereumEnabled: true, DashEnabled: true})
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
			if _, err := service.Dash.ProcessTransaction(payment.TxHash); err != nil {
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

// ProcessDashPaymentConfirmation is used to process a dash payment confirmation message
func (qm *Manager) ProcessDashPaymentConfirmation(msgs <-chan amqp.Delivery, db *gorm.DB, cfg *config.TemporalConfig) error {
	service, err := service.GeneratePaymentService(cfg, &service.Opts{EthereumEnabled: false, DashEnabled: true})
	if err != nil {
		return err
	}
	qm.LogInfo("processing payment messages")
	for d := range msgs {
		qm.LogInfo("new message received")
		msg := DashPaymentConfirmation{}
		if err = json.Unmarshal(d.Body, &msg); err != nil {
			qm.LogError(err, "failed to unmarshal message")
			d.Ack(false)
			continue
		}
		paymentForward, err := service.Dash.C.GetPaymentForwardByID(msg.PaymentForwardID)
		if err != nil {
			qm.LogError(err, "failed to get payment forward by id")
			d.Ack(false)
			continue
		}
		payment, err := service.PM.FindPaymentByNumber(msg.UserName, msg.PaymentNumber)
		if err != nil {
			qm.LogError(err, "failed to search for payment")
			d.Ack(false)
			continue
		}
		opts := dash.ProcessPaymentOpts{
			Number:         payment.Number,
			ChargeAmount:   payment.ChargeAmount,
			PaymentForward: paymentForward,
		}
		if err = service.Dash.ProcessPayment(&opts); err != nil {
			qm.LogError(err, "failed to process payment")
		}
		// during processing, the user may have sent additional payments so need to re-grab them
		paymentForward, err = service.Dash.C.GetPaymentForwardByID(msg.PaymentForwardID)
		if err != nil {
			qm.LogError(err, "failed to get payment forward by id")
			d.Ack(false)
			continue
		}
		if len(paymentForward.ProcessedTxs) == 0 {
			qm.LogError(err, "no processed transactions detected")
			d.Ack(false)
			continue
		}
		if _, err = service.PM.ConfirmPayment(fmt.Sprintf("%s-%v", msg.UserName, msg.PaymentNumber)); err != nil {
			qm.LogError(err, "failed to confirm payment")
			d.Ack(false)
			continue
		}
		if _, err = service.UM.AddCredits(msg.UserName, payment.USDValue); err != nil {
			qm.LogError(err, "failed to grant credits")
			d.Ack(false)
			continue
		}
		qm.LogInfo("credits granted")
		continue
	}
	return nil
}
