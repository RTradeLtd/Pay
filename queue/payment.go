package queue

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/RTradeLtd/Temporal/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// ProcessEthereumBasedPayment is used to process ethereum based payment messages
func (qm *QueueManager) ProcessEthereumBasedPayment(msgs <-chan amqp.Delivery, db *gorm.DB, cfg *config.TemporalConfig) error {
	//pm := models.NewPaymentManager(db)
	for d := range msgs {
		pc := PaymentCreation{}
		if err := json.Unmarshal(d.Body, &pc); err != nil {
			qm.Logger.Error("failed to unmarshal message")
			d.Ack(false)
			continue
		}
		switch pc.Type {
		case "eth":
			if err := qm.processEthPayment(pc, cfg); err != nil {
				qm.Logger.WithFields(log.Fields{
					"service": qm.Service,
					"error":   err.Error(),
				}).Error("failed to process message")
			}
			d.Ack(false)
			continue
		case "rtc":
			if err := qm.processRTCPayment(); err != nil {
				qm.Logger.WithFields(log.Fields{
					"service": qm.Service,
					"error":   err.Error(),
				}).Error("failed to process message")
			}
			d.Ack(false)
			continue
		}
	}
	return nil
}

func (qm *QueueManager) processRTCPayment() error {
	return nil
}

func (qm *QueueManager) processEthPayment(msg PaymentCreation, cfg *config.TemporalConfig) error {
	client, err := ethclient.Dial(cfg.Ethereum.Connection.INFURA.URL)
	if err != nil {
		return err
	}
	txHash := common.HexToHash(msg.TxHash)
	tx, pending, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		return err
	}
	rcpt := &types.Receipt{}
	if pending {
		rcpt, err = bind.WaitMined(context.Background(), client, tx)
	} else {
		rcpt, err = client.TransactionReceipt(context.Background(), tx.Hash())
	}
	if err != nil {
		return err
	}
	if rcpt.Status != uint64(1) {
		return errors.New("tx status is not 1")
	}
	return nil
}
