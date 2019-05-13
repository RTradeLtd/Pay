# Build pay binary
pay:
	@make cli

# Build pay binary
.PHONY: cli
cli:
	@echo "===================  building Temporal CLI  ==================="
	rm -f pay
	go build -ldflags "-X main.Version=$(TEMPORALVERSION)" ./cmd/pay
	@echo "===================          done           ==================="

# Rebuild vendored dependencies
.PHONY: vendor
vendor:
	@echo "=================== generating dependencies ==================="
	GO111MODULE=on go mod vendor
	@echo "===================          done           ==================="


# Rebuild generate code
COUNTERFEITER=go run github.com/maxbrunsfeld/counterfeiter/v6
.PHONY: gen
gen:
	@echo "===================    regenerating code    ==================="
	$(COUNTERFEITER) -o ./mocks/bch.mock.go \
		github.com/gcash/bchd/bchrpc/pb.BchrpcClient
	@echo "===================          done           ==================="