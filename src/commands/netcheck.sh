#!/bin/bash

# 描述: 检查网络连接状态
# 补全:
#   help: 显示帮助信息
TARGETS=("baidu.com" "google.com" "github.com")

for TARGET in "${TARGETS[@]}"; do
  echo "🔍 ping $TARGET"
  ping -c 2 $TARGET
  echo ""
done
