# siti-cli

> 🚀 个人命令行工具集，简化日常开发操作

[![GitHub](https://img.shields.io/badge/GitHub-siti--cli-blue?logo=github)](https://github.com/SeSiTing/homebrew-siti-cli)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## ✨ 核心功能

- 🔄 **AI 配置管理** - 快速切换 AI 服务商（MiniMax、智谱、OpenRouter 等）
- 🌐 **代理管理** - 一键开关终端代理
- 🔌 **端口管理** - 快速释放占用的端口
- 🛠️ **实用工具** - 网络检测、IP 显示、日志清理等

## 📦 安装

siti-cli 提供两种安装方式，根据你的使用场景选择：

### 方式 A：Homebrew 安装（推荐）

适合喜欢用包管理器统一管理软件的用户，也适合 CI/CD 和自动化场景。

```bash
brew tap SeSiTing/siti-cli
brew install siti-cli
```

**配置 Shell Wrapper（推荐）**

为了让 `siti ai switch` 和 `siti proxy` 等命令能在当前终端立即生效，需要配置 shell wrapper：

```bash
# 查看配置内容（可选）
siti init zsh

# 添加到 shell 配置
eval "$(siti init zsh)" >> ~/.zshrc
source ~/.zshrc
```

**特点**：
- ✅ 包管理器统一管理，卸载干净
- ✅ 版本化更新，稳定可靠
- ✅ 适合 CI/CD 和自动化场景
- ✅ 可选择是否配置 wrapper
- ℹ️  需要手动配置 wrapper（或自动配置可能失败）

### 方式 B：独立安装脚本

适合日常开发使用，提供最佳用户体验。

```bash
curl -fsSL https://raw.githubusercontent.com/SeSiTing/homebrew-siti-cli/main/install.sh | bash
```

安装过程中会**交互式询问**是否安装 shell wrapper，选择 `y` 即可自动配置。

**非交互式安装**（用于自动化脚本）：
```bash
curl -fsSL https://raw.githubusercontent.com/SeSiTing/homebrew-siti-cli/main/install.sh | bash -s -- --unattended
```

**特点**：
- ✅ 一键安装，交互式配置
- ✅ 自动安装 shell wrapper
- ✅ 最佳用户体验
- ✅ 支持非交互模式
- ℹ️  不适合需要包管理器统一管理的场景

### 安装方式对比

| 特性 | Homebrew | 独立脚本 |
|------|----------|---------|
| **安装命令** | `brew install siti-cli` | `curl ... \| bash` |
| **Wrapper 配置** | 需手动 `siti init` | 自动配置（可选） |
| **更新方式** | `brew upgrade` 或 `siti upgrade` | `siti upgrade` |
| **卸载方式** | `brew uninstall siti-cli` | 删除 `~/.siti-cli` 和配置 |
| **适用场景** | 包管理、CI/CD、自动化 | 日常开发、快速上手 |
| **安装位置** | `/opt/homebrew` 或 `/usr/local` | `~/.siti-cli` |

### 更新

无论哪种安装方式，都可以使用统一的更新命令：

```bash
siti upgrade
```

命令会自动检测安装方式并执行相应的更新操作。

### 手动安装

如果你想从源码安装或进行开发：

```bash
git clone https://github.com/SeSiTing/homebrew-siti-cli.git ~/.siti-cli
echo 'export PATH="$HOME/.siti-cli/bin:$PATH"' >> ~/.zshrc
eval "$(~/.siti-cli/bin/siti init zsh)" >> ~/.zshrc
source ~/.zshrc
```

**详细说明：** 查看 [安装指南](docs/INSTALL.md)

## 🎯 快速开始

```bash
# 查看所有命令
siti --help

# AI 配置管理
siti ai list              # 列出所有 AI 服务商
siti ai switch minimax    # 切换到 MiniMax（当前终端生效）
siti ai switch minimax --persist  # 持久化切换（修改 ~/.zshrc）
siti ai current           # 查看当前配置

# 代理管理
siti proxy on             # 开启代理
siti proxy off            # 关闭代理

# 端口管理
siti killports 3000       # 释放 3000 端口
siti killports 3000-3010  # 释放端口范围

# 升级
siti upgrade              # 升级到最新版本
```

### AI 配置说明

`siti ai` 命令通过读取 `~/.zshrc` 中的环境变量来管理 AI 服务商配置：

**环境变量命名规范**:
```bash
# 格式：<PROVIDER>_BASE_URL 和 <PROVIDER>_API_KEY（可选）
export MINIMAX_BASE_URL="https://api.minimaxi.com/anthropic"
export MINIMAX_API_KEY="your-api-key"  # 可选

export ZHIPU_BASE_URL="https://open.bigmodel.cn/api/anthropic"
export ZHIPU_API_KEY="your-api-key"

# 如果没有 API_KEY，会使用 DEFAULT_AUTH_TOKEN 兜底
export DEFAULT_AUTH_TOKEN="default-token"
```

**多环境配置建议**:
```bash
# ✅ 推荐：使用语义化名称
export LLMS_BASE_URL="https://llms-test.blacklake.tech"
export DEV8_BASE_URL="http://10.83.20.125:3009"
export DEV9_BASE_URL="http://10.83.20.127:3009"

# 使用
siti ai switch llms
siti ai switch dev8
```

**详细使用：** 查看 [快速开始](docs/QUICK_START.md)

## 🌟 特色

- ✅ **零配置** - 自动发现 AI 服务商配置
- ✅ **立即生效** - 命令在当前终端立即生效
- ✅ **易于扩展** - 支持自定义命令
- ✅ **跨平台** - 支持 macOS 和 Linux

## 📚 文档

- [快速开始](docs/QUICK_START.md) - 5 分钟上手指南
- [安装指南](docs/INSTALL.md) - 详细安装说明和对比
- [更新日志](CHANGELOG.md) - 版本历史

## 🤝 贡献

欢迎贡献代码、报告问题或提出建议！

- 报告问题：[GitHub Issues](https://github.com/SeSiTing/homebrew-siti-cli/issues)
- 贡献代码：[Pull Requests](https://github.com/SeSiTing/homebrew-siti-cli/pulls)

## 📄 许可证

MIT License - 查看 [LICENSE](LICENSE) 文件了解详情

---

**Star ⭐ 如果你觉得有用！**
