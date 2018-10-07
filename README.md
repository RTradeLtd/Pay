# Temporal_Payment-ETH

Temporal's Ethereum Based Payment Processor 

The Ethereum payment processor consists of a gRPC API Server, and a queue runner. The gRPC API Server is used to generate signed messages to submit to the payment contract, while the queue runner is responsible for validating the transactions.

## Docker Images:

* Queue Runner `rtradetech/temporal-payment-runner`
* gRPC API Server `rtradetech/temporal-payment-gapi-server`
