#!/bin/bash

# 描述: 管理 AI API 配置切换
# 补全:
#   switch: 切换 AI 服务商
#   current: 显示当前配置
#   list: 列出所有服务商
#   test: 测试当前配置
# 用法:
#   siti ai switch <provider>    切换到指定服务商
#   siti ai current              显示当前配置
#   siti ai list                 列出所有服务商
#   siti ai test                 测试当前配置

ZSHRC="$HOME/.zshrc"

# 列出所有可用的 AI 服务商
list_providers() {
  echo "可用的 AI 服务商:"
  
  # 从 ~/.zshrc 提取所有 *_BASE_URL（排除 ANTHROPIC_BASE_URL）
  grep -E '^export [A-Z_]+_BASE_URL=' "$ZSHRC" 2>/dev/null | \
    grep -v 'ANTHROPIC_BASE_URL' | \
    while IFS= read -r line; do
      # 提取变量名和值
      provider=$(echo "$line" | sed -E 's/export ([A-Z_]+)_BASE_URL=.*/\1/')
      url=$(echo "$line" | sed -E 's/.*="(.*)"/\1/')
      
      # 转换为小写显示
      provider_lower=$(echo "$provider" | tr '[:upper:]' '[:lower:]')
      
      # 检查是否为当前使用的
      if grep -q "ANTHROPIC_BASE_URL=\"\$$provider" "$ZSHRC" 2>/dev/null; then
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
  
  # 从 ~/.zshrc 读取当前配置
  local base_url_line=$(grep '^export ANTHROPIC_BASE_URL=' "$ZSHRC" 2>/dev/null | tail -1)
  local auth_token_line=$(grep '^export ANTHROPIC_AUTH_TOKEN=' "$ZSHRC" 2>/dev/null | tail -1)
  
  if [ -n "$base_url_line" ]; then
    # 提取引用的变量名
    local provider_var=$(echo "$base_url_line" | sed -E 's/.*"\$([A-Z_]+)_BASE_URL".*/\1/')
    if [ -n "$provider_var" ]; then
      local provider=$(echo "$provider_var" | tr '[:upper:]' '[:lower:]')
      echo "  服务商: $provider"
      
      # 显示实际的 URL（如果环境变量已加载）
      if [ -n "$ANTHROPIC_BASE_URL" ]; then
        echo "  BASE_URL: $ANTHROPIC_BASE_URL"
      fi
      
      # 显示 TOKEN（脱敏）
      if [ -n "$ANTHROPIC_AUTH_TOKEN" ]; then
        local token_preview="${ANTHROPIC_AUTH_TOKEN:0:20}"
        echo "  AUTH_TOKEN: ${token_preview}..."
      fi
    else
      echo "  BASE_URL: $(echo "$base_url_line" | sed -E 's/.*="(.*)"/\1/')"
    fi
  else
    echo "  ❌ 未配置"
  fi
  
  exit 0
}

# 切换服务商
switch_provider() {
  local provider="$1"
  
  if [ -z "$provider" ]; then
    echo "❌ 请指定服务商名称" >&2
    echo "运行 'siti ai list' 查看可用服务商" >&2
    exit 1
  fi
  
  # 转换为大写
  local provider_upper=$(echo "$provider" | tr '[:lower:]' '[:upper:]')
  
  # 检查服务商是否存在
  if ! grep -q "^export ${provider_upper}_BASE_URL=" "$ZSHRC" 2>/dev/null; then
    echo "❌ 服务商 '$provider' 不存在" >&2
    echo "" >&2
    list_providers >&2
    exit 1
  fi
  
  # 备份 ~/.zshrc
  cp "$ZSHRC" "${ZSHRC}.backup.$(date +%Y%m%d_%H%M%S)"
  
  # 使用 sed 替换 ANTHROPIC_BASE_URL
  sed -i.tmp -E "s|^export ANTHROPIC_BASE_URL=.*|export ANTHROPIC_BASE_URL=\"\$${provider_upper}_BASE_URL\"|" "$ZSHRC"
  
  # 使用 sed 替换 ANTHROPIC_AUTH_TOKEN
  sed -i.tmp -E "s|^export ANTHROPIC_AUTH_TOKEN=.*|export ANTHROPIC_AUTH_TOKEN=\"\$${provider_upper}_API_KEY\"|" "$ZSHRC"
  
  # 删除临时文件
  rm -f "${ZSHRC}.tmp"
  
  # 输出 export 命令（供 eval 使用）
  echo "export ANTHROPIC_BASE_URL=\"\$${provider_upper}_BASE_URL\";"
  echo "export ANTHROPIC_AUTH_TOKEN=\"\$${provider_upper}_API_KEY\";"
  echo "echo '✅ 已切换到 $provider';"
  
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

# 主逻辑
case "$1" in
  switch)
    switch_provider "$2"
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
  ""|--help|-h)
    echo "用法:"
    echo "  siti ai switch <provider>  切换 AI 服务商"
    echo "  siti ai current            显示当前配置"
    echo "  siti ai list               列出所有服务商"
    echo "  siti ai test               测试当前配置"
    echo ""
    echo "示例:"
    echo "  siti ai list               # 查看所有服务商"
    echo "  siti ai switch minimax     # 切换到 MiniMax"
    echo "  siti ai switch zhipu       # 切换到智谱"
    echo "  siti ai current            # 查看当前配置"
    exit 0
    ;;
  *)
    echo "❌ 未知命令: $1" >&2
    echo "运行 'siti ai --help' 查看帮助" >&2
    exit 1
    ;;
esac
