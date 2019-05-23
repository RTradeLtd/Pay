package signer

import (
	"io/ioutil"
	"math/big"
	"os"
	"testing"

	"github.com/RTradeLtd/config/v2"
	"github.com/ethereum/go-ethereum/common"
)

var (
	key     = `{"address":"7e4a2359c745a982a54653128085eac69e446de1","crypto":{"cipher":"aes-128-ctr","ciphertext":"eea2004c17292a9e94217bf53efbc31ff4ae62f3dd57f0938ab61c949a565dc1","cipherparams":{"iv":"6f6a7a89b556604940ac87ab1e78cfd1"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"8088e943ac0f37c8b4d01592d8bee96468853b6f1f13ca64d201cd68e7dc7b12"},"mac":"f856d734705f35e2acf854a44eb40796518730bd835ecaec01d1f3e7a7037813"},"id":"99e2cd49-4b51-4f01-b34c-aaa0efd332c3","version":3}`
	pass    = "password123"
	cfgPath = "../test/config.json"
)

func TestSigner(t *testing.T) {
	defer os.Remove("key.txt")
	if err := ioutil.WriteFile("key.txt", []byte(key), os.FileMode(0644)); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Ethereum.Account.KeyFile = "key.txt"
	cfg.Ethereum.Account.KeyPass = pass

	signer, err := NewPaymentSigner(cfg)
	if err != nil {
		t.Fatal(err)
	}
	addr := common.HexToAddress("")
	if _, err := signer.GenerateSignedPaymentMessagePrefixed(
		addr, 0, big.NewInt(1), big.NewInt(1),
	); err != nil {
		t.Fatal(err)
	}
}
