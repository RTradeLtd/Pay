package service

import (
	"context"
	"errors"

	"github.com/RTradeLtd/Pay/bch"
	"github.com/RTradeLtd/Pay/dash"
	"github.com/RTradeLtd/Pay/ethereum"
	"github.com/RTradeLtd/config/v2"
	"github.com/RTradeLtd/database/models"
	"github.com/RTradeLtd/database/v2"
)

// PaymentService is our service which handles payment management
type PaymentService struct {
	Client *ethereum.Client
	Dash   *dash.DashClient
	BCH    *bch.Client
	PM     *models.PaymentManager
	UM     *models.UserManager
}

// Opts is used to configure our payment service
type Opts struct {
	DashEnabled        bool
	EthereumEnabled    bool
	BitcoinCashEnabled bool
	BCHURL             string
	DevMode            bool
}

// NewPaymentService is used to generate our payment service
func NewPaymentService(ctx context.Context, cfg *config.TemporalConfig, opts *Opts, connectionType string) (*PaymentService, error) {
	dbm, err := database.New(cfg, database.Options{LogMode: true, SSLModeDisable: false})
	if err != nil {
		return nil, err
	}
	pm := models.NewPaymentManager(dbm.DB)
	um := models.NewUserManager(dbm.DB)
	ps := &PaymentService{PM: pm, UM: um}
	if opts.EthereumEnabled {
		ethClient, err := ethereum.NewClient(cfg, connectionType)
		if err != nil {
			return nil, err
		}
		ps.Client = ethClient
	}
	if opts.DashEnabled {
		dashClient, err := dash.GenerateDashClient(cfg)
		if err != nil {
			return nil, err
		}
		ps.Dash = dashClient
	}
	if opts.BitcoinCashEnabled {
		// TODO: move this to config
		if opts.BCHURL == "" {
			return nil, errors.New("bch url not specified")
		}
		bchClient, err := bch.NewClient(ctx, cfg, opts.DevMode)
		if err != nil {
			return nil, err
		}
		ps.BCH = bchClient
	}
	return ps, nil
}
