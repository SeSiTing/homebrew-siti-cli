#!/bin/zsh
BACKUP_DIR="$HOME/backups/zshrc"
mkdir -p $BACKUP_DIR
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
cp ~/.zshrc "$BACKUP_DIR/.zshrc.bak.$TIMESTAMP"
echo "✅ .zshrc 已备份到 $BACKUP_DIR/.zshrc.bak.$TIMESTAMP"
