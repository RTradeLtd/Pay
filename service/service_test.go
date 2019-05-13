package service

import (
	"testing"

	"github.com/RTradeLtd/config"
)

var (
	cfgPath = "../test/config.json"
)

func TestNewPaymetService(t *testing.T) {
	t.Skip("skipping integration test")
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewPaymentService(cfg, &Opts{true, true}, "infura"); err != nil {
		t.Fatal(err)
	}
}
