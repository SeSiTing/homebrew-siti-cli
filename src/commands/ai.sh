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
ZSHENV="$HOME/.zshenv"

# 合并读取两个配置文件（zshenv 优先）
_cat_configs() {
  cat "$ZSHENV" "$ZSHRC" 2>/dev/null
}

# 查找某个 pattern 首次出现在哪个配置文件（找不到则兜底 zshrc）
_find_config_file() {
  local pattern="$1"
  if grep -q "$pattern" "$ZSHENV" 2>/dev/null; then
    echo "$ZSHENV"
  elif grep -q "$pattern" "$ZSHRC" 2>/dev/null; then
    echo "$ZSHRC"
  else
    echo "$ZSHRC"
  fi
}

# 读取跳过列表（从环境变量或配置文件）
get_skip_list() {
  local skip_list="${SITI_AI_SKIP:-}"
  if [ -z "$skip_list" ]; then
    skip_list=$(_cat_configs | grep '^export SITI_AI_SKIP=' | tail -1 | sed -E 's/.*="(.*)"/\1/')
  fi
  echo "$skip_list"
}

# 列出所有可用的 AI 服务商
list_providers() {
  echo "可用的 AI 服务商:"
  
  local skip_list
  skip_list=$(get_skip_list)
  
  # 从配置文件提取所有 *_BASE_URL（排除 ANTHROPIC_BASE_URL，去重保留首次定义）
  _cat_configs | \
    grep -E '^export [A-Z0-9_]+_BASE_URL=' | \
    grep -v 'ANTHROPIC_BASE_URL' | \
    grep -v 'SITI_AI_SKIP' | \
    awk '!seen[$0]++' | \
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
      
      # 优先用环境变量判断当前（反映临时切换），其次查配置文件（持久配置）
      if [ -n "$ANTHROPIC_BASE_URL" ] && [ "$ANTHROPIC_BASE_URL" = "$url" ]; then
        printf "  • %-15s %s ← 当前\n" "$provider_lower" "$url"
      elif [ -z "$ANTHROPIC_BASE_URL" ] && _cat_configs | grep -qE "^export ANTHROPIC_BASE_URL=\"\\\$${provider}_BASE_URL\"" 2>/dev/null; then
        printf "  • %-15s %s ← 当前\n" "$provider_lower" "$url"
      else
        printf "  • %-15s %s\n" "$provider_lower" "$url"
      fi
    done
  
  exit 0
}

# 显示当前配置
show_current() {
  echo "当前 AI API 配置:"

  # 优先从当前环境变量读取（反映临时切换）
  if [ -n "$ANTHROPIC_BASE_URL" ]; then
    # 尝试从 zshrc 反向查找服务商名称（匹配 URL）
    local provider_name=""
    while IFS= read -r line; do
      local pvar purl
      pvar=$(echo "$line" | sed -E 's/export ([A-Z0-9_]+)_BASE_URL=.*/\1/')
      purl=$(echo "$line" | sed -E 's/.*="(.*)"/\1/')
      if [ "$purl" = "$ANTHROPIC_BASE_URL" ]; then
        provider_name=$(echo "$pvar" | tr '[:upper:]' '[:lower:]')
        break
      fi
    done < <(_cat_configs | grep -E '^export [A-Z0-9_]+_BASE_URL=' | grep -v 'ANTHROPIC_BASE_URL' | awk '!seen[$0]++')

    if [ -n "$provider_name" ]; then
      echo "  服务商: $provider_name"
    fi
    echo "  BASE_URL: $ANTHROPIC_BASE_URL"

    if [ -n "$ANTHROPIC_AUTH_TOKEN" ]; then
      local token_preview="${ANTHROPIC_AUTH_TOKEN:0:20}"
      echo "  AUTH_TOKEN: ${token_preview}..."
    fi

    if [ -n "$ANTHROPIC_MODEL" ]; then
      echo "  ANTHROPIC_MODEL: $ANTHROPIC_MODEL"
    fi
    if [ -n "$ANTHROPIC_DEFAULT_SONNET_MODEL" ]; then
      echo "  ANTHROPIC_DEFAULT_SONNET_MODEL: $ANTHROPIC_DEFAULT_SONNET_MODEL"
    fi
    if [ -n "$ANTHROPIC_DEFAULT_OPUS_MODEL" ]; then
      echo "  ANTHROPIC_DEFAULT_OPUS_MODEL: $ANTHROPIC_DEFAULT_OPUS_MODEL"
    fi
    if [ -n "$ANTHROPIC_DEFAULT_HAIKU_MODEL" ]; then
      echo "  ANTHROPIC_DEFAULT_HAIKU_MODEL: $ANTHROPIC_DEFAULT_HAIKU_MODEL"
    fi
    if [ -n "$ANTHROPIC_REASONING_MODEL" ]; then
      echo "  ANTHROPIC_REASONING_MODEL: $ANTHROPIC_REASONING_MODEL"
    fi
  else
    echo "  ❌ 未配置（ANTHROPIC_BASE_URL 未设置）"
    echo "  提示: 运行 'source ~/.zshrc' 或重新打开终端"
  fi

  exit 0
}

# 切换服务商
switch_provider() {
  local provider="$1"
  local persist_flag="$2"
  
  # 检测 shell wrapper 是否已配置（检查配置文件内容，不依赖子进程）
  if ! _cat_configs | grep -q "# siti shell wrapper" 2>/dev/null; then
    echo "⚠️  检测到 shell wrapper 未配置，切换后不会在当前终端生效" >&2
    echo "" >&2
    echo "请运行以下命令配置 shell wrapper（仅需一次）：" >&2
    echo "  eval \"\$(siti init zsh)\" >> ~/.zshrc" >&2
    echo "  source ~/.zshrc" >&2
    echo "" >&2
    read -p "是否继续（仅持久化到 ~/.zshrc）？[y/N] " response
    if [[ ! "$response" =~ ^[yY]$ ]]; then
      echo "已取消" >&2
      exit 1
    fi
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
  
  # 检查服务商是否存在（扫描两个配置文件）
  if ! _cat_configs | grep -q "^export ${provider_upper}_BASE_URL="; then
    echo "❌ 服务商 '$provider' 不存在" >&2
    echo "" >&2
    list_providers >&2
    exit 1
  fi
  
  # 决定 AUTH_TOKEN 引用（兜底到 DEFAULT_AUTH_TOKEN）
  local auth_token_ref
  if _cat_configs | grep -q "^export ${provider_upper}_API_KEY="; then
    auth_token_ref="\$${provider_upper}_API_KEY"
  else
    # 兜底：使用 DEFAULT_AUTH_TOKEN
    auth_token_ref="\$DEFAULT_AUTH_TOKEN"
  fi
  
  # 检查是否有模型变量（如 ALI_MODEL, MINIMAX_MODEL 等）
  local has_model_var=false
  if _cat_configs | grep -q "^export ${provider_upper}_MODEL="; then
    has_model_var=true
  fi
  
  # 持久模式：找到 ANTHROPIC_BASE_URL 所在的配置文件并修改
  if [[ "$persist_flag" == "--persist" ]]; then
    local persist_file
    persist_file=$(_find_config_file "^export ANTHROPIC_BASE_URL=")
    
    # 备份目标文件
    cp "$persist_file" "${persist_file}.backup.$(date +%Y%m%d_%H%M%S)"
    
    # 替换 ANTHROPIC_BASE_URL
    sed -i.tmp -E "s|^export ANTHROPIC_BASE_URL=.*|export ANTHROPIC_BASE_URL=\"\$${provider_upper}_BASE_URL\"|" "$persist_file"
    
    # 替换 ANTHROPIC_AUTH_TOKEN
    sed -i.tmp -E "s|^export ANTHROPIC_AUTH_TOKEN=.*|export ANTHROPIC_AUTH_TOKEN=\"${auth_token_ref}\"|" "$persist_file"
    
    # 处理 5 个 ANTHROPIC 模型变量
    if [ "$has_model_var" = true ]; then
      sed -i.tmp -E "s|^export ANTHROPIC_MODEL=.*|export ANTHROPIC_MODEL=\"\$${provider_upper}_MODEL\"|" "$persist_file"
      sed -i.tmp -E "s|^export ANTHROPIC_DEFAULT_SONNET_MODEL=.*|export ANTHROPIC_DEFAULT_SONNET_MODEL=\"\$${provider_upper}_MODEL\"|" "$persist_file"
      sed -i.tmp -E "s|^export ANTHROPIC_DEFAULT_OPUS_MODEL=.*|export ANTHROPIC_DEFAULT_OPUS_MODEL=\"\$${provider_upper}_MODEL\"|" "$persist_file"
      sed -i.tmp -E "s|^export ANTHROPIC_DEFAULT_HAIKU_MODEL=.*|export ANTHROPIC_DEFAULT_HAIKU_MODEL=\"\$${provider_upper}_MODEL\"|" "$persist_file"
      sed -i.tmp -E "s|^export ANTHROPIC_REASONING_MODEL=.*|export ANTHROPIC_REASONING_MODEL=\"\$${provider_upper}_MODEL\"|" "$persist_file"
    else
      sed -i.tmp -E "s|^export ANTHROPIC_MODEL=.*|# export ANTHROPIC_MODEL # 已清除|" "$persist_file"
      sed -i.tmp -E "s|^export ANTHROPIC_DEFAULT_SONNET_MODEL=.*|# export ANTHROPIC_DEFAULT_SONNET_MODEL # 已清除|" "$persist_file"
      sed -i.tmp -E "s|^export ANTHROPIC_DEFAULT_OPUS_MODEL=.*|# export ANTHROPIC_DEFAULT_OPUS_MODEL # 已清除|" "$persist_file"
      sed -i.tmp -E "s|^export ANTHROPIC_DEFAULT_HAIKU_MODEL=.*|# export ANTHROPIC_DEFAULT_HAIKU_MODEL # 已清除|" "$persist_file"
      sed -i.tmp -E "s|^export ANTHROPIC_REASONING_MODEL=.*|# export ANTHROPIC_REASONING_MODEL # 已清除|" "$persist_file"
    fi
    
    rm -f "${persist_file}.tmp"
    echo "✅ 已持久化切换到 $provider → $(basename "$persist_file") [下次打开终端自动生效]" >&2
  fi
  
  # 输出 export 命令（临时模式和持久模式都输出，供当前 shell 立即生效）
  echo "export ANTHROPIC_BASE_URL=\"\$${provider_upper}_BASE_URL\";"
  echo "export ANTHROPIC_AUTH_TOKEN=\"${auth_token_ref}\";"
  
  # 输出模型变量的设置或清除
  if [ "$has_model_var" = true ]; then
    echo "export ANTHROPIC_MODEL=\"\$${provider_upper}_MODEL\";"
    echo "export ANTHROPIC_DEFAULT_SONNET_MODEL=\"\$${provider_upper}_MODEL\";"
    echo "export ANTHROPIC_DEFAULT_OPUS_MODEL=\"\$${provider_upper}_MODEL\";"
    echo "export ANTHROPIC_DEFAULT_HAIKU_MODEL=\"\$${provider_upper}_MODEL\";"
    echo "export ANTHROPIC_REASONING_MODEL=\"\$${provider_upper}_MODEL\";"
  else
    echo "unset ANTHROPIC_MODEL;"
    echo "unset ANTHROPIC_DEFAULT_SONNET_MODEL;"
    echo "unset ANTHROPIC_DEFAULT_OPUS_MODEL;"
    echo "unset ANTHROPIC_DEFAULT_HAIKU_MODEL;"
    echo "unset ANTHROPIC_REASONING_MODEL;"
  fi
  
  if [[ "$persist_flag" != "--persist" ]]; then
    echo "✅ 已切换到 $provider [仅当前终端有效]" >&2
  fi
  
  exit 10  # 退出码 10 表示需要 eval
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
  if ! _cat_configs | grep -q "# siti shell wrapper" 2>/dev/null; then
    echo "⚠️  检测到 shell wrapper 未配置，清除后不会在当前终端生效" >&2
    echo "" >&2
    echo "请运行以下命令配置 shell wrapper（仅需一次）：" >&2
    echo "  eval \"\$(siti init zsh)\" >> ~/.zshrc" >&2
    echo "  source ~/.zshrc" >&2
    echo "" >&2
    read -p "是否继续（仅持久化到 ~/.zshrc）？[y/N] " response
    if [[ ! "$response" =~ ^[yY]$ ]]; then
      echo "已取消" >&2
      exit 1
    fi
  fi

  # 需要清除的变量列表（包含 5 个模型变量）
  local vars=("ANTHROPIC_AUTH_TOKEN" "ANTHROPIC_API_KEY" "ANTHROPIC_BASE_URL" "ANTHROPIC_MODEL" "ANTHROPIC_DEFAULT_SONNET_MODEL" "ANTHROPIC_DEFAULT_OPUS_MODEL" "ANTHROPIC_DEFAULT_HAIKU_MODEL" "ANTHROPIC_REASONING_MODEL")

  # 持久模式：找到各变量所在的配置文件并注释掉
  if [[ "$persist_flag" == "--persist" ]]; then
    for var in "${vars[@]}"; do
      local target_file
      target_file=$(_find_config_file "^export ${var}=")
      sed -i.tmp -E "s|^export ${var}=.*|# export ${var} # 已清除|" "$target_file"
      rm -f "${target_file}.tmp"
    done

    echo "✅ 已清除环境变量 [下次打开终端自动生效]" >&2
  fi

  # 输出 unset 命令（临时模式和持久模式都输出，供当前 shell 立即生效）
  for var in "${vars[@]}"; do
    echo "unset ${var};"
  done

  if [[ "$persist_flag" != "--persist" ]]; then
    echo "✅ 已清除环境变量 [仅当前终端有效]" >&2
  fi
  echo "👉 提示: 运行 \"claude login\" 切换到 OAuth 登录模式" >&2

  exit 10  # 退出码 10 表示需要 eval
}

# 主逻辑
case "$1" in
  switch)
    switch_provider "$2" "$3"
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
    persist_flag="$2"
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
    echo "  --persist    持久化切换（修改配置文件，下次打开终端自动生效）"
    echo "               不加此参数则仅在当前终端临时切换"
    echo ""
    echo "规则:"
    echo "  • 服务商需要在 ~/.zshrc 或 ~/.zshenv 中定义 <PROVIDER>_BASE_URL"
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
