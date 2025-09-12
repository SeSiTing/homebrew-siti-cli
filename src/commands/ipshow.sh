#!/bin/bash

echo "🌐 内网 IP："
ipconfig getifaddr en0

echo "🌎 公网 IP："
curl -s ifconfig.me
