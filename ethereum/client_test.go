package ethereum_test

import (
	"testing"

	"github.com/RTradeLtd/Temporal_Payment-ETH/ethereum"
	"github.com/RTradeLtd/config"
)

var (
	key  = `{"address":"7e4a2359c745a982a54653128085eac69e446de1","crypto":{"cipher":"aes-128-ctr","ciphertext":"eea2004c17292a9e94217bf53efbc31ff4ae62f3dd57f0938ab61c949a565dc1","cipherparams":{"iv":"6f6a7a89b556604940ac87ab1e78cfd1"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"8088e943ac0f37c8b4d01592d8bee96468853b6f1f13ca64d201cd68e7dc7b12"},"mac":"f856d734705f35e2acf854a44eb40796518730bd835ecaec01d1f3e7a7037813"},"id":"99e2cd49-4b51-4f01-b34c-aaa0efd332c3","version":3}`
	pass = "password123"
	cfgPath   = filepath.Join(os.Getenv("home"), "config.json")
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


func TestEth_ProcessEthPaymentTx(t *testing.T) {
	cfg, err := config.LoadConfig(cfgPath)
	if err != 
}