#! /bin/bash

# Script used to manage a monero node

IP=0.0.0.0
WALLET_DIR="wallets"

case "$1" in 
    monerod)
        monerod --rpc-bind-ip "$IP" --confirm-external-bind
        ;;
    wallet)
        monero-wallet-rpc --rpc-bind-ip "$IP" --disable-rpc-login --wallet-dir "$WALLET_DIR"
        ;;

esac