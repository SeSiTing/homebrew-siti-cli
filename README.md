# siti-cli

> 🚀 个人命令行工具集，简化日常开发操作

[![GitHub](https://img.shields.io/badge/GitHub-siti--cli-blue?logo=github)](https://github.com/SeSiTing/siti-cli)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## ✨ 核心功能

- 🔄 **AI 配置管理** - 快速切换 AI 服务商（MiniMax、智谱、OpenRouter 等）
- 🌐 **代理管理** - 一键开关终端代理
- 🔌 **端口管理** - 快速释放占用的端口
- 🛠️ **实用工具** - 网络检测、IP 显示、日志清理等

## 📦 一键安装

```bash
curl -fsSL https://raw.githubusercontent.com/SeSiTing/siti-cli/main/install.sh | bash
```

安装后运行 `source ~/.zshrc` 使配置生效。

### 其他安装方式

```bash
# Homebrew
brew tap SeSiTing/siti-cli
brew install siti-cli

# 手动安装
git clone https://github.com/SeSiTing/siti-cli.git ~/.siti-cli
echo 'export PATH="$HOME/.siti-cli/bin:$PATH"' >> ~/.zshrc
```

**详细说明：** 查看 [安装指南](docs/INSTALL.md)

## 🎯 快速开始

```bash
# AI 配置管理
siti ai list              # 列出所有 AI 服务商
siti ai switch minimax    # 切换到 MiniMax
siti ai current           # 查看当前配置

# 代理管理
siti proxy on             # 开启代理
siti proxy off            # 关闭代理

# 端口管理
siti killports 3000       # 释放 3000 端口
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

- 报告问题：[GitHub Issues](https://github.com/SeSiTing/siti-cli/issues)
- 贡献代码：[Pull Requests](https://github.com/SeSiTing/siti-cli/pulls)

## 📄 许可证

MIT License - 查看 [LICENSE](LICENSE) 文件了解详情

---

**Star ⭐ 如果你觉得有用！**
