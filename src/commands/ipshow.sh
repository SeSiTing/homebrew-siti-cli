#!/bin/bash

# 描述: 显示当前IP地址
# 补全:
#   help: 显示帮助信息
echo "🌐 内网 IP："
ipconfig getifaddr en0

echo "🌎 公网 IP："
curl -s ifconfig.me
