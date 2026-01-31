#!/bin/bash

# 描述: 释放被占用的端口，支持端口范围和预设组
# 补全:
#   check: 仅检查端口占用情况，不释放
#   --dev: 清理常用开发端口
#   --db: 清理常用数据库端口
#   --web: 清理常用 Web 服务端口
#   --java: 清理 Java 常用端口（8080-8090）
#   --all: 显示所有占用端口，逐个询问是否清理
#   help: 显示帮助信息
# 用法:
#   siti killports              释放默认端口范围
#   siti killports 3000 5000    释放指定端口
#   siti killports 3000-3010    释放端口范围
#   siti killports 3000,5000,8080  释放逗号分隔的端口
#   siti killports --dev        清理开发端口（3000/5173/8080等）
#   siti killports --db         清理数据库端口（3306/5432/27017等）
#   siti killports --web        清理 Web 端口（80/443/8000/8080等）
#   siti killports --java       清理 Java 端口（8080-8090）
#   siti killports --all        显示所有占用端口并逐个确认
#   siti killports check        仅检查端口占用情况，不释放
# 默认端口范围（2024～2030、8000～8010、8080～8090 和 9000～9010）

# 定义默认端口范围
DEFAULT_PORTS=($(seq 2024 2030) $(seq 8000 8010) $(seq 8080 8090) $(seq 9000 9010))

# 预设端口组
DEV_PORTS=(3000 5173 8080 8000 4200 4000 3001 5000 9000)    # 常用开发端口
DB_PORTS=(3306 5432 27017 6379 5984 9200 9042 7000)         # 常用数据库端口
WEB_PORTS=(80 443 8000 8080 8888 9000 3000)                 # 常用 Web 端口
JAVA_PORTS=($(seq 8080 8090))                                # Java 常用端口范围

# 默认模式为kill
MODE="kill"
INTERACTIVE_ALL=false

# 解析端口范围或逗号分隔列表
parse_port_range() {
  local input="$1"
  local start_port
  local end_port
  
  # 处理逗号分隔的列表 (例如: 3000,5000,8080)
  if [[ "$input" == *","* ]]; then
    echo "$input" | tr ',' ' '
    return 0
  fi
  
  # 检查是否为范围格式 (例如: 3000-3010)
  if [[ "$input" =~ ^([0-9]+)-([0-9]+)$ ]]; then
    start_port="${BASH_REMATCH[1]}"
    end_port="${BASH_REMATCH[2]}"
    
    # 验证端口范围有效性
    if [ "$start_port" -gt "$end_port" ]; then
      echo "错误: 无效的端口范围 $input (起始端口大于结束端口)" >&2
      return 1
    fi
    
    # 生成端口序列
    seq "$start_port" "$end_port"
  else
    # 不是范围或列表，返回原始值
    echo "$input"
  fi
}

# 获取所有占用的端口
get_all_ports() {
  lsof -iTCP -sTCP:LISTEN -n -P 2>/dev/null | awk 'NR>1 {split($9,a,":"); print a[2]}' | sort -u
}

# 处理命令行参数
if [ "$1" == "help" ] || [ "$1" == "--help" ]; then
  echo "用法: siti killports [选项] [端口...]"
  echo
  echo "选项:"
  echo "  check              仅检查端口占用情况，不终止进程"
  echo "  --dev              清理常用开发端口 (3000/5173/8080等)"
  echo "  --db               清理常用数据库端口 (3306/5432/27017等)"
  echo "  --web              清理常用 Web 端口 (80/443/8000/8080等)"
  echo "  --java             清理 Java 常用端口范围 (8080-8090)"
  echo "  --all              显示所有占用端口，逐个询问是否清理"
  echo "  help, --help       显示此帮助信息"
  echo
  echo "端口指定方式:"
  echo "  单个端口:          siti killports 8080"
  echo "  多个端口:          siti killports 8080 9000 3000"
  echo "  端口范围:          siti killports 3000-3010"
  echo "  逗号分隔:          siti killports 3000,5000,8080"
  echo "  混合方式:          siti killports 8080 3000-3010 9000,5000"
  echo
  echo "预设组:"
  echo "  --dev              3000,5173,8080,8000,4200,4000,3001,5000,9000"
  echo "  --db               3306,5432,27017,6379,5984,9200,9042,7000"
  echo "  --web              80,443,8000,8080,8888,9000,3000"
  echo "  --java             8080-8090"
  echo
  echo "默认端口范围: 2024-2030, 8000-8010, 8080-8090, 9000-9010"
  exit 0
fi

# 检查模式设置
if [ "$1" == "check" ]; then
  MODE="check"
  shift
fi

# 处理预设和特殊选项
if [ "$1" == "--dev" ]; then
  PORTS=("${DEV_PORTS[@]}")
  echo "📦 使用开发端口预设"
  shift
elif [ "$1" == "--db" ]; then
  PORTS=("${DB_PORTS[@]}")
  echo "💾 使用数据库端口预设"
  shift
elif [ "$1" == "--web" ]; then
  PORTS=("${WEB_PORTS[@]}")
  echo "🌐 使用 Web 端口预设"
  shift
elif [ "$1" == "--java" ]; then
  PORTS=("${JAVA_PORTS[@]}")
  echo "☕ 使用 Java 端口预设 (8080-8090)"
  shift
elif [ "$1" == "--all" ]; then
  INTERACTIVE_ALL=true
  ALL_USED_PORTS=$(get_all_ports)
  if [ -z "$ALL_USED_PORTS" ]; then
    echo "✅ 没有发现占用的端口"
    exit 0
  fi
  echo "🔍 发现以下占用的端口:"
  for port in $ALL_USED_PORTS; do
    PIDS=$(lsof -ti tcp:"$port" 2>/dev/null)
    if [ -n "$PIDS" ]; then
      FIRST_PID=$(echo "$PIDS" | head -n1)
      CMDLINE=$(ps -p "$FIRST_PID" -o args= 2>/dev/null | cut -c 1-50)
      echo "  $port: $CMDLINE"
    fi
  done
  PORTS=($ALL_USED_PORTS)
  shift
fi

# 处理端口参数（如果还没有设置 PORTS）
if [ -z "${PORTS+x}" ]; then
  PORTS=()
  if [ $# -eq 0 ]; then
    # 使用默认端口
    PORTS=("${DEFAULT_PORTS[@]}")
  else
    # 处理用户指定的端口和端口范围
    for arg in "$@"; do
      # 解析端口范围或逗号分隔
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
fi

if [ "$MODE" == "check" ]; then
  echo "🔍 正在扫描端口占用（检查模式）..."
else
  echo "🔍 正在扫描端口占用..."
fi

# 记录统计
TOTAL_CHECKED=0
KILLED_COUNT=0

# 第一轮：扫描并收集占用的端口
OCCUPIED_PORTS=()
OCCUPIED_INFO=()

for PORT in "${PORTS[@]}"; do
  TOTAL_CHECKED=$((TOTAL_CHECKED + 1))
  
  PIDS=$(lsof -ti tcp:"$PORT" 2>/dev/null)
  if [ -n "$PIDS" ]; then
    FIRST_PID=$(echo "$PIDS" | head -n1)
    CMDLINE=$(ps -p "$FIRST_PID" -o args= 2>/dev/null)
    
    # 判断进程类型
    if echo "$CMDLINE" | grep -qi "java"; then
      TYPE="☕ Java"
    elif echo "$CMDLINE" | grep -qi "python"; then
      TYPE="🐍 Python"
    elif echo "$CMDLINE" | grep -qi "node"; then
      TYPE="🟢 Node.js"
    elif echo "$CMDLINE" | grep -qi "docker"; then
      TYPE="🐳 Docker"
    elif echo "$CMDLINE" | grep -qi "postgres"; then
      TYPE="🐘 PostgreSQL"
    elif echo "$CMDLINE" | grep -qi "mysql"; then
      TYPE="🐬 MySQL"
    elif echo "$CMDLINE" | grep -qi "redis"; then
      TYPE="🔴 Redis"
    else
      TYPE="🧩 Other"
    fi
    
    PID_LIST=$(echo "$PIDS" | tr '\n' ' ')
    CMDLINE_SHORT=$(echo "$CMDLINE" | cut -c 1-40)
    
    OCCUPIED_PORTS+=("$PORT")
    OCCUPIED_INFO+=("$PORT|$TYPE|$PID_LIST|$CMDLINE_SHORT")
  fi
done

# 如果没有占用的端口，直接退出
if [ ${#OCCUPIED_PORTS[@]} -eq 0 ]; then
  echo ""
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "✅ 扫描了 $TOTAL_CHECKED 个端口，没有发现占用"
  exit 0
fi

# 显示占用的端口列表
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "⚠️  发现 ${#OCCUPIED_PORTS[@]} 个端口被占用:"
echo ""

for info in "${OCCUPIED_INFO[@]}"; do
  IFS='|' read -r port type pids cmd <<< "$info"
  echo "  端口 $port - $type"
  echo "    PIDs: [$pids]"
  echo "    命令: $cmd"
  echo ""
done

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 如果是检查模式，直接退出
if [ "$MODE" == "check" ]; then
  echo "📝 检查模式: 扫描了 $TOTAL_CHECKED 个端口，未终止任何进程"
  exit 0
fi

# 如果是 --all 模式，逐个询问
if [[ "$INTERACTIVE_ALL" == "true" ]]; then
  echo "🔧 将逐个询问是否清理这些端口"
  echo ""
  
  for info in "${OCCUPIED_INFO[@]}"; do
    IFS='|' read -r port type pids cmd <<< "$info"
    echo "端口 $port - $type"
    echo "  命令: $cmd"
    read -p "  是否清理? [y/N] " confirm
    
    if [[ "$confirm" =~ ^[Yy] ]]; then
      for pid in $pids; do
        if kill -9 "$pid" 2>/dev/null; then
          KILLED_COUNT=$((KILLED_COUNT + 1))
        fi
      done
      echo "  ✅ 已清理"
    else
      echo "  ⏭️  跳过"
    fi
    echo ""
  done
else
  # 批量模式：一次性确认
  read -p "⚠️  是否清理以上所有端口? [y/N] " confirm
  
  if [[ ! "$confirm" =~ ^[Yy] ]]; then
    echo "❌ 已取消"
    exit 0
  fi
  
  echo ""
  echo "🔧 正在清理端口..."
  
  for info in "${OCCUPIED_INFO[@]}"; do
    IFS='|' read -r port type pids cmd <<< "$info"
    
    for pid in $pids; do
      if kill -9 "$pid" 2>/dev/null; then
        KILLED_COUNT=$((KILLED_COUNT + 1))
      fi
    done
    
    echo "  ✅ 端口 $port 已清理"
  done
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ 完成: 扫描了 $TOTAL_CHECKED 个端口，清理了 $KILLED_COUNT 个进程"