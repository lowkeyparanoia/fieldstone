#!/usr/bin/env bash
# Build the slugify WASM plugin. Requires:
#   rustup target add wasm32-unknown-unknown
set -euo pipefail
cd "$(dirname "$0")"

rustc \
  --edition 2021 \
  --target wasm32-unknown-unknown \
  --crate-type cdylib \
  -C panic=abort \
  -C opt-level=s \
  -C lto=fat \
  -o slugify.wasm \
  src/lib.rs

echo "built: $(pwd)/slugify.wasm ($(wc -c < slugify.wasm) bytes)"
