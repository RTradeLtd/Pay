#! /bin/bash

CONFIG_FILE="$1"

pay -config "$CONFIG_FILE" grpc server &
pay -config "$CONFIG_FILE" queue payment ethereum &