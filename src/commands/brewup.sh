#!/bin/bash

# 描述: Homebrew 一键升级全流程
# 补全:
#   --interactive: 逐步确认执行
#   -i: --interactive 的简写
# 用法:
#   siti brewup              自动执行全流程
#   siti brewup -i           逐步确认每个步骤
#   siti brewup --interactive 逐步确认每个步骤

# 检查是否为交互模式
INTERACTIVE=false
if [[ "$1" == "--interactive" ]] || [[ "$1" == "-i" ]]; then
  INTERACTIVE=true
fi

# 错误收集数组
ERRORS=()

# 步骤标题
TOTAL_STEPS=5

# 辅助函数：显示步骤标题
step_header() {
  local step_num="$1"
  local title="$2"
  echo ""
  echo "🔄 [$step_num/$TOTAL_STEPS] $title"
}

# 辅助函数：显示步骤完成
step_done() {
  local step_num="$1"
  echo "✅ [$step_num/$TOTAL_STEPS] 完成"
}

# 辅助函数：记录错误
record_error() {
  local error_msg="$1"
  ERRORS+=("$error_msg")
  echo "❌ $error_msg" >&2
}

# 辅助函数：交互确认
confirm_step() {
  local step_title="$1"
  if [[ "$INTERACTIVE" == "true" ]]; then
    echo ""
    read -p "❓ 执行此步骤? [Y/n] " confirm
    if [[ "$confirm" =~ ^[Nn] ]]; then
      echo "⏭️  跳过: $step_title"
      return 1
    fi
  fi
  return 0
}

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🍺 Homebrew 一键升级全流程"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 记录开始时间
START_TIME=$(date +%s)

# ============================================
# 步骤 1: 更新 Homebrew 自身
# ============================================
step_header 1 "更新 Homebrew 自身"
if confirm_step "brew update"; then
  if brew update; then
    step_done 1
  else
    record_error "步骤 1 失败: brew update"
  fi
else
  step_done 1
fi

# ============================================
# 步骤 2: 升级所有 formula
# ============================================
step_header 2 "升级所有 formula"
if confirm_step "brew upgrade (formula)"; then
  # 记录升级前的包数量
  BEFORE_FORMULA=$(brew list --formula | wc -l | tr -d ' ')
  
  if brew upgrade; then
    AFTER_FORMULA=$(brew list --formula | wc -l | tr -d ' ')
    echo "📦 Formula: $BEFORE_FORMULA 个 → $AFTER_FORMULA 个"
    step_done 2
  else
    record_error "步骤 2 失败: brew upgrade"
  fi
else
  step_done 2
fi

# ============================================
# 步骤 3: 升级所有 cask（包括自动更新的）
# ============================================
step_header 3 "升级所有 cask（包括自动更新的应用）"
if confirm_step "brew upgrade --cask --greedy"; then
  # 记录升级前的 cask 数量
  BEFORE_CASK=$(brew list --cask | wc -l | tr -d ' ')
  
  if brew upgrade --cask --greedy; then
    AFTER_CASK=$(brew list --cask | wc -l | tr -d ' ')
    echo "📦 Cask: $BEFORE_CASK 个 → $AFTER_CASK 个"
    step_done 3
  else
    record_error "步骤 3 失败: brew upgrade --cask --greedy"
  fi
else
  step_done 3
fi

# ============================================
# 步骤 4: 删除无用依赖
# ============================================
step_header 4 "删除无用依赖"
if confirm_step "brew autoremove"; then
  if brew autoremove; then
    step_done 4
  else
    record_error "步骤 4 失败: brew autoremove"
  fi
else
  step_done 4
fi

# ============================================
# 步骤 5: 清理缓存和旧版本
# ============================================
step_header 5 "清理缓存和旧版本"
if confirm_step "brew cleanup --prune=all"; then
  # 获取清理前的缓存大小
  if [ -d "$(brew --cache)" ]; then
    BEFORE_SIZE=$(du -sh "$(brew --cache)" 2>/dev/null | awk '{print $1}')
  else
    BEFORE_SIZE="0B"
  fi
  
  if brew cleanup --prune=all; then
    # 获取清理后的缓存大小
    if [ -d "$(brew --cache)" ]; then
      AFTER_SIZE=$(du -sh "$(brew --cache)" 2>/dev/null | awk '{print $1}')
    else
      AFTER_SIZE="0B"
    fi
    
    echo "💾 缓存大小: $BEFORE_SIZE → $AFTER_SIZE"
    step_done 5
  else
    record_error "步骤 5 失败: brew cleanup --prune=all"
  fi
else
  step_done 5
fi

# ============================================
# 最终汇总
# ============================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 计算总耗时
END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))
MINUTES=$((DURATION / 60))
SECONDS=$((DURATION % 60))

echo "⏱️  总耗时: ${MINUTES}分${SECONDS}秒"

# 显示错误汇总
if [ ${#ERRORS[@]} -gt 0 ]; then
  echo ""
  echo "⚠️  执行过程中遇到 ${#ERRORS[@]} 个错误:"
  for err in "${ERRORS[@]}"; do
    echo "  • $err"
  done
  echo ""
  echo "❌ 升级流程完成，但存在错误"
  exit 1
else
  echo ""
  echo "✅ 所有步骤执行成功！"
  exit 0
fi
