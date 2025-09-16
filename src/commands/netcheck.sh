#!/bin/bash

# 描述: 检查网络连接状态
# 补全:
#   help: 显示帮助信息

# 定义要检查的目标网站
TARGETS=("baidu.com" "google.com" "github.com")

# 对每个目标进行ping测试
for TARGET in "${TARGETS[@]}"; do
  echo "🔍 ping $TARGET"
  ping -c 2 "$TARGET"
  echo ""
done
