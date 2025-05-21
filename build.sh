#!/bin/sh

docker buildx build --platform linux/arm64 -t retro-console-arm64 -f Dockerfile .
docker create --name extract retro-console-arm64
docker cp extract:/build/pak-store pak-store
docker cp extract:/usr/lib/libSDL2_gfx-1.0.so.0 ./lib/libSDL2_gfx-1.0.so.0
docker rm extract
