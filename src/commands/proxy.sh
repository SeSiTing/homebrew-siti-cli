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
# 转换命令为小写（兼容 bash 3.x）
CMD_LOWER=$(echo "$CMD" | tr '[:upper:]' '[:lower:]')

enable_proxy() {
  echo "export http_proxy='http://${PROXY_HOST}:${PROXY_PORT}';"
  echo "export HTTP_PROXY='http://${PROXY_HOST}:${PROXY_PORT}';"
  echo "export https_proxy='http://${PROXY_HOST}:${PROXY_PORT}';"
  echo "export HTTPS_PROXY='http://${PROXY_HOST}:${PROXY_PORT}';"
  echo "export all_proxy='socks5://${PROXY_HOST}:${PROXY_PORT}';"
  echo "export ALL_PROXY='socks5://${PROXY_HOST}:${PROXY_PORT}';"
  echo "echo '✅ 终端代理已开启 (${PROXY_HOST}:${PROXY_PORT})';"
  exit 10  # 退出码 10 表示需要 eval
}

disable_proxy() {
  echo "unset http_proxy HTTP_PROXY;"
  echo "unset https_proxy HTTPS_PROXY;"
  echo "unset all_proxy ALL_PROXY;"
  echo "echo '🚫 终端代理已关闭';"
  exit 10  # 退出码 10 表示需要 eval
}

check_proxy() {
  echo "当前代理状态:"
  # 优先检查小写版本
  local http_val="${http_proxy:-$HTTP_PROXY}"
  local https_val="${https_proxy:-$HTTPS_PROXY}"
  local all_val="${all_proxy:-$ALL_PROXY}"
  if [ -n "$http_val" ]; then
    echo "  ✅ 代理已开启"
    echo "  http_proxy:  $http_val"
    echo "  https_proxy: $https_val"
    echo "  all_proxy:   $all_val"
  else
    echo "  ❌ 代理未开启"
  fi
  # no_proxy 忽略大小写都检查，有两个就都打印
  [ -n "${no_proxy}" ] && echo "  no_proxy:    $no_proxy"
  [ -n "${NO_PROXY}" ] && echo "  NO_PROXY:    $NO_PROXY"
  exit 0  # 正常退出，不需要 eval
}

case "$CMD_LOWER" in
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
