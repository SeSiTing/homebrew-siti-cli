#!/bin/bash

# 描述: 卸载 siti-cli（仅独立安装；Homebrew 请用 brew uninstall）

# 用法:
#   siti uninstall -y           - 确认卸载（删除所有文件和配置）
#   siti uninstall --dry-run    - 预览将删除的内容（不实际删除）
#   siti uninstall --help       - 显示帮助信息
#
# 参数:
#   -y, --yes        跳过确认，直接卸载
#   --dry-run        仅显示将删除的内容，不实际执行
#   -h, --help       显示帮助信息
#
# 说明:
#   Homebrew 安装请使用: brew uninstall siti-cli
#   独立安装将删除 ~/.siti-cli 及 .zshrc 中的相关配置

set -e

SHELL_RC="$HOME/.zshrc"
[ "$(basename "$SHELL")" = "bash" ] && SHELL_RC="$HOME/.bashrc"

show_help() {
  echo "siti uninstall - 卸载 siti-cli（仅独立安装）"
  echo ""
  echo "用法:"
  echo "  siti uninstall -y           确认卸载（删除所有文件和配置）"
  echo "  siti uninstall --dry-run    预览将删除的内容（不实际删除）"
  echo "  siti uninstall --help       显示帮助信息"
  echo ""
  echo "参数:"
  echo "  -y, --yes      跳过确认，直接卸载"
  echo "  --dry-run      仅显示将删除的内容，不实际执行"
  echo "  -h, --help      显示帮助信息"
  echo ""
  echo "说明:"
  echo "  Homebrew 安装请使用: brew uninstall siti-cli"
  echo "  独立安装将删除 ~/.siti-cli 及 .zshrc 中的相关配置"
}

echo "🗑️  siti-cli 卸载"
echo ""

# 检测安装方式
if [ "$INSTALL_METHOD" = "homebrew" ]; then
  echo "检测到 Homebrew 安装"
  echo ""
  echo "请使用以下命令卸载："
  echo "  brew uninstall siti-cli"
  echo ""
  echo "Homebrew 会自动清理 shell 配置（wrapper、补全）。"
  echo "用户数据目录 ~/.siti-cli 会保留，如需删除请手动执行: rm -rf ~/.siti-cli"
  exit 0
fi

# 独立安装：解析参数
DRY_RUN=false
SKIP_CONFIRM=false

while [[ $# -gt 0 ]]; do
  case $1 in
    -y|--yes)
      SKIP_CONFIRM=true
      shift
      ;;
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    -h|--help|help)
      show_help
      exit 0
      ;;
    *)
      echo "未知参数: $1" >&2
      echo "用法: siti uninstall [-y|--yes] [--dry-run] [-h|--help]" >&2
      exit 1
      ;;
  esac
done

# 未检测到独立安装
if [ ! -d "$HOME/.siti-cli" ] && [ ! -L "$HOME/.local/bin/siti" ]; then
  echo "未检测到独立安装的 siti-cli（~/.siti-cli 或 ~/.local/bin/siti 不存在）"
  exit 0
fi

echo "检测到独立安装（~/.siti-cli）"
echo ""
echo "将删除以下内容："
echo ""
if [ -L "$HOME/.local/bin/siti" ]; then
  echo "  • 符号链接: ~/.local/bin/siti"
fi
if [ -d "$HOME/.siti-cli" ]; then
  echo "  • 安装目录: ~/.siti-cli"
  du -sh "$HOME/.siti-cli" 2>/dev/null | awk '{print "    大小: " $1}' || true
fi
if grep -q "# siti shell wrapper" "$SHELL_RC" 2>/dev/null; then
  echo "  • Shell 配置: wrapper、补全、PATH ($SHELL_RC)"
fi
echo ""

# Dry-run 模式：仅预览
if [ "$DRY_RUN" = true ]; then
  echo "ℹ️  预览模式（--dry-run），不会实际删除"
  exit 0
fi

# 需要 -y 确认
if [ "$SKIP_CONFIRM" != true ]; then
  echo "请使用 -y 或 --yes 标志确认卸载："
  echo "  siti uninstall -y"
  echo ""
  echo "或使用 --dry-run 仅预览："
  echo "  siti uninstall --dry-run"
  exit 1
fi

# 执行卸载（已确认）
echo "正在卸载..."
echo ""

# 备份配置文件
if [ -f "$SHELL_RC" ]; then
  backup_file="${SHELL_RC}.backup.$(date +%Y%m%d%H%M%S)"
  cp "$SHELL_RC" "$backup_file"
  echo "✅ 已备份配置: $backup_file"
fi

# 便携式从 RC 中删除块：删除从 pattern1 到 pattern2 的行
remove_rc_block() {
  local pattern1="$1"
  local pattern2="$2"
  if ! grep -q "$pattern1" "$SHELL_RC" 2>/dev/null; then
    return 0
  fi
  local tmpfile="${SHELL_RC}.siti-uninstall.$$"
  if sed "/$pattern1/,/$pattern2/d" "$SHELL_RC" > "$tmpfile" 2>/dev/null && [ -s "$tmpfile" ]; then
    mv "$tmpfile" "$SHELL_RC"
    return 0
  fi
  rm -f "$tmpfile"
  return 1
}

# 删除 shell wrapper
if grep -q "# siti shell wrapper" "$SHELL_RC" 2>/dev/null; then
  if remove_rc_block "# siti shell wrapper - auto-generated" "^}$"; then
    echo "✅ 已删除 shell wrapper"
  fi
fi

# 删除补全配置
if grep -q "# siti-cli completion" "$SHELL_RC" 2>/dev/null; then
  if remove_rc_block "# siti-cli completion" "^fi$"; then
    echo "✅ 已删除补全配置"
  fi
fi

# 删除 PATH 配置（新标记：多行块）
if grep -q "# siti-cli PATH configuration - auto-generated" "$SHELL_RC" 2>/dev/null; then
  if remove_rc_block "# siti-cli PATH configuration - auto-generated" "^export PATH=.*local/bin"; then
    echo "✅ 已删除 PATH 配置"
  fi
fi
# 清理旧式 "# siti-cli" + export PATH 块
if grep -q "^# siti-cli$" "$SHELL_RC" 2>/dev/null; then
  if remove_rc_block "^# siti-cli$" "export PATH=.*"; then
    echo "✅ 已删除旧式 PATH 配置"
  fi
fi

# 删除符号链接
if [ -L "$HOME/.local/bin/siti" ]; then
  rm "$HOME/.local/bin/siti"
  echo "✅ 已删除符号链接 ~/.local/bin/siti"
fi

# 删除安装目录
if [ -d "$HOME/.siti-cli" ]; then
  rm -rf "$HOME/.siti-cli"
  echo "✅ 已删除安装目录 ~/.siti-cli"
fi

echo ""
echo "✅ siti-cli 卸载完成"
echo ""
echo "请运行以下命令使配置生效："
echo "  source $SHELL_RC"
echo ""
