#!/bin/bash

# 描述: 管理 AI API 配置切换
# 补全:
#   switch: 切换 AI 服务商
#   current: 显示当前配置
#   list: 列出所有服务商
#   test: 测试当前配置
#   unset: 清除环境变量（切换到 OAuth 登录模式）
# 用法:
#   siti ai switch <provider> [--persist]    切换到指定服务商（默认临时，加 --persist 持久化）
#   siti ai current                          显示当前配置
#   siti ai list                             列出所有服务商
#   siti ai test                             测试当前配置
#   siti ai unset [--persist]                清除环境变量（切换到 OAuth 登录模式）

ZSHRC="$HOME/.zshrc"

# 读取跳过列表（从环境变量或 zshrc）
get_skip_list() {
  local skip_list="${SITI_AI_SKIP:-}"
  if [ -z "$skip_list" ]; then
    skip_list=$(grep '^export SITI_AI_SKIP=' "$ZSHRC" 2>/dev/null | sed -E 's/.*="(.*)"/\1/')
  fi
  echo "$skip_list"
}

# 列出所有可用的 AI 服务商
list_providers() {
  echo "可用的 AI 服务商:"
  
  local skip_list
  skip_list=$(get_skip_list)
  
  # 从 ~/.zshrc 提取所有 *_BASE_URL（排除 ANTHROPIC_BASE_URL）
  grep -E '^export [A-Z0-9_]+_BASE_URL=' "$ZSHRC" 2>/dev/null | \
    grep -v 'ANTHROPIC_BASE_URL' | \
    grep -v 'SITI_AI_SKIP' | \
    while IFS= read -r line; do
      # 提取变量名和值
      provider=$(echo "$line" | sed -E 's/export ([A-Z0-9_]+)_BASE_URL=.*/\1/')
      
      # 检查是否在跳过列表中（逗号分隔）
      if [[ ",$skip_list," == *",$provider,"* ]]; then
        continue
      fi
      
      url=$(echo "$line" | sed -E 's/.*="(.*)"/\1/')
      
      # 转换为小写显示
      provider_lower=$(echo "$provider" | tr '[:upper:]' '[:lower:]')
      
      # 检查是否为当前使用的：
      # 1. 先看环境变量实际值是否匹配（当前 shell 已切换）
      # 2. 再看 ~/.zshrc 中的间接引用（持久化配置）
      local is_current=false
      if [ -n "$ANTHROPIC_BASE_URL" ] && [ "$ANTHROPIC_BASE_URL" = "$url" ]; then
        is_current=true
      elif grep -qE "^export ANTHROPIC_BASE_URL=\"\\\$${provider}_BASE_URL\"" "$ZSHRC" 2>/dev/null; then
        # zshrc 中有间接引用，且当前环境变量未覆盖时才标记
        if [ -z "$ANTHROPIC_BASE_URL" ]; then
          is_current=true
        fi
      fi

      if [ "$is_current" = true ]; then
        printf "  • %-15s %s ← 当前\n" "$provider_lower" "$url"
      else
        printf "  • %-15s %s\n" "$provider_lower" "$url"
      fi
    done
  
  exit 0
}

resolve_current_provider() {
  local url="${ANTHROPIC_BASE_URL:-}"
  [ -z "$url" ] && return

  local skip_list
  skip_list=$(get_skip_list)

  grep -E '^export [A-Z0-9_]+_BASE_URL=' "$ZSHRC" 2>/dev/null | \
    grep -v 'ANTHROPIC_BASE_URL' | grep -v 'SITI_AI_SKIP' | \
    while IFS= read -r line; do
      local p u
      p=$(echo "$line" | sed -E 's/export ([A-Z0-9_]+)_BASE_URL=.*/\1/')
      [[ ",$skip_list," == *",$p,"* ]] && continue
      u=$(echo "$line" | sed -E 's/.*="(.*)"/\1/')
      [ "$url" = "$u" ] && echo "$p" | tr '[:upper:]' '[:lower:]' && break
    done
}

show_current() {
  echo "当前 AI API 配置:"

  local provider
  if [ -n "$ANTHROPIC_BASE_URL" ]; then
    provider=$(resolve_current_provider)
  fi

  if [ -z "$provider" ]; then
    local base_url_line=$(grep '^export ANTHROPIC_BASE_URL=' "$ZSHRC" 2>/dev/null | tail -1)
    [ -z "$base_url_line" ] && echo "  ❌ 未配置" && exit 0
    provider=$(echo "$base_url_line" | sed -E 's/.*"\$([A-Z0-9_]+)_BASE_URL".*/\1/')
    [ -n "$provider" ] && provider=$(echo "$provider" | tr '[:upper:]' '[:lower:]')
  fi

  if [ -n "$provider" ]; then
    echo "  服务商: $provider"
    [ -n "$ANTHROPIC_BASE_URL" ] && echo "  BASE_URL: $ANTHROPIC_BASE_URL"
    if [ -n "$ANTHROPIC_AUTH_TOKEN" ]; then
      echo "  AUTH_TOKEN: ${ANTHROPIC_AUTH_TOKEN:0:20}..."
    fi
  else
    echo "  ❌ 未配置"
  fi
  exit 0
}

# 切换服务商
switch_provider() {
  local provider persist_flag
  for arg in "$@"; do
    case "$arg" in
      --persist|-p) persist_flag="--persist" ;;
      *) [ -z "$provider" ] && provider="$arg"
    esac
  done

  # 检测 shell wrapper 是否已配置
  if ! grep -q "# siti shell wrapper" "$ZSHRC" 2>/dev/null; then
    echo "⚠️  检测到 shell wrapper 未配置，切换后不会在当前终端生效" >&2
    echo "" >&2
    echo "请运行以下命令配置 shell wrapper（仅需一次）：" >&2
    echo "  eval \"\$(siti init zsh)\" >> ~/.zshrc" >&2
    echo "  source ~/.zshrc" >&2
    exit 1
  fi
  
  if [ -z "$provider" ]; then
    echo "❌ 请指定服务商名称" >&2
    echo "运行 'siti ai list' 查看可用服务商" >&2
    exit 1
  fi
  
  # 转换为大写
  local provider_upper=$(echo "$provider" | tr '[:lower:]' '[:upper:]')
  
  # 读取跳过列表
  local skip_list
  skip_list=$(get_skip_list)
  
  # 检查是否在跳过列表中
  if [[ ",$skip_list," == *",$provider_upper,"* ]]; then
    echo "❌ 服务商 '$provider' 在跳过列表中（SITI_AI_SKIP），不允许切换" >&2
    exit 1
  fi
  
  # 检查服务商是否存在
  if ! grep -q "^export ${provider_upper}_BASE_URL=" "$ZSHRC" 2>/dev/null; then
    echo "❌ 服务商 '$provider' 不存在" >&2
    echo "" >&2
    list_providers >&2
    exit 1
  fi
  
  # 决定 AUTH_TOKEN 引用（兜底到 DEFAULT_AUTH_TOKEN）
  local auth_token_ref
  if grep -q "^export ${provider_upper}_API_KEY=" "$ZSHRC" 2>/dev/null; then
    auth_token_ref="\$${provider_upper}_API_KEY"
  else
    # 兜底：使用 DEFAULT_AUTH_TOKEN
    auth_token_ref="\$DEFAULT_AUTH_TOKEN"
  fi

  # 处理 MODEL：检查是否有对应服务商的 MODEL 定义
  local model_ref=""
  local has_model_def=false

  if grep -q "^export ${provider_upper}_MODEL=" "$ZSHRC" 2>/dev/null; then
    has_model_def=true
    model_ref="\$${provider_upper}_MODEL"
  fi

  # 持久模式：修改 ~/.zshrc
  if [[ "$persist_flag" == "--persist" ]]; then
    # 备份 ~/.zshrc
    cp "$ZSHRC" "${ZSHRC}.backup.$(date +%Y%m%d_%H%M%S)"

    # 使用 sed 替换 ANTHROPIC_BASE_URL
    sed -i.tmp -E "s|^export ANTHROPIC_BASE_URL=.*|export ANTHROPIC_BASE_URL=\"\$${provider_upper}_BASE_URL\"|" "$ZSHRC"

    # 使用 sed 替换 ANTHROPIC_AUTH_TOKEN
    sed -i.tmp -E "s|^export ANTHROPIC_AUTH_TOKEN=.*|export ANTHROPIC_AUTH_TOKEN=\"${auth_token_ref}\"|" "$ZSHRC"

    # 处理 ANTHROPIC_MODEL：如果有对应 model 定义则设置，否则清除
    if [ "$has_model_def" = true ]; then
      # 如果已有 ANTHROPIC_MODEL 行则替换，否则追加
      if grep -q "^export ANTHROPIC_MODEL=" "$ZSHRC" 2>/dev/null; then
        sed -i.tmp -E "s|^export ANTHROPIC_MODEL=.*|export ANTHROPIC_MODEL=\"${model_ref}\"|" "$ZSHRC"
      else
        echo "export ANTHROPIC_MODEL=\"${model_ref}\"" >> "$ZSHRC"
      fi
    else
      # 如果没有对应 model 定义，注释掉或删除已有的 ANTHROPIC_MODEL
      sed -i.tmp -E "s|^export ANTHROPIC_MODEL=.*|# export ANTHROPIC_MODEL # 已清除|" "$ZSHRC"
    fi

    # 删除临时文件
    rm -f "${ZSHRC}.tmp"

    echo "✅ 已持久化切换到 $provider（下次打开终端自动生效）" >&2
  fi

  echo "export ANTHROPIC_BASE_URL=\"\$${provider_upper}_BASE_URL\";"
  echo "export ANTHROPIC_AUTH_TOKEN=\"${auth_token_ref}\";"

  if [ "$has_model_def" = true ]; then
    echo "export ANTHROPIC_MODEL=\"${model_ref}\";"
  else
    echo "if [ -n \"\$ANTHROPIC_MODEL\" ]; then unset ANTHROPIC_MODEL; fi;"
  fi

  if [[ "$persist_flag" != "--persist" ]]; then
    echo "✅ 已切换到 $provider（仅当前终端有效）" >&2
  fi

  exit 10
}

# 测试当前配置
test_config() {
  echo "🔍 测试 AI API 配置..."

  if [ -z "$ANTHROPIC_BASE_URL" ]; then
    echo "❌ ANTHROPIC_BASE_URL 未设置"
    echo "请运行 'source ~/.zshrc' 或重新打开终端"
    exit 1
  fi

  if [ -z "$ANTHROPIC_AUTH_TOKEN" ]; then
    echo "❌ ANTHROPIC_AUTH_TOKEN 未设置"
    echo "请运行 'source ~/.zshrc' 或重新打开终端"
    exit 1
  fi

  echo "  ✅ BASE_URL: $ANTHROPIC_BASE_URL"
  echo "  ✅ AUTH_TOKEN: ${ANTHROPIC_AUTH_TOKEN:0:20}..."
  echo ""
  echo "配置已加载，可以正常使用"

  exit 0
}

# 清除环境变量（切换到 OAuth 登录模式）
unset_env() {
  local persist_flag="$1"

  # 检测 shell wrapper 是否已配置
  if ! grep -q "# siti shell wrapper" "$ZSHRC" 2>/dev/null; then
    echo "⚠️  检测到 shell wrapper 未配置，清除后不会在当前终端生效" >&2
    echo "" >&2
    echo "请运行以下命令配置 shell wrapper（仅需一次）：" >&2
    echo "  eval \"\$(siti init zsh)\" >> ~/.zshrc" >&2
    echo "  source ~/.zshrc" >&2
    exit 1
  fi

  # 需要清除的变量列表
  local vars=("ANTHROPIC_AUTH_TOKEN" "ANTHROPIC_API_KEY" "ANTHROPIC_BASE_URL" "ANTHROPIC_MODEL")

  # 持久模式：修改 ~/.zshrc
  if [[ "$persist_flag" == "--persist" ]]; then
    for var in "${vars[@]}"; do
      # 使用 sed 注释掉相关配置行
      sed -i.tmp -E "s|^export ${var}=.*|# export ${var} # 已清除|" "$ZSHRC"
    done
    rm -f "${ZSHRC}.tmp"

    echo "echo '✅ 已清除环境变量 [下次打开终端自动生效]';"
  fi

  for var in "${vars[@]}"; do
    echo "unset ${var};"
  done

  if [[ "$persist_flag" != "--persist" ]]; then
    echo "✅ 已清除环境变量（仅当前终端有效）" >&2
  fi
  echo "👉 提示: 运行 \"claude login\" 切换到 OAuth 登录模式" >&2

  exit 10
}

# 主逻辑
case "$1" in
  switch)
    switch_provider "${@:2}"
    ;;
  current)
    show_current
    ;;
  list)
    list_providers
    ;;
  test)
    test_config
    ;;
  unset)
    # 支持 --persist 和 -p 两种形式
    local persist_flag="$2"
    if [[ "$persist_flag" == "-p" ]]; then
      persist_flag="--persist"
    fi
    unset_env "$persist_flag"
    ;;
  ""|--help|-h)
    echo "用法:"
    echo "  siti ai switch <provider> [--persist]  切换 AI 服务商"
    echo "  siti ai current                        显示当前配置"
    echo "  siti ai list                           列出所有服务商"
    echo "  siti ai test                           测试当前配置"
    echo "  siti ai unset [--persist]              清除环境变量（切换到 OAuth 登录模式）"
    echo ""
    echo "选项:"
    echo "  --persist    持久化切换（修改 ~/.zshrc，下次打开终端自动生效）"
    echo "               不加此参数则仅在当前终端临时切换"
    echo ""
    echo "规则:"
    echo "  • 服务商需要在 ~/.zshrc 中定义 <PROVIDER>_BASE_URL"
    echo "  • 如果 <PROVIDER>_API_KEY 不存在，会使用 DEFAULT_AUTH_TOKEN 兜底"
    echo "  • 使用 SITI_AI_SKIP 环境变量跳过特定服务商（逗号分隔）"
  echo "    示例: export SITI_AI_SKIP=\"OPENAI,BAILIAN\""
    echo ""
    echo "示例:"
    echo "  siti ai list                    # 查看所有服务商"
    echo "  siti ai switch minimax          # 临时切换到 MiniMax（仅当前终端）"
    echo "  siti ai switch zhipu --persist  # 持久化切换到智谱（修改 zshrc）"
    echo "  siti ai current                 # 查看当前配置"
    echo "  siti ai unset                   # 清除环境变量（临时，切回 OAuth 登录）"
    echo "  siti ai unset --persist         # 清除环境变量（持久化）"
    exit 0
    ;;
  *)
    echo "❌ 未知命令: $1" >&2
    echo "运行 'siti ai --help' 查看帮助" >&2
    exit 1
    ;;
esac
