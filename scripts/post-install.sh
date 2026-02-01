#!/bin/bash

# siti-cli 安装后初始化
# 注意：禁用 set -e，避免权限问题导致整个脚本失败
# set -e

SITI_DIR="$HOME/.siti"

echo "正在初始化 siti-cli..."

# 创建目录
mkdir -p "$SITI_DIR"/{commands,logs,cache,config}

# 创建配置文件
if [ ! -f "$SITI_DIR/config/siti.conf" ]; then
  cat > "$SITI_DIR/config/siti.conf" << 'EOF'
# siti-cli 配置文件
LOG_LEVEL="info"
LOG_FILE="$HOME/.siti/logs/siti.log"
CACHE_DIR="$HOME/.siti/cache"
USER_COMMANDS_DIR="$HOME/.siti/commands"
EOF
fi

# 配置 shell 补全（尝试写入，失败则静默跳过）
SHELL_RC="$HOME/.zshrc"
[ "$(basename "$SHELL")" = "bash" ] && SHELL_RC="$HOME/.bashrc"

if ! grep -q "siti-cli completion" "$SHELL_RC" 2>/dev/null; then
  if cat >> "$SHELL_RC" 2>/dev/null << 'EOF'

# siti-cli completion
if command -v siti >/dev/null 2>&1; then
  if [ -f /opt/homebrew/share/zsh/site-functions/_siti ]; then
    fpath=(/opt/homebrew/share/zsh/site-functions $fpath)
  elif [ -f /usr/local/share/zsh/site-functions/_siti ]; then
    fpath=(/usr/local/share/zsh/site-functions $fpath)
  fi
fi
EOF
  then
    echo "✅ Shell 补全配置已添加"
  else
    echo "⚠️  无法写入 $SHELL_RC（可能是权限问题），请手动配置补全" >&2
  fi
else
  echo "✅ Shell 补全配置已存在"
fi

# 创建示例命令
if [ ! -f "$SITI_DIR/commands/hello.sh" ]; then
  cat > "$SITI_DIR/commands/hello.sh" << 'EOF'
#!/bin/bash
# 描述: 示例用户自定义命令
name="${1:-World}"
echo "Hello, $name! 这是一个用户自定义命令示例。"
EOF
  chmod +x "$SITI_DIR/commands/hello.sh"
fi

# 自动安装 shell wrapper（尝试写入，失败则静默跳过）
WRAPPER_MARKER="# siti shell wrapper - auto-generated"
if ! grep -q "$WRAPPER_MARKER" "$SHELL_RC" 2>/dev/null; then
  if cat >> "$SHELL_RC" 2>/dev/null << 'WRAPPER_EOF'

# siti shell wrapper - auto-generated
# 使需要修改环境变量的命令（如 proxy、ai）在当前终端立即生效
siti() {
  local output
  local exit_code
  output=$(command siti "$@" 2>&1)
  exit_code=$?
  if [ $exit_code -eq 10 ]; then
    eval "$output"
    return 0
  else
    echo "$output"
    return $exit_code
  fi
}
WRAPPER_EOF
  then
    echo "✅ Shell wrapper 已安装（siti ai、siti proxy 等命令将在当前终端生效）"
  else
    echo "⚠️  无法写入 $SHELL_RC（可能是权限问题），请手动配置 wrapper" >&2
  fi
else
  echo "✅ Shell wrapper 已存在，跳过安装"
fi

echo "✅ siti-cli 初始化完成！"
echo ""
echo "如果自动配置失败，请手动运行："
echo "  eval \"\$(siti init zsh)\" >> ~/.zshrc"
echo "  source ~/.zshrc"
echo ""
echo "运行 'siti --help' 查看所有命令"
