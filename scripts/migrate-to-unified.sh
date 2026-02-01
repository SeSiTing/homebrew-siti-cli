#!/bin/bash

# 迁移脚本：将 ~/.siti 合并到 ~/.siti-cli
# 可由 post-install.sh 或 install.sh 调用，也可单独运行：
#   SITI_DIR="$HOME/.siti-cli" bash scripts/migrate-to-unified.sh

SITI_DIR="${SITI_DIR:-$HOME/.siti-cli}"

if [ ! -d "$HOME/.siti" ]; then
  exit 0
fi

echo "检测到旧版目录 ~/.siti，正在迁移到 $SITI_DIR..."
BACKUP_DIR="$HOME/.siti.backup.$(date +%Y%m%d%H%M%S)"

if ! cp -R "$HOME/.siti" "$BACKUP_DIR" 2>/dev/null; then
  echo "⚠️  无法备份 ~/.siti，跳过迁移（请手动迁移后删除 ~/.siti）" >&2
  exit 1
fi

echo "✅ 已备份到 $BACKUP_DIR"
mkdir -p "$SITI_DIR"/{commands,logs,cache,config}

for subdir in commands config logs cache; do
  if [ -d "$HOME/.siti/$subdir" ] && [ "$(ls -A "$HOME/.siti/$subdir" 2>/dev/null)" ]; then
    cp -R "$HOME/.siti/$subdir"/* "$SITI_DIR/$subdir/" 2>/dev/null || true
    echo "✅ 已迁移 $subdir/"
  fi
done

rm -rf "$HOME/.siti"
echo "✅ 旧目录已删除"
exit 0
