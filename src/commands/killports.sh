#!/bin/bash

# 描述: 释放被占用的端口，支持端口范围
# 补全:
#   check: 仅检查端口占用情况，不释放
#   help: 显示帮助信息
# 用法:
#   siti killports           释放默认端口范围
#   siti killports 3000 5000 释放指定端口
#   siti killports 3000-3010 释放端口范围
#   siti killports check     仅检查端口占用情况，不释放
# 默认端口范围（2024～2030、8000～8010、8080～8090 和 9000～9010）
DEFAULT_PORTS=($(seq 2024 2030) $(seq 8000 8010) $(seq 8080 8090) $(seq 9000 9010))

MODE="kill"

# 解析端口范围函数
parse_port_range() {
  local range="$1"
  local start_port
  local end_port
  
  # 检查是否为范围格式 (例如: 3000-3010)
  if [[ "$range" =~ ^([0-9]+)-([0-9]+)$ ]]; then
    start_port="${BASH_REMATCH[1]}"
    end_port="${BASH_REMATCH[2]}"
    
    # 验证端口范围有效性
    if [ "$start_port" -gt "$end_port" ]; then
      echo "错误: 无效的端口范围 $range (起始端口大于结束端口)" >&2
      return 1
    fi
    
    # 生成端口序列
    seq "$start_port" "$end_port"
  else
    # 不是范围，返回原始值
    echo "$range"
  fi
}

# 处理命令行参数
if [ "$1" == "help" ] || [ "$1" == "--help" ]; then
  echo "用法: siti killports [选项] [端口...]"
  echo
  echo "选项:"
  echo "  check              仅检查端口占用情况，不终止进程"
  echo "  help, --help       显示此帮助信息"
  echo
  echo "端口指定方式:"
  echo "  单个端口:          siti killports 8080"
  echo "  多个端口:          siti killports 8080 9000 3000"
  echo "  端口范围:          siti killports 3000-3010"
  echo "  混合方式:          siti killports 8080 3000-3010 9000"
  echo
  echo "默认端口范围: 2024-2030, 8000-8010, 8080-8090, 9000-9010"
  exit 0
fi

if [ "$1" == "check" ]; then
  MODE="check"
  shift
fi

# 处理端口参数
PORTS=()
if [ $# -eq 0 ]; then
  # 使用默认端口
  PORTS=("${DEFAULT_PORTS[@]}")
else
  # 处理用户指定的端口和端口范围
  for arg in "$@"; do
    # 解析端口范围
    range_ports=$(parse_port_range "$arg")
    if [ $? -ne 0 ]; then
      exit 1
    fi
    
    # 添加到端口列表
    for port in $range_ports; do
      PORTS+=("$port")
    done
  done
fi

echo "🔍 正在扫描端口占用..."

for PORT in "${PORTS[@]}"; do
  PIDS=$(lsof -ti tcp:$PORT)
  if [ -n "$PIDS" ]; then
    FIRST_PID=$(echo "$PIDS" | head -n1)
    CMDLINE=$(ps -p $FIRST_PID -o args=)

    if echo "$CMDLINE" | grep -qi "java"; then
      TYPE="☕ Java"
    elif echo "$CMDLINE" | grep -qi "python"; then
      TYPE="🐍 Python"
    elif echo "$CMDLINE" | grep -qi "node"; then
      TYPE="🟢 Node.js"
    else
      TYPE="🧩 Other"
    fi

    PID_LIST=$(echo $PIDS | tr '\n' ' ')
    echo "⚠️  $PORT 端口被占用 - $TYPE - PIDs: [$PID_LIST]"

    if [ "$MODE" == "kill" ]; then
      for pid in $PIDS; do
        kill -9 $pid >/dev/null 2>&1
      done
    fi
  fi
done

if [ "$MODE" == "check" ]; then
  echo "📝 检查模式，未终止任何进程"
else
  echo "✅ 完成，占用进程已处理"
fi