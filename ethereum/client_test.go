package ethereum_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/RTradeLtd/Pay/ethereum"
	"github.com/RTradeLtd/config/v2"
)

var (
	key     = `{"address":"cdd092c76eb443cd945d9747a3f960f3e29686fc","crypto":{"cipher":"aes-128-ctr","ciphertext":"fc6607717861d1ec6bc77ed958315b57ea418a23fd5dc318697d98544f8b1f50","cipherparams":{"iv":"ba9d37b35e9330099288940415352484"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"c69d7288501031a6fc21dc1c52f4a52c91f0ae25159f530a09f8dd5e7dd244fb"},"mac":"b7394fe7d97b8b73ee12e98515f353d7c47adaa8364c37dc23a4bf006c101f73"},"id":"f19e5d1f-a1a1-493e-9cce-c8cf86d8a061","version":3}`
	pass    = "password123"
	cfgPath = "../test/config.json"
	// randomized name for testing
	testName = strconv.FormatInt(time.Now().Unix(), 10) + ".eth"
	testHash = "QmbWK7PAbLPgm43b8RNmAtfv3tYbjjSas8EafVbcgWkwUH"
)

func TestEth_NewClient(t *testing.T) {
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ethereum.NewClient(cfg, "infura"); err != nil {
		t.Fatal(err)
	}
}

func TestENS(t *testing.T) {
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	c, err := ethereum.NewClient(cfg, "infura")
	if err != nil {
		t.Fatal(err)
	}
	if err = c.UnlockAccount(key, pass); err != nil {
		t.Fatal(err)
	}
	type args struct {
		sub, parent string
	}
	tests := []struct {
		name     string
		args     args
		wantName string
	}{
		{"1", args{"hello.", "world"}, "hello.world"},
		{"2", args{"hello", ".world"}, "hello.world"},
		{"3", args{"hello.", ".world"}, "hello.world"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if name := c.GetCombinedName(tt.args.sub, tt.args.parent); name != tt.wantName {
				t.Fatalf(
					"GetCombinedName() returned %s, wantName = %s",
					name, tt.wantName,
				)
			}
		})
	}
	if err := c.RegisterName(testName); err != nil {
		t.Fatal(err)
	}
	if err := c.SetResolver(testName); err != nil {
		t.Fatal(err)
	}
	if err := c.RegisterSubDomain("ipfstemporal", testName); err != nil {
		t.Fatal(err)
	}
	if err := c.UpdateContentHash("ipfstemporal", testName, testHash); err != nil {
		t.Fatal(err)
	}
}

func TestEth_UnlockAccount(t *testing.T) {
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	c, err := ethereum.NewClient(cfg, "infura")
	if err != nil {
		t.Fatal(err)
	}
	if err = c.UnlockAccount(key, pass); err != nil {
		t.Fatal(err)
	}
}
