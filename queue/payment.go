package queue

import (
	"encoding/json"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// ProcessEthereumBasedPayment is used to process ethereum based payment messages
func (qm *QueueManager) ProcessEthereumBasedPayment(msgs <-chan amqp.Delivery, db *gorm.DB) error {
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
			if err := qm.processEthPayment(); err != nil {
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

func (qm *QueueManager) processEthPayment() error {
	return nil
}
