#!/bin/sh
PAK_DIR="$(dirname "$0")"
cd "$PAK_DIR" || exit 1

./pak-store
