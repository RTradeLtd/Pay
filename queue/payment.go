package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/RTradeLtd/Pay/dash"
	"github.com/RTradeLtd/Pay/service"
	"github.com/streadway/amqp"
)

// ProcessETHPayment is used to process ethereum and rtc based payments
func (qm *Manager) ProcessETHPayment(ctx context.Context, wg *sync.WaitGroup, msgs <-chan amqp.Delivery) error {
	service, err := service.GeneratePaymentService(qm.cfg, &service.Opts{EthereumEnabled: true, DashEnabled: false}, "rpc")
	if err != nil {
		return err
	}
	qm.l.Info("processing payment confirmations")
	for {
		select {
		case d := <-msgs:
			wg.Add(1)
			go qm.processETHPayment(d, wg, service)
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

func (qm *Manager) processETHPayment(d amqp.Delivery, wg *sync.WaitGroup, service *service.PaymentService) {
	defer wg.Done()
	qm.l.Info("new ethereum based payment message received")
	pc := EthPaymentConfirmation{}
	if err := json.Unmarshal(d.Body, &pc); err != nil {
		qm.l.Error("failed to unmarshal message")
		d.Ack(false)
		return
	}
	payment, err := service.PM.FindPaymentByNumber(pc.UserName, pc.PaymentNumber)
	if err != nil {
		qm.l.Error("failed to find payment message", "error", err.Error())
		d.Ack(false)
		return
	}
	switch payment.Blockchain {
	case "ethereum":
		if err = service.Client.ProcessPaymentTx(payment.TxHash); err != nil {
			qm.l.Error("failed to validate ethereum payment", "error", err.Error())
			d.Ack(false)
			return
		}
	default:
		qm.l.Error("invalid blockchain for crypto payments", "error", err.Error())
		d.Ack(false)
		return
	}
	if _, err = service.PM.ConfirmPayment(payment.TxHash); err != nil {
		qm.l.Error("failed to confirm payment in database", "error", err.Error())
		d.Ack(false)
		return
	}
	// grant credits to the user
	if _, err = service.UM.AddCredits(pc.UserName, payment.USDValue); err != nil {
		qm.l.Error("failed to add credits for user", "error", err.Error())
		d.Ack(false)
		return
	}
	qm.l.Info("successfully confirmed payment")
	d.Ack(false)
	return
}

// ProcessDASHPayment is used to process dash based payments
func (qm *Manager) ProcessDASHPayment(ctx context.Context, wg *sync.WaitGroup, msgs <-chan amqp.Delivery) error {
	service, err := service.GeneratePaymentService(qm.cfg, &service.Opts{EthereumEnabled: false, DashEnabled: true}, "rpc")
	if err != nil {
		return err
	}
	qm.l.Info("processing dash payment confirmations")
	for {
		select {
		case d := <-msgs:
			wg.Add(1)
			go qm.processDashPaymentConfirmation(d, wg, service)
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

func (qm *Manager) processDashPaymentConfirmation(d amqp.Delivery, wg *sync.WaitGroup, service *service.PaymentService) {
	defer wg.Done()
	qm.l.Info("new dash payment message received")
	msg := DashPaymentConfirmation{}
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		qm.l.Error("failed to unmarshal message", "error", err.Error())
		d.Ack(false)
		return
	}
	paymentForward, err := service.Dash.C.GetPaymentForwardByID(msg.PaymentForwardID)
	if err != nil {
		qm.l.Error("failed to get payment forward by id", "error", err.Error())
		d.Ack(false)
		return
	}
	payment, err := service.PM.FindPaymentByNumber(msg.UserName, msg.PaymentNumber)
	if err != nil {
		qm.l.Error("failed to search for payment by number", "error", err.Error())
		d.Ack(false)
		return
	}
	opts := dash.ProcessPaymentOpts{
		Number:         payment.Number,
		ChargeAmount:   payment.ChargeAmount,
		PaymentForward: paymentForward,
	}
	if err = service.Dash.ProcessPayment(&opts); err != nil {
		qm.l.Error("failed to process dash payment", "error", err.Error())
		d.Ack(false)
		return
	}
	// during processing, the user may have sent additional payments so need to re-grab them
	paymentForward, err = service.Dash.C.GetPaymentForwardByID(msg.PaymentForwardID)
	if err != nil {
		qm.l.Error("failed to get payment forward by id", "error", err.Error())
		d.Ack(false)
		return
	}
	if len(paymentForward.ProcessedTxs) == 0 {
		qm.l.Error("no processed transactions detected", "error", err.Error())
		d.Ack(false)
		return
	}
	if _, err = service.PM.ConfirmPayment(fmt.Sprintf("%s-%v", msg.UserName, msg.PaymentNumber)); err != nil {
		qm.l.Error("failed to confirm payment", "error", err.Error())
		d.Ack(false)
		return
	}
	if _, err = service.UM.AddCredits(msg.UserName, payment.USDValue); err != nil {
		qm.l.Error("failed to add credits to user", "error", err.Error())
		d.Ack(false)
		return
	}
	qm.l.Info("granted credits")
	d.Ack(false)
	return
}
