#!/bin/bash

# siti-cli 卸载脚本
# 作者: siti
# 版本: 1.0.0

echo "正在卸载 siti-cli..."

# 检测shell类型
SHELL_TYPE="$(basename "$SHELL")"
RC_FILE="$HOME/.zshrc"  # 默认使用zshrc

if [ "$SHELL_TYPE" = "bash" ]; then
  RC_FILE="$HOME/.bashrc"
fi

# 清理配置文件中的 siti-cli 配置
cleanup_config() {
  if [ -f "$RC_FILE" ]; then
    local temp_file=$(mktemp)
    
    # 移除 siti-cli 相关的配置块
    awk '
      /^# === siti-cli 配置 ===/ { in_siti_block=1; next }
      in_siti_block && /^# === [^s]|^$/ { in_siti_block=0 }
      /^# === siti-cli 补全配置 ===/ { in_completion_block=1; next }
      in_completion_block && /^# === [^s]|^$/ { in_completion_block=0 }
      !in_siti_block && !in_completion_block { print }
    ' "$RC_FILE" > "$temp_file"
    
    mv "$temp_file" "$RC_FILE"
    echo "已从 $RC_FILE 中移除 siti-cli 配置"
  fi
}

# 执行清理
cleanup_config

echo "siti-cli 卸载完成！"
echo "请运行以下命令使配置生效:"
echo "  source $RC_FILE"
echo ""
echo "注意: 本脚本只清理配置文件，不会删除 siti-cli 文件。"
echo "如需完全删除，请手动删除 siti-cli 目录。"
