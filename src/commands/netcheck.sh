#!/bin/bash

TARGETS=("baidu.com" "google.com" "github.com")

for TARGET in "${TARGETS[@]}"; do
  echo "🔍 ping $TARGET"
  ping -c 2 $TARGET
  echo ""
done
