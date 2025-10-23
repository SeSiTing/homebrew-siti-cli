#!/bin/bash

# 描述: 管理终端代理设置
# 补全:
#   on: 开启终端代理
#   off: 关闭终端代理
#   check: 检查当前代理状态
#   status: 显示代理状态
# 用法:
#   siti proxy on      开启终端代理
#   siti proxy off     关闭终端代理
#   siti proxy check   检查当前代理状态

# 代理服务器配置
PROXY_HOST="127.0.0.1"
PROXY_PORT="7890"

CMD="$1"

enable_proxy() {
  echo "export http_proxy='http://${PROXY_HOST}:${PROXY_PORT}';"
  echo "export https_proxy='http://${PROXY_HOST}:${PROXY_PORT}';"
  echo "export all_proxy='socks5://${PROXY_HOST}:${PROXY_PORT}';"
  echo "echo '✅ 终端代理已开启 (${PROXY_HOST}:${PROXY_PORT})';"
}

disable_proxy() {
  echo "unset http_proxy;"
  echo "unset https_proxy;"
  echo "unset all_proxy;"
  echo "echo '🚫 终端代理已关闭';"
}

check_proxy() {
  echo "当前代理状态:"
  if [ -n "$http_proxy" ]; then
    echo "  ✅ 代理已开启"
    echo "  http_proxy:  $http_proxy"
    echo "  https_proxy: $https_proxy"
    echo "  all_proxy:   $all_proxy"
  else
    echo "  ❌ 代理未开启"
  fi
}

case "$CMD" in
  "on")
    enable_proxy
    ;;
  "off")
    disable_proxy
    ;;
  "check"|"status"|"")
    check_proxy
    ;;
  *)
    echo "❌ 未知命令: $CMD" >&2
    echo "用法:" >&2
    echo "  siti proxy on    # 开启代理" >&2
    echo "  siti proxy off   # 关闭代理" >&2
    echo "  siti proxy check # 检查状态" >&2
    exit 1
    ;;
esac
