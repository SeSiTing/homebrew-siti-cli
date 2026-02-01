#!/bin/bash

# siti-cli 卸载清理脚本
set -e

SHELL_RC="$HOME/.zshrc"
[ "$(basename "$SHELL")" = "bash" ] && SHELL_RC="$HOME/.bashrc"

echo "正在清理 siti-cli 配置..."

# 删除 shell wrapper
if grep -q "# siti shell wrapper - auto-generated" "$SHELL_RC" 2>/dev/null; then
  # 备份配置文件
  cp "$SHELL_RC" "${SHELL_RC}.backup.$(date +%Y%m%d_%H%M%S)"
  
  # 删除 wrapper 块（从标记行到函数结束的 }）
  sed -i.tmp '/# siti shell wrapper - auto-generated/,/^}$/d' "$SHELL_RC"
  rm -f "${SHELL_RC}.tmp"
  
  echo "✅ 已删除 shell wrapper"
fi

# 删除补全配置
if grep -q "# siti-cli completion" "$SHELL_RC" 2>/dev/null; then
  sed -i.tmp '/# siti-cli completion/,/^fi$/d' "$SHELL_RC"
  rm -f "${SHELL_RC}.tmp"
  echo "✅ 已删除补全配置"
fi

echo ""
echo "✅ siti-cli 配置已清理"
echo "📁 用户数据保留在: ~/.siti/"
echo "   如需完全删除，请运行: rm -rf ~/.siti"
