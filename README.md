# Pay

Pay is a service used to facilitate payment for credits on Temporal. It is a seperate package to avoid licensing headaches between go-ethereum and Temporal.
We used a gRPC API Service to communicate between this service, and the Temporal API.

## Docker Images:

* Queue Runner `rtradetech/temporal-payment-runner` <- outdated
* gRPC API Server `rtradetech/temporal-payment-grpc-server`
