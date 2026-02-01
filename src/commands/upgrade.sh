#!/bin/bash

# 描述: 升级 siti-cli 到最新版本

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

# 获取当前版本
CURRENT_VERSION="${VERSION:-unknown}"

# 检测安装方式（从父脚本传递的 INSTALL_METHOD）
if [ -z "$INSTALL_METHOD" ]; then
  # 如果没有传递，尝试自动检测
  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  if [[ "$SCRIPT_DIR" =~ ^(/opt/homebrew|/usr/local)/share/siti-cli ]]; then
    INSTALL_METHOD="homebrew"
  elif [[ -d "$HOME/.siti-cli" ]] && [[ -L "$HOME/.local/bin/siti" ]]; then
    INSTALL_METHOD="standalone"
  else
    INSTALL_METHOD="source"
  fi
fi

echo ""
print_header "🚀 siti-cli 升级工具"
echo ""
print_info "当前版本: v${CURRENT_VERSION}"
print_info "安装方式: ${INSTALL_METHOD}"
echo ""

case "$INSTALL_METHOD" in
  homebrew)
    print_header "📦 通过 Homebrew 更新..."
    echo ""
    
    # 检查 brew 是否可用
    if ! command -v brew &> /dev/null; then
      print_error "未找到 Homebrew，请手动更新"
      exit 1
    fi
    
    # 更新 Homebrew
    print_info "正在更新 Homebrew..."
    if brew update; then
      print_success "Homebrew 更新完成"
    else
      print_warning "Homebrew 更新失败，继续尝试升级 siti-cli"
    fi
    
    echo ""
    
    # 升级 siti-cli
    print_info "正在升级 siti-cli..."
    if brew upgrade siti-cli 2>&1 | tee /tmp/brew-upgrade.log; then
      if grep -q "already installed" /tmp/brew-upgrade.log; then
        print_info "siti-cli 已是最新版本"
      else
        print_success "siti-cli 升级完成！"
        echo ""
        print_info "运行 'source ~/.zshrc' 或重新打开终端使新版本生效"
      fi
    else
      print_error "升级失败，请查看错误信息"
      rm -f /tmp/brew-upgrade.log
      exit 1
    fi
    
    rm -f /tmp/brew-upgrade.log
    ;;
    
  standalone)
    print_header "🔄 通过 Git 更新..."
    echo ""
    
    # 检查安装目录
    INSTALL_DIR="$HOME/.siti-cli"
    if [ ! -d "$INSTALL_DIR" ]; then
      print_error "未找到安装目录: $INSTALL_DIR"
      print_info "请使用以下命令重新安装："
      echo "  curl -fsSL https://raw.githubusercontent.com/SeSiTing/siti-cli/main/install.sh | bash"
      exit 1
    fi
    
    # 检查是否是 Git 仓库
    if [ ! -d "$INSTALL_DIR/.git" ]; then
      print_error "$INSTALL_DIR 不是 Git 仓库"
      print_info "请使用以下命令重新安装："
      echo "  curl -fsSL https://raw.githubusercontent.com/SeSiTing/siti-cli/main/install.sh | bash"
      exit 1
    fi
    
    # 进入安装目录
    cd "$INSTALL_DIR"
    
    # 检查是否有未提交的更改
    if ! git diff-index --quiet HEAD -- 2>/dev/null; then
      print_warning "检测到本地修改，将尝试保留这些更改"
      echo ""
    fi
    
    # 拉取最新代码
    print_info "正在从 GitHub 拉取最新版本..."
    if git pull --rebase --autostash origin main; then
      print_success "更新完成！"
      
      # 显示更新日志
      echo ""
      print_header "📝 更新内容："
      git log --oneline --no-merges HEAD@{1}..HEAD 2>/dev/null || true
      
      echo ""
      print_info "运行 'source ~/.zshrc' 或重新打开终端使新版本生效"
    else
      print_error "更新失败"
      echo ""
      print_info "可能的原因："
      echo "  • 网络连接问题"
      echo "  • Git 冲突"
      echo "  • 本地修改冲突"
      echo ""
      print_info "尝试手动更新："
      echo "  cd ~/.siti-cli"
      echo "  git stash  # 保存本地修改"
      echo "  git pull origin main"
      echo "  git stash pop  # 恢复本地修改"
      exit 1
    fi
    ;;
    
  source)
    print_warning "检测到开发模式安装"
    echo ""
    print_info "请手动更新："
    echo "  cd $(dirname "$(dirname "$SCRIPT_DIR")")"
    echo "  git pull origin main"
    echo ""
    ;;
    
  *)
    print_error "无法识别的安装方式: $INSTALL_METHOD"
    echo ""
    print_info "请选择合适的更新方式："
    echo "  • Homebrew: brew upgrade siti-cli"
    echo "  • 独立安装: cd ~/.siti-cli && git pull"
    echo "  • 源码模式: cd <项目目录> && git pull"
    exit 1
    ;;
esac

echo ""
print_success "升级流程完成！"
echo ""
