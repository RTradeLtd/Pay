#! /bin/bash

# Cross-compile Temporal using gox, injecting appropriate tags.
go get -u github.com/mitchellh/gox

RELEASE=$(git describe --tags)
TARGETS="linux/amd64"

mkdir -p release

gox -output="release/pay-$(git describe --tags)-{{.OS}}-{{.Arch}}" \
    -ldflags "-X main.Version=$RELEASE" \
    -osarch="$TARGETS" \
    ./cmd/pay


ls ./release/pay* > files

for i in $(cat files); do
    sha256sum "$i" > "$i.sha256"
done
