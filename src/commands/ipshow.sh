#!/bin/bash

# 描述: 显示当前IP地址
echo "🌐 内网 IP："
ipconfig getifaddr en0

echo "🌎 公网 IP："
curl -s ifconfig.me
