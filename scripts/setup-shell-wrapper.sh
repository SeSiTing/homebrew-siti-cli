#!/bin/bash

# siti-cli Shell 包装函数安装脚本
# 使需要修改环境变量的命令在当前终端立即生效

set -e

WRAPPER_MARKER="# siti shell wrapper - auto-generated"

# 颜色定义
COLOR_GREEN="\033[32m"
COLOR_YELLOW="\033[33m"
COLOR_RED="\033[31m"
COLOR_RESET="\033[0m"

print_success() {
  echo -e "${COLOR_GREEN}✅ $1${COLOR_RESET}"
}

print_warning() {
  echo -e "${COLOR_YELLOW}⚠️  $1${COLOR_RESET}"
}

print_error() {
  echo -e "${COLOR_RED}❌ $1${COLOR_RESET}"
}

# 检测 shell 类型
detect_shell() {
  if [ -n "$ZSH_VERSION" ]; then
    echo "zsh"
  elif [ -n "$BASH_VERSION" ]; then
    echo "bash"
  else
    echo "unknown"
  fi
}

# 获取配置文件路径
get_config_file() {
  local shell_type="$1"
  
  # 优先检查实际存在的配置文件
  if [ -f "$HOME/.zshrc" ]; then
    echo "$HOME/.zshrc"
    return 0
  elif [ -f "$HOME/.bashrc" ]; then
    echo "$HOME/.bashrc"
    return 0
  fi
  
  # 如果都不存在，根据 shell 类型返回默认值
  case "$shell_type" in
    zsh)
      echo "$HOME/.zshrc"
      ;;
    bash)
      echo "$HOME/.bashrc"
      ;;
    *)
      return 1
      ;;
  esac
}

# 检查是否已安装
check_installed() {
  local config_file="$1"
  
  if [ -f "$config_file" ] && grep -q "$WRAPPER_MARKER" "$config_file"; then
    return 0
  else
    return 1
  fi
}

# 安装包装函数
install_wrapper() {
  local config_file="$1"
  
  # 检查是否已安装
  if check_installed "$config_file"; then
    print_warning "siti shell wrapper 已安装"
    echo "配置文件: $config_file"
    return 0
  fi
  
  # 备份配置文件
  cp "$config_file" "${config_file}.backup.$(date +%Y%m%d_%H%M%S)"
  
  # 添加包装函数
  cat >> "$config_file" << 'EOF'

# siti shell wrapper - auto-generated
# 使需要修改环境变量的命令（如 proxy、ai）在当前终端立即生效
# 基于退出码约定：exit 10 表示需要 eval
siti() {
  local output
  local exit_code
  
  # 执行命令并捕获输出和退出码
  output=$(command siti "$@" 2>&1)
  exit_code=$?
  
  # 退出码 10 表示需要 eval（修改当前 shell 环境）
  if [ $exit_code -eq 10 ]; then
    eval "$output"
    return 0
  else
    # 其他情况直接显示输出
    echo "$output"
    return $exit_code
  fi
}
EOF
  
  print_success "siti shell wrapper 已安装到 $config_file"
  echo ""
  echo "请运行以下命令使其生效:"
  echo "  source $config_file"
  echo ""
  echo "或者重新打开终端"
}

# 卸载包装函数
uninstall_wrapper() {
  local config_file="$1"
  
  if ! check_installed "$config_file"; then
    print_warning "siti shell wrapper 未安装"
    return 0
  fi
  
  # 备份配置文件
  cp "$config_file" "${config_file}.backup.$(date +%Y%m%d_%H%M%S)"
  
  # 删除包装函数（从标记行到 } 结束）
  sed -i.tmp "/$WRAPPER_MARKER/,/^}$/d" "$config_file"
  rm -f "${config_file}.tmp"
  
  print_success "siti shell wrapper 已卸载"
  echo "请运行: source $config_file"
}

# 显示帮助
show_help() {
  cat << EOF
siti-cli Shell 包装函数安装脚本

用法:
  ./setup-shell-wrapper.sh [选项]

选项:
  install     安装 shell 包装函数（默认）
  uninstall   卸载 shell 包装函数
  status      查看安装状态
  --help, -h  显示帮助信息

说明:
  安装 shell 包装函数后，以下命令将在当前终端立即生效：
  - siti proxy on/off    （代理开关）
  - siti ai switch       （AI 配置切换）
  - 未来所有使用 exit 10 的命令

示例:
  ./setup-shell-wrapper.sh install    # 安装
  ./setup-shell-wrapper.sh status     # 查看状态
  ./setup-shell-wrapper.sh uninstall  # 卸载
EOF
}

# 显示状态
show_status() {
  local shell_type=$(detect_shell)
  local config_file=$(get_config_file "$shell_type")
  
  echo "Shell 类型: $shell_type"
  echo "配置文件: $config_file"
  echo ""
  
  if check_installed "$config_file"; then
    print_success "siti shell wrapper 已安装"
    echo ""
    echo "支持的命令:"
    echo "  • siti proxy on/off    - 代理管理"
    echo "  • siti ai switch       - AI 配置切换"
  else
    print_warning "siti shell wrapper 未安装"
    echo ""
    echo "安装后可以直接使用以下命令（无需 eval）:"
    echo "  siti proxy on          # 代理开启"
    echo "  siti ai switch minimax # AI 配置切换"
    echo ""
    echo "运行以下命令安装:"
    echo "  ./setup-shell-wrapper.sh install"
  fi
}

# 主函数
main() {
  local action="${1:-install}"
  
  # 检测 shell 类型
  local shell_type=$(detect_shell)
  if [ "$shell_type" = "unknown" ]; then
    print_error "不支持的 shell 类型"
    echo "仅支持 bash 和 zsh"
    exit 1
  fi
  
  # 获取配置文件
  local config_file=$(get_config_file "$shell_type")
  if [ ! -f "$config_file" ]; then
    print_error "配置文件不存在: $config_file"
    exit 1
  fi
  
  # 执行操作
  case "$action" in
    install)
      install_wrapper "$config_file"
      ;;
    uninstall)
      uninstall_wrapper "$config_file"
      ;;
    status)
      show_status
      ;;
    --help|-h|help)
      show_help
      ;;
    *)
      print_error "未知操作: $action"
      echo ""
      show_help
      exit 1
      ;;
  esac
}

main "$@"
