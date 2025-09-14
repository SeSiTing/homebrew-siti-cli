#!/bin/bash

# 描述: 清理日志文件
# 补全:
#   help: 显示帮助信息
find . -type f -name "*.log" -delete
echo "🧹 已清理当前目录下所有日志文件 (*.log)"
