# Rebuild vendored dependencies
.PHONY: vendor
vendor:
	@echo "=================== generating dependencies ==================="
	# Nuke vendor directory
	rm -rf vendor

	# Update standard dependencies
	dep ensure -v

	# Vendor ethereum - this step is required for some of the cgo components, as
	# dep doesn't seem to resolve them
	go get -u github.com/ethereum/go-ethereum
	cp -r $(GOPATH)/src/github.com/ethereum/go-ethereum/crypto/secp256k1/libsecp256k1 \
  	./vendor/github.com/ethereum/go-ethereum/crypto/secp256k1/
	@echo "=================== done ==================="