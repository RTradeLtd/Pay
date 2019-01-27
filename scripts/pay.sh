#! /bin/bash

CONFIG_FILE="$1"
if [[ "$CONFIG_FILE" == "" ]]; then
    echo "please provide location to config file"
    echo "example: ./pay.sh /tmp/config.json"
    exit 1
fi

pay -config "$CONFIG_FILE" grpc server &
pay -config "$CONFIG_FILE" queue payment ethereum &