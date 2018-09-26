# Build temporal if binary is not already present
temporal-payment-eth:
	@make cli

# Build Temporal
.PHONY: cli
cli:
	@echo "===================  building Temporal CLI  ==================="
	rm -f temporal
	go build -ldflags "-X main.Version=$(TEMPORALVERSION)" ./cmd/temporal
	@echo "===================          done           ==================="

# Rebuild vendored dependencies
.PHONY: vendor
vendor:
	@echo "=================== generating dependencies ==================="
	# Nuke vendor directory
	rm -rf vendor
	# Update standard dependencies
	dep ensure -v
	# Generate IPFS dependencies
	rm -rf vendor/github.com/ipfs/go-ipfs
	git clone https://github.com/ipfs/go-ipfs.git vendor/github.com/ipfs/go-ipfs
	( cd vendor/github.com/ipfs/go-ipfs ; git checkout $(IPFSVERSION) ; gx install --local --nofancy )
	mv vendor/github.com/ipfs/go-ipfs/vendor/* vendor
	# Vendor ethereum - this step is required for some of the cgo components, as
	# dep doesn't seem to resolve them
	go get -u github.com/ethereum/go-ethereum
	cp -r $(GOPATH)/src/github.com/ethereum/go-ethereum/crypto/secp256k1/libsecp256k1 \
  	./vendor/github.com/ethereum/go-ethereum/crypto/secp256k1/
	# Remove problematic dependencies
	find . -name test-vectors -type d -exec rm -r {} +
	@echo "=================== done ==================="