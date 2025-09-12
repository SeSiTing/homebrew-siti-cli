#!/bin/bash

# 默认端口范围（8000～8010 和 8080～8090）
DEFAULT_PORTS=($(seq 2024 2030) $(seq 8000 8010) $(seq 8080 8090) $(seq 9000 9010))

MODE="kill"

if [ "$1" == "check" ]; then
  MODE="check"
  shift
fi

if [ $# -eq 0 ]; then
  PORTS=("${DEFAULT_PORTS[@]}")
else
  PORTS=("$@")
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