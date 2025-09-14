#!/bin/zsh

# 描述: 备份 zshrc 配置文件
# 补全:
#   help: 显示帮助信息
BACKUP_DIR="$HOME/backups/zshrc"
mkdir -p $BACKUP_DIR
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
cp ~/.zshrc "$BACKUP_DIR/.zshrc.bak.$TIMESTAMP"
echo "✅ .zshrc 已备份到 $BACKUP_DIR/.zshrc.bak.$TIMESTAMP"
