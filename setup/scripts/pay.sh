#! /bin/bash

CONFIG_FILE="/home/rtrade/config.json"
MODE="$1"

case "$MODE" in 

    server)
        pay -config "$CONFIG_FILE" grpc server
        ;;
    queue-payment-eth)
        pay -config "$CONFIG_FILE" queue payment ethereum
        ;;
    *)
        echo "invalid invocation"
        echo "example: ./pay.sh <mode>"
        echo "mode: server, queue-payment-eth"
        exit 1
        ;;

esac