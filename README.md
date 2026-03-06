# siti-cli

个人命令行工具集：AI 配置切换、代理、端口管理、网络检测等。支持 macOS / Linux。

[![GitHub](https://img.shields.io/badge/GitHub-siti--cli-blue?logo=github)](https://github.com/SeSiTing/homebrew-siti-cli)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 安装

**推荐：一键安装**（交互式配置 wrapper，一步到位）

```bash
curl -fsSL https://raw.githubusercontent.com/SeSiTing/homebrew-siti-cli/main/install.sh | bash
```

安装后执行 `source ~/.zshrc` 或重新打开终端。非交互式：加 `--unattended`。

**或：Homebrew**

```bash
brew tap SeSiTing/siti-cli
brew install siti-cli
```

安装后若 `siti ai switch` 不生效，需手动配置 wrapper：`eval "$(siti init zsh)" >> ~/.zshrc`，再 `source ~/.zshrc`。

| 安装方式   | 安装命令                              | 更新命令                          | 卸载命令                                    |
|------------|---------------------------------------|-----------------------------------|---------------------------------------------|
| 一键安装   | `curl -fsSL .../install.sh \| bash`  | `siti upgrade`                    | `siti uninstall -y`                         |
| Homebrew   | `brew tap SeSiTing/siti-cli && brew install siti-cli` | `brew upgrade siti-cli` 或 `siti upgrade` | `brew uninstall siti-cli` 并手动清理残留* |
| 手动安装   | 克隆仓库后 `make install`             | `git pull && make install`        | `make uninstall`                            |

*Homebrew 卸载后需手动清理残留（如 ~/.zshrc 中的 shell wrapper）：
```bash
# 1. 卸载 Homebrew 包
brew uninstall siti-cli

# 2. 清理 ~/.zshrc 中的 siti 相关配置（可选，保留用户数据）
#   - 搜索并删除 "# siti shell wrapper" 段落
#   - 搜索并删除 "# siti-cli PATH configuration" 段落
#   - 搜索并删除 "export PATH=.../.local/bin" 行

# 3. 保留用户数据在 ~/.siti-cli（如需彻底删除）
rm -rf ~/.siti-cli
```

## 快速开始

```bash
siti --help
siti ai list              # 列出 AI 服务商
siti ai switch <provider> # 切换（当前终端生效）
siti proxy on/off         # 代理开关
siti killports 3000       # 释放端口
siti upgrade              # 升级
```

### AI 配置（~/.zshrc）

```bash
# 格式：<PROVIDER>_BASE_URL，可选 <PROVIDER>_API_KEY
export MINIMAX_BASE_URL="https://api.minimaxi.com/anthropic"
export ZHIPU_BASE_URL="https://open.bigmodel.cn/api/anthropic"
export DEFAULT_AUTH_TOKEN="your-token"   # 无 API_KEY 时兜底

# 跳过某服务商（其他程序仍可用原变量名）
export SITI_AI_SKIP="OPENAI,BAILIAN"
```

详见 [快速开始](docs/QUICK_START.md)、[安装指南](docs/INSTALL.md)。

## 文档

- [快速开始](docs/QUICK_START.md) - 上手与自定义命令
- [安装指南](docs/INSTALL.md) - 安装方式对比与手动安装
- [更新日志](CHANGELOG.md)

## 贡献

[Issues](https://github.com/SeSiTing/homebrew-siti-cli/issues) · [Pull Requests](https://github.com/SeSiTing/homebrew-siti-cli/pulls)

MIT License · [LICENSE](LICENSE)
