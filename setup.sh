#!/bin/bash

# siti-cli 简化配置脚本
# 作者: siti
# 版本: 1.0.0

# 获取脚本所在目录的绝对路径
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="$SCRIPT_DIR/bin"
COMPLETIONS_DIR="$SCRIPT_DIR/completions"

# 确保脚本有执行权限
chmod +x "$BIN_DIR/siti"

# 确保补全脚本有执行权限
if [ -d "$COMPLETIONS_DIR" ]; then
  chmod +x "$COMPLETIONS_DIR"/*.bash 2>/dev/null || true
  chmod +x "$COMPLETIONS_DIR"/_* 2>/dev/null || true
fi

# 检测shell类型
SHELL_TYPE="$(basename "$SHELL")"
RC_FILE="$HOME/.zshrc"  # 默认使用zshrc

if [ "$SHELL_TYPE" = "bash" ]; then
  RC_FILE="$HOME/.bashrc"
fi

echo "正在配置 siti-cli..."

# 检查并清理旧的 siti-cli 配置
cleanup_old_config() {
  local temp_file=$(mktemp)
  # 移除旧的 siti-cli 配置块
  awk '
    /^# === siti-cli 配置 ===/ { in_block=1; next }
    in_block && /^# === [^s]|^$/ { in_block=0 }
    !in_block { print }
  ' "$RC_FILE" > "$temp_file"
  mv "$temp_file" "$RC_FILE"
}

# 清理旧配置
cleanup_old_config

# 检查PATH配置
if grep -q "$BIN_DIR" "$RC_FILE"; then
  echo "siti-cli 已经在 $RC_FILE 中配置。"
else
  # 添加到PATH
  echo "" >> "$RC_FILE"
  echo "# === siti-cli 配置 ===" >> "$RC_FILE"
  echo "export PATH=\"$BIN_DIR:\$PATH\"" >> "$RC_FILE"
  echo "" >> "$RC_FILE"
  echo "# proxy 命令包装器" >> "$RC_FILE"
  echo "siti() {" >> "$RC_FILE"
  echo "  if [[ \"\$1\" == \"proxy\" && (\"\$2\" == \"on\" || \"\$2\" == \"off\") ]]; then" >> "$RC_FILE"
  echo "    eval \"\$(command siti \"\$@\")\"" >> "$RC_FILE"
  echo "  else" >> "$RC_FILE"
  echo "    command siti \"\$@\"" >> "$RC_FILE"
  echo "  fi" >> "$RC_FILE"
  echo "}" >> "$RC_FILE"
  echo "成功将 siti-cli 添加到 $RC_FILE"
fi

# 配置补全功能
if [ -d "$COMPLETIONS_DIR" ]; then
  echo "正在配置命令补全功能..."
  
  if [ "$SHELL_TYPE" = "bash" ]; then
    # Bash 补全配置
    BASH_COMPLETION_FILE="$COMPLETIONS_DIR/siti.bash"
    if [ -f "$BASH_COMPLETION_FILE" ]; then
      if ! grep -q "siti.bash" "$RC_FILE"; then
        echo "" >> "$RC_FILE"
        echo "# === siti-cli 补全配置 ===" >> "$RC_FILE"
        echo "source \"$BASH_COMPLETION_FILE\"" >> "$RC_FILE"
        echo "成功配置 Bash 补全功能"
      else
        echo "Bash 补全功能已配置"
      fi
    fi
  else
    # Zsh 补全配置
    ZSH_COMPLETION_FILE="$COMPLETIONS_DIR/_siti"
    if [ -f "$ZSH_COMPLETION_FILE" ]; then
      # 检查 fpath 配置
      if ! grep -q "$COMPLETIONS_DIR" "$RC_FILE"; then
        echo "" >> "$RC_FILE"
        echo "# === siti-cli 补全配置 ===" >> "$RC_FILE"
        echo "fpath=(\"$COMPLETIONS_DIR\" \$fpath)" >> "$RC_FILE"
        echo "autoload -U compinit && compinit" >> "$RC_FILE"
        echo "成功配置 Zsh 补全功能"
      else
        echo "Zsh 补全功能已配置"
      fi
    fi
  fi
fi

echo "配置完成！"
echo "请运行以下命令使配置生效:"
echo "  source $RC_FILE"
echo "然后你就可以使用 'siti' 命令了。例如:"
echo "  siti --help"
echo "  siti netcheck"
echo "  siti proxy on"
echo ""
echo "💡 提示: 现在支持命令补全功能！"
echo "   - 输入 'siti ' 然后按 Tab 键查看所有可用命令"
echo "   - 输入 'siti proxy ' 然后按 Tab 键查看子命令"
echo ""
echo "🔧 管理命令:"
echo "   - 重新安装: ./setup.sh"
echo "   - 卸载配置: ./uninstall.sh"
echo "   - 验证安装: siti --version"