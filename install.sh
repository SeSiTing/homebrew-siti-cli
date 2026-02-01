#!/bin/bash

# siti-cli 一键安装脚本
# 使用方式: curl -fsSL https://raw.githubusercontent.com/SeSiTing/siti-cli/main/install.sh | bash
#
# 参数:
#   --unattended    非交互模式，自动安装所有组件
#   --skip-wrapper  跳过 shell wrapper 安装
#
# 示例:
#   # 交互式安装（默认）
#   curl -fsSL https://raw.githubusercontent.com/SeSiTing/siti-cli/main/install.sh | bash
#
#   # 非交互式安装
#   curl -fsSL https://raw.githubusercontent.com/SeSiTing/siti-cli/main/install.sh | bash -s -- --unattended

set -e

# 颜色定义
COLOR_GREEN="\033[32m"
COLOR_YELLOW="\033[33m"
COLOR_RED="\033[31m"
COLOR_BLUE="\033[34m"
COLOR_BOLD="\033[1m"
COLOR_RESET="\033[0m"

# 默认选项
UNATTENDED=false
SKIP_WRAPPER=false

# 解析参数
while [[ $# -gt 0 ]]; do
  case $1 in
    --unattended)
      UNATTENDED=true
      shift
      ;;
    --skip-wrapper)
      SKIP_WRAPPER=true
      shift
      ;;
    --help|-h)
      echo "siti-cli 安装脚本"
      echo ""
      echo "用法:"
      echo "  curl -fsSL https://raw.githubusercontent.com/SeSiTing/siti-cli/main/install.sh | bash"
      echo ""
      echo "参数:"
      echo "  --unattended     非交互模式，自动安装所有组件"
      echo "  --skip-wrapper   跳过 shell wrapper 安装"
      echo "  --help           显示帮助信息"
      echo ""
      echo "示例:"
      echo "  # 交互式安装"
      echo "  curl -fsSL https://raw.githubusercontent.com/SeSiTing/siti-cli/main/install.sh | bash"
      echo ""
      echo "  # 非交互式安装"
      echo "  curl -fsSL https://raw.githubusercontent.com/SeSiTing/siti-cli/main/install.sh | bash -s -- --unattended"
      exit 0
      ;;
    *)
      echo "未知参数: $1"
      echo "运行 'bash install.sh --help' 查看帮助"
      exit 1
      ;;
  esac
done

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
  
  # 非交互模式直接返回默认值
  if [ "$UNATTENDED" = true ]; then
    return 0
  fi
  
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
    # 检查并修复错误的 remote URL
    current_remote=$(git config --get remote.origin.url 2>/dev/null || echo "")
    if [[ "$current_remote" == *"SeSiTing/siti-cli.git"* ]]; then
      print_warning "检测到旧仓库地址，正在更新..."
      git remote set-url origin https://github.com/SeSiTing/homebrew-siti-cli.git
      print_success "仓库地址已更新"
    fi
    git pull origin main
  else
    print_info "正在下载 siti-cli..."
    git clone https://github.com/SeSiTing/homebrew-siti-cli.git "$install_dir"
  fi
  
  # 创建符号链接
  ln -sf "$install_dir/bin/siti" "$bin_dir/siti"
  chmod +x "$install_dir/bin/siti"
  
  # 记录安装方式
  echo "standalone" > "$install_dir/.install-source"
  
  # 创建用户数据目录（统一使用 ~/.siti-cli）
  mkdir -p "$install_dir"/{commands,logs,cache,config}
  
  # 迁移旧目录 ~/.siti 到 ~/.siti-cli
  if [ -d "$HOME/.siti" ]; then
    print_info "检测到旧版目录 ~/.siti，正在迁移到 ~/.siti-cli..."
    backup_dir="$HOME/.siti.backup.$(date +%Y%m%d%H%M%S)"
    if cp -R "$HOME/.siti" "$backup_dir" 2>/dev/null; then
      for subdir in commands config logs cache; do
        if [ -d "$HOME/.siti/$subdir" ] && [ "$(ls -A "$HOME/.siti/$subdir" 2>/dev/null)" ]; then
          cp -R "$HOME/.siti/$subdir"/* "$install_dir/$subdir/" 2>/dev/null || true
        fi
      done
      rm -rf "$HOME/.siti"
      print_success "已迁移并删除旧目录 ~/.siti（备份: $backup_dir）"
    else
      print_warning "无法备份 ~/.siti，跳过迁移"
    fi
  fi
  
  # 创建默认配置文件（若不存在）
  if [ ! -f "$install_dir/config/siti.conf" ]; then
    cat > "$install_dir/config/siti.conf" << EOF
# siti-cli 配置文件
LOG_LEVEL="info"
LOG_FILE="$HOME/.siti-cli/logs/siti.log"
CACHE_DIR="$HOME/.siti-cli/cache"
USER_COMMANDS_DIR="$HOME/.siti-cli/commands"
EOF
  fi
  
  # 创建示例命令（若不存在）
  if [ ! -f "$install_dir/commands/hello.sh" ]; then
    cat > "$install_dir/commands/hello.sh" << 'HELLO_EOF'
#!/bin/bash
# 描述: 示例用户自定义命令
name="${1:-World}"
echo "Hello, $name! 这是一个用户自定义命令示例。"
HELLO_EOF
    chmod +x "$install_dir/commands/hello.sh"
  fi
  
  print_success "siti-cli 已安装到 $install_dir"
  echo ""
}

# 去除重复的 siti-cli PATH 配置
cleanup_duplicates() {
  local config_file=$(get_config_file)
  local path_marker="# siti-cli PATH configuration - auto-generated"
  
  # 已使用新标记则只清理旧块
  if grep -q "$path_marker" "$config_file" 2>/dev/null; then
    while grep -q "^# siti-cli$" "$config_file" 2>/dev/null; do
      cp "$config_file" "${config_file}.bak.$$"
      sed '/^# siti-cli$/,/^export PATH=.*$/d' "${config_file}.bak.$$" > "$config_file"
      rm -f "${config_file}.bak.$$"
    done
    return 0
  fi
  
  local count=$(grep -c "^# siti-cli$" "$config_file" 2>/dev/null || echo "0")
  if [ "$count" -gt 1 ]; then
    print_warning "检测到 $count 个重复的 siti-cli PATH 配置，正在清理..."
    cp "$config_file" "${config_file}.backup.$(date +%Y%m%d_%H%M%S)"
    while grep -q "^# siti-cli$" "$config_file" 2>/dev/null; do
      cp "$config_file" "${config_file}.bak.$$"
      sed '/^# siti-cli$/,/^export PATH=.*$/d' "${config_file}.bak.$$" > "$config_file"
      rm -f "${config_file}.bak.$$"
    done
    cat >> "$config_file" << 'EOF'

# siti-cli PATH configuration - auto-generated
export PATH="$HOME/.local/bin:$PATH"
EOF
    print_success "重复配置已清理"
  fi
}

# 添加到 PATH
setup_path() {
  local config_file=$(get_config_file)
  local bin_dir="$HOME/.local/bin"
  local path_marker="# siti-cli PATH configuration - auto-generated"
  
  # 检查是否已添加（使用唯一标记）
  if grep -q "$path_marker" "$config_file" 2>/dev/null; then
    print_info "PATH 已配置"
    return 0
  fi
  
  print_info "添加 siti-cli 到 PATH..."
  
  # 添加到配置文件
  cat >> "$config_file" << EOF

$path_marker
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
  
  print_info "更多信息: https://github.com/SeSiTing/homebrew-siti-cli"
  echo ""
}

# 主函数
main() {
  local os=$(detect_os)
  local shell_type=$(detect_shell)
  
  # 显示欢迎信息
  if [ "$UNATTENDED" != true ]; then
    clear
  fi
  print_header "╔════════════════════════════════════════╗"
  print_header "║      siti-cli 安装程序                 ║"
  print_header "║  个人命令行工具集                      ║"
  print_header "╚════════════════════════════════════════╝"
  echo ""
  
  print_info "操作系统: $os"
  print_info "Shell: $shell_type"
  if [ "$UNATTENDED" = true ]; then
    print_info "模式: 非交互式（--unattended）"
  fi
  echo ""
  
  # 检查依赖
  if ! command -v git &> /dev/null; then
    print_error "未找到 git，请先安装 git"
    exit 1
  fi
  
  # 安装 siti-cli
  install_siti
  
  # 清理重复的 PATH 配置（如有）
  cleanup_duplicates
  
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
  
  # 跳过 wrapper 安装
  if [ "$SKIP_WRAPPER" = true ]; then
    print_warning "跳过 shell 包装函数安装（--skip-wrapper）"
    echo ""
  elif ask_user "是否安装 shell 包装函数？" "y"; then
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
    echo "  eval \"\$(siti init zsh)\" >> ~/.zshrc"
    echo ""
  fi
  
  # 显示完成信息
  show_completion
}

# 运行主函数
main "$@"
