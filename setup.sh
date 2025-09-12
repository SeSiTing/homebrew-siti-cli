#!/bin/bash

# siti-cli 简化配置脚本
# 作者: siti
# 版本: 1.0.0

# 获取脚本所在目录的绝对路径
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="$SCRIPT_DIR/bin"

# 确保脚本有执行权限
chmod +x "$BIN_DIR/siti"

# 检测shell类型
SHELL_TYPE="$(basename "$SHELL")"
RC_FILE="$HOME/.zshrc"  # 默认使用zshrc

if [ "$SHELL_TYPE" = "bash" ]; then
  RC_FILE="$HOME/.bashrc"
fi

echo "正在配置 siti-cli..."

# 检查PATH配置
if grep -q "$BIN_DIR" "$RC_FILE"; then
  echo "siti-cli 已经在 $RC_FILE 中配置。"
else
  # 添加到PATH
  echo "" >> "$RC_FILE"
  echo "# === siti-cli 配置 ===" >> "$RC_FILE"
  echo "export PATH=\"$BIN_DIR:\$PATH\"" >> "$RC_FILE"
  echo "成功将 siti-cli 添加到 $RC_FILE"
fi

echo "配置完成！"
echo "请运行以下命令使配置生效:"
echo "  source $RC_FILE"
echo "然后你就可以使用 'siti' 命令了。例如:"
echo "  siti --help"
echo "  siti netcheck"
echo "  siti proxy on"