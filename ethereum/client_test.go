package ethereum_test

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/RTradeLtd/Pay/ethereum"
	"github.com/RTradeLtd/config/v2"
)

var (
	// to prevent tx nonce issues we use two differrent keys
	_Buildkey = `{"address":"cdd092c76eb443cd945d9747a3f960f3e29686fc","crypto":{"cipher":"aes-128-ctr","ciphertext":"fc6607717861d1ec6bc77ed958315b57ea418a23fd5dc318697d98544f8b1f50","cipherparams":{"iv":"ba9d37b35e9330099288940415352484"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"c69d7288501031a6fc21dc1c52f4a52c91f0ae25159f530a09f8dd5e7dd244fb"},"mac":"b7394fe7d97b8b73ee12e98515f353d7c47adaa8364c37dc23a4bf006c101f73"},"id":"f19e5d1f-a1a1-493e-9cce-c8cf86d8a061","version":3}`
	_PrKey    = `{"address":"9ac4be338f0d7edcb64a45ee8268d2812468ef6a","crypto":{"cipher":"aes-128-ctr","ciphertext":"48afb646ff12ba574c34b5879bc2d6dc5b69a0c1ffd32b1e74a7e608a908ef11","cipherparams":{"iv":"44676e919bd70308689e8fd9815468cd"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"f567b75a65557b0626fa95644fcd5ddb63b4480daea3d8b9f3ce6f8b66745af9"},"mac":"be2d72f009264d0e6c79be3f6f19b0ae41df926d49f0eb9e4546470f38d3d523"},"id":"ecd3157e-47fe-4b65-a20d-1f1701d2def3","version":3}`
	pass      = "password123"
	cfgPath   = "../test/config.json"
	// randomized name for testing
	tld      = ".eth"
	testHash = "QmbWK7PAbLPgm43b8RNmAtfv3tYbjjSas8EafVbcgWkwUH"
	randName = strconv.FormatInt(time.Now().Unix(), 10)
)

func getKey() string {
	// if pr build, use pr specific key
	if os.Getenv("TRAVIS_PULL_REQUEST") != "false" {
		return _PrKey
	}
	// if branch build, use branch specific key
	return _Buildkey
}

func getName() string {
	var tldPad string
	if os.Getenv("TRAVIS_PULL_REQUEST") != "false" {
		tldPad = "pr"
	} else {
		tldPad = "br"
	}
	return randName + tldPad + tld
}
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
	if err = c.UnlockAccount(getKey(), pass); err != nil {
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
	if err := c.RegisterName(getName()); err != nil {
		t.Fatal(err)
	}
	if err := c.SetResolver(getName()); err != nil {
		t.Fatal(err)
	}
	if err := c.RegisterSubDomain("ipfstemporal", getName()); err != nil {
		t.Fatal(err)
	}
	if err := c.UpdateContentHash("ipfstemporal", getName(), testHash); err != nil {
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
	if err = c.UnlockAccount(getKey(), pass); err != nil {
		t.Fatal(err)
	}
}
