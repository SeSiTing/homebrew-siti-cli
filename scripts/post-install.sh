#!/bin/bash

# siti-cli 安装后初始化
set -e

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

# 配置 shell 补全
SHELL_RC="$HOME/.zshrc"
[ "$(basename "$SHELL")" = "bash" ] && SHELL_RC="$HOME/.bashrc"

if ! grep -q "siti-cli completion" "$SHELL_RC" 2>/dev/null; then
  cat >> "$SHELL_RC" << 'EOF'

# siti-cli completion
if command -v siti >/dev/null 2>&1; then
  if [ -f /opt/homebrew/share/zsh/site-functions/_siti ]; then
    fpath=(/opt/homebrew/share/zsh/site-functions $fpath)
  elif [ -f /usr/local/share/zsh/site-functions/_siti ]; then
    fpath=(/usr/local/share/zsh/site-functions $fpath)
  fi
fi
EOF
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

echo "✅ siti-cli 初始化完成！"
echo "运行 'siti --help' 查看所有命令"
echo "运行 'siti hello' 测试自定义命令"
