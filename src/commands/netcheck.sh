#!/bin/bash

# 描述: 检查网络连接状态
TARGETS=("baidu.com" "google.com" "github.com")

for TARGET in "${TARGETS[@]}"; do
  echo "🔍 ping $TARGET"
  ping -c 2 $TARGET
  echo ""
done
