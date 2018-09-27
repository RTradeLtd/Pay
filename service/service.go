package service

import (
	"github.com/RTradeLtd/Temporal_Payment-ETH/ethereum"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
	"github.com/RTradeLtd/database/models"
)

// PaymentService is our service which handles payment management
type PaymentService struct {
	Client *ethereum.Client
	PM     *models.PaymentManager
}

// GeneratePaymentService is used to generate our payment service
func GeneratePaymentService(cfg *config.TemporalConfig, connectionType string) (*PaymentService, error) {
	dbm, err := database.Initialize(cfg, database.DatabaseOptions{LogMode: true})
	if err != nil {
		return nil, err
	}
	pm := models.NewPaymentManager(dbm.DB)
	client, err := ethereum.NewClient(cfg, "infura")
	if err != nil {
		return nil, err
	}
	return &PaymentService{Client: client, PM: pm}, nil
}
