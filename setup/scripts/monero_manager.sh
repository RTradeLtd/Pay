#! /bin/bash

# Script used to manage a monero node

IP=0.0.0.0
WALLET_DIR="wallets"

case "$1" in 
    monerod)
        monerod --rpc-bind-ip "$IP" --confirm-external-bind
        ;;
    wallet)
        monero-wallet-rpc --rpc-bind-ip 0.0.0.0  --disable-rpc-login --wallet-file ./wallets/wallet --daemon-address 192.168.1.236:18081 --rpc-bind-port 18083 --confirm-external-bind --password ""
        ;;

esac