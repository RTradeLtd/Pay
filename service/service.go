package service

import (
	"github.com/RTradeLtd/Pay/dash"
	"github.com/RTradeLtd/Pay/ethereum"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
	"github.com/RTradeLtd/database/models"
)

// PaymentService is our service which handles payment management
type PaymentService struct {
	Client *ethereum.Client
	Dash   *dash.DashClient
	PM     *models.PaymentManager
	UM     *models.UserManager
}

// GeneratePaymentService is used to generate our payment service
func GeneratePaymentService(cfg *config.TemporalConfig) (*PaymentService, error) {
	dbm, err := database.Initialize(cfg, database.DatabaseOptions{LogMode: true})
	if err != nil {
		return nil, err
	}
	pm := models.NewPaymentManager(dbm.DB)
	um := models.NewUserManager(dbm.DB)
	ps := &PaymentService{PM: pm, UM: um}
	ethClient, err := ethereum.NewClient(cfg, "infura")
	if err != nil {
		return nil, err
	}
	ps.Client = ethClient
	dashClient, err := dash.GenerateDashClient(cfg)
	if err != nil {
		return nil, err
	}
	ps.Dash = dashClient
	return ps, nil
}
