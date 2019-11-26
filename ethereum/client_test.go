package ethereum_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/RTradeLtd/Pay/ethereum"
	"github.com/RTradeLtd/config/v2"
)

var (
	key     = `{"address":"7e4a2359c745a982a54653128085eac69e446de1","crypto":{"cipher":"aes-128-ctr","ciphertext":"eea2004c17292a9e94217bf53efbc31ff4ae62f3dd57f0938ab61c949a565dc1","cipherparams":{"iv":"6f6a7a89b556604940ac87ab1e78cfd1"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"8088e943ac0f37c8b4d01592d8bee96468853b6f1f13ca64d201cd68e7dc7b12"},"mac":"f856d734705f35e2acf854a44eb40796518730bd835ecaec01d1f3e7a7037813"},"id":"99e2cd49-4b51-4f01-b34c-aaa0efd332c3","version":3}`
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
