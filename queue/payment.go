package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/RTradeLtd/Pay/dash"
	"github.com/RTradeLtd/Pay/log"
	"github.com/RTradeLtd/Pay/service"
	"github.com/RTradeLtd/database/models"
	"github.com/streadway/amqp"
)

// ProcessETHPayment is used to process ethereum and rtc based payments
func (qm *Manager) ProcessETHPayment(ctx context.Context, wg *sync.WaitGroup, msgs <-chan amqp.Delivery) error {
	service, err := service.NewPaymentService(ctx, qm.cfg, &service.Opts{EthereumEnabled: true}, "rpc")
	if err != nil {
		return err
	}
	logger, err := log.NewLogger(qm.cfg.LogDir+"pay_eth_email_publisher.log", false)
	if err != nil {
		return err
	}
	qmEmail, err := New(EmailSendQueue, qm.cfg, logger, true)
	if err != nil {
		return err
	}
	um := models.NewUserManager(qm.db)
	qm.l.Info("processing payment confirmations")
	for {
		select {
		case d := <-msgs:
			wg.Add(1)
			go qm.processETHPayment(d, wg, service, qmEmail, um)
		case <-ctx.Done():
			qm.Close()
			wg.Done()
			return nil
		case msg := <-qm.ErrCh:
			qm.Close()
			wg.Done()
			qm.l.Errorw(
				"a protocol connection error stopping rabbitmq was received",
				"error", msg.Error())
			return errors.New(ErrReconnect)
		}
	}
}

func (qm *Manager) processETHPayment(d amqp.Delivery, wg *sync.WaitGroup, service *service.PaymentService, qmEmail *Manager, um *models.UserManager) {
	defer wg.Done()
	qm.l.Info("new ethereum based payment message received")
	pc := EthPaymentConfirmation{}
	if err := json.Unmarshal(d.Body, &pc); err != nil {
		qm.l.Error("failed to unmarshal message")
		d.Ack(false)
		return
	}
	logger := qm.l.With("user", pc.UserName).With("number", pc.PaymentNumber).With("currency", "ethereum")
	payment, err := service.PM.FindPaymentByNumber(pc.UserName, pc.PaymentNumber)
	if err != nil {
		logger.Errorw("failed to find payment message", "error", err.Error())
		d.Ack(false)
		return
	}
	switch payment.Blockchain {
	case "ethereum":
		// occassionally we may be given the hash before our node can find it in the blockchain or mempool
		// if this happens, we will wait 15 seconds before trying again. a total of 3 attempts are made
		// after which, we stop processing this transaction
		var found bool
		for count := 0; count < 3; count++ {
			if err := service.Client.ProcessPaymentTx(payment.TxHash); err != nil {
				if count < 3 {
					logger.Warnw("failed to find payment, waiting before attempting again", "error", err.Error())
					time.Sleep(time.Second * 15)
					continue
				}
			} else {
				found = true
				break
			}
		}
		// make sure we were able to find the transaction
		if !found {
			logger.Errorw("failed to find payment transaction after 3 repeated attempts")
			d.Ack(false)
			return
		}
	default:
		logger.Errorw("invalid blockchain for crypto payments")
		d.Ack(false)
		return
	}
	if _, err = service.PM.ConfirmPayment(payment.TxHash); err != nil {
		logger.Errorw("failed to confirm payment in database", "error", err.Error())
		d.Ack(false)
		return
	}
	// grant credits to the user
	if _, err = service.UM.AddCredits(pc.UserName, payment.USDValue); err != nil {
		logger.Errorw("failed to add credits for user", "error", err.Error())
		d.Ack(false)
		return
	}
	logger.Infow("successfully confirmed payment", "credits", payment.USDValue)
	user, err := um.FindByUserName(payment.UserName)
	if err != nil {
		logger.Errorw("failed to find email for user", "error", err.Error())
		d.Ack(false)
		return
	}
	if !user.EmailEnabled {
		logger.Warnw("user has not activated their email and won't receive notifications")
		d.Ack(false)
		return
	}
	es := EmailSend{
		Subject:     "Ethereum Payment Confirmed",
		Content:     fmt.Sprintf("Your ethereum payment for %v credits has been confirmed", payment.USDValue),
		ContentType: "text/html",
		UserNames:   []string{payment.UserName},
		Emails:      []string{user.EmailAddress},
	}
	if err := qmEmail.PublishMessage(es); err != nil {
		logger.Warnw("failed to send payment confirmation email")
	}
	d.Ack(false)
	return
}

// ProcessDASHPayment is used to process dash based payments
func (qm *Manager) ProcessDASHPayment(ctx context.Context, wg *sync.WaitGroup, msgs <-chan amqp.Delivery) error {
	service, err := service.NewPaymentService(ctx, qm.cfg, &service.Opts{DashEnabled: true}, "rpc")
	if err != nil {
		return err
	}
	logger, err := log.NewLogger(qm.cfg.LogDir+"pay_eth_email_publisher.log", false)
	if err != nil {
		return err
	}
	qmEmail, err := New(EmailSendQueue, qm.cfg, logger, true)
	if err != nil {
		return err
	}
	um := models.NewUserManager(qm.db)
	qm.l.Info("processing dash payment confirmations")
	for {
		select {
		case d := <-msgs:
			wg.Add(1)
			go qm.processDashPaymentConfirmation(d, wg, service, qmEmail, um)
		case <-ctx.Done():
			qm.Close()
			wg.Done()
			return nil
		case msg := <-qm.ErrCh:
			qm.Close()
			wg.Done()
			qm.l.Errorw(
				"a protocol connection error stopping rabbitmq was received",
				"error", msg.Error())
			return errors.New(ErrReconnect)
		}
	}
}

func (qm *Manager) processDashPaymentConfirmation(d amqp.Delivery, wg *sync.WaitGroup, service *service.PaymentService, qmEmail *Manager, um *models.UserManager) {
	defer wg.Done()
	qm.l.Info("new dash payment message received")
	msg := DashPaymentConfirmation{}
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		qm.l.Error("failed to unmarshal message", "error", err.Error())
		d.Ack(false)
		return
	}
	logger := qm.l.With("user", msg.UserName).With("number", msg.PaymentNumber).With("currency", "dash")
	paymentForward, err := service.Dash.C.GetPaymentForwardByID(msg.PaymentForwardID)
	if err != nil {
		logger.Errorw("failed to get payment forward by id", "error", err.Error())
		d.Ack(false)
		return
	}
	payment, err := service.PM.FindPaymentByNumber(msg.UserName, msg.PaymentNumber)
	if err != nil {
		logger.Errorw("failed to search for payment by number", "error", err.Error())
		d.Ack(false)
		return
	}
	opts := dash.ProcessPaymentOpts{
		Number:         payment.Number,
		ChargeAmount:   payment.ChargeAmount,
		PaymentForward: paymentForward,
	}
	if err = service.Dash.ProcessPayment(&opts, qm.l.With("user", msg.UserName)); err != nil {
		logger.Errorw("failed to process dash payment", "error", err.Error())
		d.Ack(false)
		return
	}
	// during processing, the user may have sent additional payments so need to re-grab them
	paymentForward, err = service.Dash.C.GetPaymentForwardByID(msg.PaymentForwardID)
	if err != nil {
		logger.Errorw("failed to get payment forward by id", "error", err.Error())
		d.Ack(false)
		return
	}
	if len(paymentForward.ProcessedTxs) == 0 {
		logger.Errorw("no processed transactions detected", "error", err.Error())
		d.Ack(false)
		return
	}
	if _, err = service.PM.ConfirmPayment(payment.TxHash); err != nil {
		logger.Errorw("failed to confirm payment", "error", err.Error())
		d.Ack(false)
		return
	}
	if _, err = service.UM.AddCredits(msg.UserName, payment.USDValue); err != nil {
		logger.Errorw("failed to add credits to user", "error", err.Error())
		d.Ack(false)
		return
	}
	logger.Infow("successfully confirmed payment", "credits", payment.USDValue)
	user, err := um.FindByUserName(payment.UserName)
	if err != nil {
		logger.Errorw("failed to find email for user", "error", err.Error())
		d.Ack(false)
		return
	}
	if !user.EmailEnabled {
		logger.Warnw("user has not activated their email and won't receive notifications")
		d.Ack(false)
		return
	}
	es := EmailSend{
		Subject:     "DASH Payment Confirmed",
		Content:     fmt.Sprintf("Your dash payment for %v credits has been confirmed", payment.USDValue),
		ContentType: "text/html",
		UserNames:   []string{payment.UserName},
		Emails:      []string{user.EmailAddress},
	}
	if err := qmEmail.PublishMessage(es); err != nil {
		logger.Errorw("failed to send payment confirmation email", "error", err.Error())
	}
	d.Ack(false)
	return
}
