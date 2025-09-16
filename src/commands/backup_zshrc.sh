#!/bin/zsh

# 描述: 备份 zshrc 配置文件
# 补全:
#   help: 显示帮助信息

# 定义备份目录
BACKUP_DIR="$HOME/backups/zshrc"

# 创建备份目录（如果不存在）
mkdir -p "$BACKUP_DIR"

# 生成时间戳
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# 备份文件
BACKUP_FILE="$BACKUP_DIR/.zshrc.bak.$TIMESTAMP"
cp ~/.zshrc "$BACKUP_FILE"

echo "✅ .zshrc 已备份到 $BACKUP_FILE"
