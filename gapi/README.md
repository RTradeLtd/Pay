# gapi 

Gapi is a GRPC API used by Temporal to facilitate purchasing of credits via ETH and RTC. It allows us to seperate the go-ethereum logic from the core Temporal code base to avoid licensing head-aches, allowing us to license the Temporal payment system under a different license than the core Temporal code base. The `client` package, contains no go-ethereum code whatsoever, and is licensed under MIT, allowing it to be imported by Temporal without go-ethereum dependencies! Using this, we can then request, and receive signed payment messages and display them to the user, allowing them to submit the data to our payment smart contract for validation.

## Licensing

The following packages are licensed differently than Temporal_Payment-ETH, and are licensed under MIT:

* `request`
* `response`
* `service`
* `client`