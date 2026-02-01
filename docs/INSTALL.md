# siti-cli 安装指南

## 🚀 快速安装

### 推荐方式：一键安装

```bash
curl -fsSL https://raw.githubusercontent.com/SeSiTing/homebrew-siti-cli/main/install.sh | bash
source ~/.zshrc
```

**安装完成后运行：**
```bash
siti --version
siti ai list
```

---

## 📦 安装方式对比

| 方式 | 命令 | 难度 | 更新 | 适合 |
|------|------|------|------|------|
| **一键安装** | 1 条命令 | ⭐ | 重新运行 | 所有用户 ⭐ |
| **Homebrew** | 2 条命令 | ⭐⭐ | `brew upgrade` | macOS 用户 |
| **手动安装** | 3+ 条命令 | ⭐⭐⭐ | `git pull` | 高级用户 |

---

## 📝 详细安装步骤

### 方式 1: 一键安装（推荐）

```bash
# 1. 运行安装脚本
curl -fsSL https://raw.githubusercontent.com/SeSiTing/homebrew-siti-cli/main/install.sh | bash

# 2. 使配置生效
source ~/.zshrc

# 3. 验证安装
siti --version
```

**会做什么？**
- ✅ 克隆仓库到 `~/.siti-cli`
- ✅ 创建符号链接到 `~/.local/bin/siti`
- ✅ 添加 PATH 到 `~/.zshrc`
- ✅ 询问是否安装 shell 包装函数
- ✅ 自动备份配置文件

**会修改什么？**

添加到 `~/.zshrc`：
```bash
# PATH 配置
export PATH="$HOME/.local/bin:$PATH"

# Shell 包装函数（如果选择安装）
siti() {
  local output=$(command siti "$@" 2>&1)
  local exit_code=$?
  [ $exit_code -eq 10 ] && eval "$output" || echo "$output"
  return $exit_code
}
```

---

### 方式 2: Homebrew

```bash
# 1. 添加 tap（仅首次）
brew tap SeSiTing/siti-cli

# 2. 安装
brew install siti-cli

# 3. 安装 shell 包装函数（推荐）
~/.siti-cli/scripts/setup-shell-wrapper.sh install
source ~/.zshrc
```

**更新：**
```bash
brew upgrade siti-cli
```

**卸载：**
```bash
brew uninstall siti-cli
```

---

### 方式 3: 手动安装

```bash
# 1. 克隆仓库
git clone https://github.com/SeSiTing/homebrew-siti-cli.git ~/.siti-cli

# 2. 添加到 PATH
echo 'export PATH="$HOME/.siti-cli/bin:$PATH"' >> ~/.zshrc

# 3. 安装 shell 包装函数（推荐）
~/.siti-cli/scripts/setup-shell-wrapper.sh install

# 4. 使配置生效
source ~/.zshrc
```

**更新：**
```bash
cd ~/.siti-cli && git pull
```

**卸载：**
```bash
rm -rf ~/.siti-cli
# 然后编辑 ~/.zshrc 删除相关配置
```

---

## ⚙️ Shell 包装函数

### 什么是 Shell 包装函数？

让命令在**当前终端立即生效**，无需手动 eval。

**安装后：**
```bash
siti proxy on           # ✅ 立即生效
siti ai switch minimax  # ✅ 立即生效
```

**不安装：**
```bash
eval "$(siti proxy on)"           # ⚠️ 需要 eval
eval "$(siti ai switch minimax)"  # ⚠️ 需要 eval
```

### 管理包装函数

```bash
# 查看状态
~/.siti-cli/scripts/setup-shell-wrapper.sh status

# 安装
~/.siti-cli/scripts/setup-shell-wrapper.sh install

# 卸载
~/.siti-cli/scripts/setup-shell-wrapper.sh uninstall
```

---

## 🔄 更新

```bash
# 一键安装方式
curl -fsSL https://raw.githubusercontent.com/SeSiTing/homebrew-siti-cli/main/install.sh | bash

# Homebrew
brew upgrade siti-cli

# 手动安装
cd ~/.siti-cli && git pull
```

---

## 🗑️ 卸载

### 完全卸载

```bash
# 1. 删除文件
rm -rf ~/.siti-cli
rm -f ~/.local/bin/siti

# 2. 编辑 ~/.zshrc，删除以下内容：
#    - export PATH="$HOME/.local/bin:$PATH"
#    - siti shell wrapper 相关代码

# 3. 重新加载
source ~/.zshrc
```

### Homebrew 卸载

```bash
brew uninstall siti-cli
# 如果安装了包装函数，需要手动从 ~/.zshrc 删除
```

---

## ❓ 常见问题

### Q: 安装后提示 "command not found: siti"

**A:** 运行 `source ~/.zshrc` 或重新打开终端

### Q: 如何验证安装成功？

**A:** 
```bash
which siti
siti --version
siti --help
```

### Q: Shell 包装函数必须安装吗？

**A:** 不是必须的，但强烈推荐。不安装需要手动 eval。

### Q: 支持哪些系统？

**A:** 
- ✅ macOS
- ✅ Linux
- ✅ 支持 zsh 和 bash

### Q: 安全吗？

**A:** 
- ✅ 开源代码，可审查
- ✅ 仅修改用户目录
- ✅ 不需要 sudo
- ✅ 可以完全卸载

---

## 📚 下一步

安装完成后，查看 [快速开始](QUICK_START.md) 了解如何使用。
