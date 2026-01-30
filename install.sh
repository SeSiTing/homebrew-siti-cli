#!/bin/bash

# siti-cli 一键安装脚本
# 使用方式: curl -fsSL https://raw.githubusercontent.com/SeSiTing/siti-cli/main/install.sh | bash

set -e

# 颜色定义
COLOR_GREEN="\033[32m"
COLOR_YELLOW="\033[33m"
COLOR_RED="\033[31m"
COLOR_BLUE="\033[34m"
COLOR_BOLD="\033[1m"
COLOR_RESET="\033[0m"

print_success() {
  echo -e "${COLOR_GREEN}✅ $1${COLOR_RESET}"
}

print_info() {
  echo -e "${COLOR_BLUE}ℹ️  $1${COLOR_RESET}"
}

print_warning() {
  echo -e "${COLOR_YELLOW}⚠️  $1${COLOR_RESET}"
}

print_error() {
  echo -e "${COLOR_RED}❌ $1${COLOR_RESET}"
}

print_header() {
  echo -e "${COLOR_BOLD}${COLOR_BLUE}$1${COLOR_RESET}"
}

# 检测操作系统
detect_os() {
  case "$(uname -s)" in
    Darwin*)
      echo "macos"
      ;;
    Linux*)
      echo "linux"
      ;;
    *)
      echo "unknown"
      ;;
  esac
}

# 检测 shell 类型
detect_shell() {
  if [ -n "$ZSH_VERSION" ]; then
    echo "zsh"
  elif [ -n "$BASH_VERSION" ]; then
    echo "bash"
  else
    # 检查默认 shell
    case "$SHELL" in
      */zsh)
        echo "zsh"
        ;;
      */bash)
        echo "bash"
        ;;
      *)
        echo "unknown"
        ;;
    esac
  fi
}

# 获取配置文件路径
get_config_file() {
  if [ -f "$HOME/.zshrc" ]; then
    echo "$HOME/.zshrc"
  elif [ -f "$HOME/.bashrc" ]; then
    echo "$HOME/.bashrc"
  else
    local shell_type=$(detect_shell)
    case "$shell_type" in
      zsh)
        echo "$HOME/.zshrc"
        ;;
      bash)
        echo "$HOME/.bashrc"
        ;;
      *)
        echo "$HOME/.profile"
        ;;
    esac
  fi
}

# 询问用户
ask_user() {
  local prompt="$1"
  local default="${2:-y}"
  
  if [ "$default" = "y" ]; then
    prompt="$prompt [Y/n] "
  else
    prompt="$prompt [y/N] "
  fi
  
  read -p "$prompt" response
  response=${response:-$default}
  
  case "$response" in
    [yY][eE][sS]|[yY])
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

# 安装 siti-cli
install_siti() {
  local install_dir="$HOME/.siti-cli"
  local bin_dir="$HOME/.local/bin"
  
  print_header "📦 安装 siti-cli..."
  echo ""
  
  # 创建目录
  mkdir -p "$bin_dir"
  
  # 克隆或更新仓库
  if [ -d "$install_dir" ]; then
    print_info "检测到已安装，正在更新..."
    cd "$install_dir"
    git pull origin main
  else
    print_info "正在下载 siti-cli..."
    git clone https://github.com/SeSiTing/siti-cli.git "$install_dir"
  fi
  
  # 创建符号链接
  ln -sf "$install_dir/bin/siti" "$bin_dir/siti"
  chmod +x "$install_dir/bin/siti"
  
  print_success "siti-cli 已安装到 $install_dir"
  echo ""
}

# 添加到 PATH
setup_path() {
  local config_file=$(get_config_file)
  local bin_dir="$HOME/.local/bin"
  
  # 检查是否已添加
  if grep -q "$bin_dir" "$config_file" 2>/dev/null; then
    print_info "PATH 已配置"
    return 0
  fi
  
  print_info "添加 siti-cli 到 PATH..."
  
  # 添加到配置文件
  cat >> "$config_file" << EOF

# siti-cli
export PATH="\$HOME/.local/bin:\$PATH"
EOF
  
  # 立即生效
  export PATH="$bin_dir:$PATH"
  
  print_success "PATH 已配置"
  echo ""
}

# 安装 shell 包装函数
setup_wrapper() {
  local config_file=$(get_config_file)
  local wrapper_marker="# siti shell wrapper - auto-generated"
  
  # 检查是否已安装
  if grep -q "$wrapper_marker" "$config_file" 2>/dev/null; then
    print_info "Shell 包装函数已安装"
    return 0
  fi
  
  print_info "安装 shell 包装函数..."
  
  # 备份配置文件
  cp "$config_file" "${config_file}.backup.$(date +%Y%m%d_%H%M%S)"
  
  # 添加包装函数
  cat >> "$config_file" << 'EOF'

# siti shell wrapper - auto-generated
# 使需要修改环境变量的命令（如 proxy、ai）在当前终端立即生效
siti() {
  local output
  local exit_code
  
  output=$(command siti "$@" 2>&1)
  exit_code=$?
  
  if [ $exit_code -eq 10 ]; then
    eval "$output"
    return 0
  else
    echo "$output"
    return $exit_code
  fi
}
EOF
  
  print_success "Shell 包装函数已安装"
  echo ""
}

# 显示完成信息
show_completion() {
  local config_file=$(get_config_file)
  
  echo ""
  print_header "🎉 安装完成！"
  echo ""
  
  print_info "请运行以下命令使配置生效："
  echo -e "  ${COLOR_BOLD}source $config_file${COLOR_RESET}"
  echo ""
  
  print_info "或者重新打开终端"
  echo ""
  
  print_header "📚 快速开始："
  echo ""
  echo "  # 查看帮助"
  echo "  siti --help"
  echo ""
  echo "  # AI 配置管理"
  echo "  siti ai list            # 列出所有 AI 服务商"
  echo "  siti ai switch minimax  # 切换到 MiniMax"
  echo "  siti ai current         # 查看当前配置"
  echo ""
  echo "  # 代理管理"
  echo "  siti proxy on           # 开启代理"
  echo "  siti proxy off          # 关闭代理"
  echo ""
  
  print_info "更多信息: https://github.com/SeSiTing/siti-cli"
  echo ""
}

# 主函数
main() {
  local os=$(detect_os)
  local shell_type=$(detect_shell)
  
  # 显示欢迎信息
  clear
  print_header "╔════════════════════════════════════════╗"
  print_header "║      siti-cli 安装程序                 ║"
  print_header "║  个人命令行工具集                      ║"
  print_header "╚════════════════════════════════════════╝"
  echo ""
  
  print_info "操作系统: $os"
  print_info "Shell: $shell_type"
  echo ""
  
  # 检查依赖
  if ! command -v git &> /dev/null; then
    print_error "未找到 git，请先安装 git"
    exit 1
  fi
  
  # 安装 siti-cli
  install_siti
  
  # 设置 PATH
  setup_path
  
  # 询问是否安装 shell 包装函数
  echo ""
  print_header "🔧 Shell 包装函数设置"
  echo ""
  print_info "Shell 包装函数可以让以下命令在当前终端立即生效："
  echo "  • siti proxy on/off    - 代理管理"
  echo "  • siti ai switch       - AI 配置切换"
  echo ""
  
  if ask_user "是否安装 shell 包装函数？" "y"; then
    echo ""
    setup_wrapper
  else
    echo ""
    print_warning "跳过 shell 包装函数安装"
    print_info "你仍然可以使用 siti-cli，但需要手动 eval："
    echo "  eval \"\$(siti proxy on)\""
    echo "  eval \"\$(siti ai switch minimax)\""
    echo ""
    print_info "稍后可以运行以下命令安装："
    echo "  ~/.siti-cli/scripts/setup-shell-wrapper.sh install"
    echo ""
  fi
  
  # 显示完成信息
  show_completion
}

# 运行主函数
main "$@"
