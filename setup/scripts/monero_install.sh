#! /bin/bash

# used to build and install a monero daemon
VERSION="release-v0.13"

echo "[INFO] Updating systems"
sudo apt update -y
echo "[INFO] Upgrading systems"
sudo apt upgrade -y
echo "[INFO] Installing build tools"
sudo apt install git build-essential cmake pkg-config libboost-all-dev libssl-dev libzmq3-dev libunbound-dev libsodium-dev libunwind8-dev liblzma-dev libreadline6-dev libldns-dev libexpat1-dev doxygen graphviz libpgm-dev -y
git clone --recursive https://github.com/monero-project/monero
cd monero || exit
git submodule init
git submodule update
git checkout "$VERSION" 
make -j 2