package signer

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"

	"github.com/RTradeLtd/config"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	solsha3 "github.com/miguelmota/go-solidity-sha3"
)

// PaymentSigner is used to signed payment messages
// and holds the ecdsa private key we used
type PaymentSigner struct {
	Key *ecdsa.PrivateKey
}

// SignedMessage is the response to a message signing request
type SignedMessage struct {
	H             [32]byte       `json:"h"`
	R             [32]byte       `json:"r"`
	S             [32]byte       `json:"s"`
	V             uint8          `json:"v"`
	Address       common.Address `json:"address"`
	PaymentMethod uint8          `json:"payment_method"`
	PaymentNumber *big.Int       `json:"payment_number"`
	ChargeAmount  *big.Int       `json:"charge_amount"`
	Hash          []byte         `json:"hash"`
	Sig           []byte         `json:"sig"`
}

// GeneratePaymentSigner is used to generate our helper struct for signing payments
// keyFilePath is the path to a key as generated by geth
func GeneratePaymentSigner(cfg *config.TemporalConfig) (*PaymentSigner, error) {

	fileBytes, err := ioutil.ReadFile(
		cfg.Ethereum.Account.KeyFile,
	)
	if err != nil {
		return nil, err
	}
	pk, err := keystore.DecryptKey(
		fileBytes,
		cfg.Ethereum.Account.KeyPass,
	)
	if err != nil {
		return nil, err
	}
	return &PaymentSigner{Key: pk.PrivateKey}, nil
}

// GenerateSignedPaymentMessagePrefixed generates a signed payment message. The format is slightly different and involves
// generating the hash to sign first, prefixing with `"\x19Ethereum Signed Message:\n32"
func (ps *PaymentSigner) GenerateSignedPaymentMessagePrefixed(ethAddress common.Address, paymentMethod uint8, paymentNumber, chargeAmountInWei *big.Int) (*SignedMessage, error) {
	//  return keccak256(abi.encodePacked(msg.sender, _paymentNumber, _paymentMethod, _chargeAmountInWei));
	hashToSign := solsha3.SoliditySHA3(
		solsha3.Address(ethAddress),
		solsha3.Uint256(paymentNumber),
		solsha3.Uint8(paymentMethod),
		solsha3.Uint256(chargeAmountInWei),
	)
	hashPrefixed := solsha3.SoliditySHA3WithPrefix(hashToSign)
	sig, err := crypto.Sign(hashPrefixed, ps.Key)
	if err != nil {
		return nil, err
	}
	var h, r, s [32]byte
	for k := range hashPrefixed {
		h[k] = hashPrefixed[k]
	}
	if len(h) > 32 || len(h) < 32 {
		return nil, errors.New("failed to parse h")
	}
	for k := range sig[0:64] {
		if k < 32 {
			r[k] = sig[k]
		}
		if k >= 32 {
			s[k-32] = sig[k]
		}
	}
	if len(r) != len(s) && len(r) != 32 {
		return nil, errors.New("failed to parse R+S")
	}

	msg := &SignedMessage{
		H:             h,
		R:             r,
		S:             s,
		V:             uint8(sig[64]) + 27,
		Address:       ethAddress,
		PaymentMethod: paymentMethod,
		PaymentNumber: paymentNumber,
		ChargeAmount:  chargeAmountInWei,
		Hash:          hashPrefixed,
		Sig:           sig[0:64],
	}

	// Here we do an off-chain validation to ensure that when validated on-chain the transaction won't rever
	// however for some reason, the data isn't validating on-chain
	pub := ps.Key.PublicKey
	compressedKey := crypto.CompressPubkey(&pub)
	valid := crypto.VerifySignature(compressedKey, msg.Hash, msg.Sig)
	if !valid {
		return nil, errors.New("failed to validate signature off-chain")
	}
	fmt.Println("successfully validated signature")
	return msg, nil
}
