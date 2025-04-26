#!/bin/zsh
env GOOS=linux GOARCH=arm64 go build -gcflags="all=-N -l" -o pak-store app/pak_store.go || exit
