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
# 代理服务器配置
PROXY_HOST="127.0.0.1"
PROXY_PORT="7890"

# 获取命令参数
CMD="$1"

# 开启代理函数
enable_proxy() {
  # 设置各种代理环境变量
  export http_proxy="http://${PROXY_HOST}:${PROXY_PORT}"
  export https_proxy="http://${PROXY_HOST}:${PROXY_PORT}"
  export all_proxy="socks5://${PROXY_HOST}:${PROXY_PORT}"
  echo "✅ 终端代理已开启 (${PROXY_HOST}:${PROXY_PORT})"
}

# 关闭代理函数
disable_proxy() {
  # 清除代理环境变量
  unset http_proxy
  unset https_proxy
  unset all_proxy
  echo "🚫 终端代理已关闭"
}

# 检查代理状态函数
check_proxy() {
  # 显示当前代理环境变量
  echo "http_proxy:  $http_proxy"
  echo "https_proxy: $https_proxy"
  echo "all_proxy:   $all_proxy"
}

# 根据命令参数执行相应操作
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
    echo "❌ 未知命令: $CMD"
    echo "用法: siti proxy [on|off|check]"
    exit 1
    ;;
esac
