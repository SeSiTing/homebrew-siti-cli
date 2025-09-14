#!/bin/bash

# 描述: 管理终端代理设置
# 补全:
#   on: 开启终端代理
#   off: 关闭终端代理
#   check: 检查当前代理状态
#   status: 显示代理状态
# 用法:
#   siti proxy on     开启终端代理
#   siti proxy off    关闭终端代理
#   siti proxy check  检查当前代理状态
# 代理配置
PROXY_HOST="127.0.0.1"
PROXY_PORT="7890"

# 命令参数
CMD="$1"

# 开启代理
enable_proxy() {
  export http_proxy="http://${PROXY_HOST}:${PROXY_PORT}"
  export https_proxy="http://${PROXY_HOST}:${PROXY_PORT}"
  export all_proxy="socks5://${PROXY_HOST}:${PROXY_PORT}"
  echo "✅ 终端代理已开启 (${PROXY_HOST}:${PROXY_PORT})"
}

# 关闭代理
disable_proxy() {
  unset http_proxy
  unset https_proxy
  unset all_proxy
  echo "🚫 终端代理已关闭"
}

# 检查代理状态
check_proxy() {
  echo "http_proxy:  $http_proxy"
  echo "https_proxy: $https_proxy"
  echo "all_proxy:   $all_proxy"
}

# 根据命令执行相应操作
case "$CMD" in
  "on")
    enable_proxy
    ;;
  "off")
    disable_proxy
    ;;
  "check"|"")
    check_proxy
    ;;
  *)
    echo "❌ 未知命令: $CMD"
    echo "用法: siti proxy [on|off|check]"
    exit 1
    ;;
esac
