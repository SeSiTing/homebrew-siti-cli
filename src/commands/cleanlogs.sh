#!/bin/bash

# 描述: 清理日志文件
# 补全:
#   help: 显示帮助信息

# 使用find命令删除所有.log文件
find . -type f -name "*.log" -delete

# 输出完成信息
echo "🧹 已清理当前目录下所有日志文件 (*.log)"
