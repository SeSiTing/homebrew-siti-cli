# siti-cli

个人命令行工具集：AI 服务商配置切换、代理管理、端口管理、网络检测等。
macOS / Linux 通用，Go 实现。

[![CI](https://github.com/SeSiTing/siti-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/SeSiTing/siti-cli/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 安装

通过 Homebrew tap 安装：

```bash
brew install SeSiTing/tap/siti-cli
```

或分两步：

```bash
brew tap SeSiTing/tap
brew install siti-cli
```

安装后**必须**配置 shell wrapper（一次性，让 `ai switch` / `proxy on` 能改父 shell 环境）：

```bash
echo 'eval "$(siti init zsh)"' >> ~/.zshrc
source ~/.zshrc
```

bash / fish 把 `zsh` 替换为对应名称即可。

## 升级 / 卸载

```bash
brew upgrade siti-cli   # 或 siti upgrade
brew uninstall siti-cli # 卸载后手动删除 ~/.zshrc 里的 wrapper 行
```

## 命令一览

```bash
siti --help                 # 查看所有命令
siti version                # 显示版本（也可用 -v / --version）

# AI 服务商
siti ai list                # 列出可用服务商
siti ai current             # 当前服务商
siti ai switch [name]       # 切换（无参数交互式选择）
siti ai test                # 测试当前 API 连通性
siti ai clear               # 清除 ANTHROPIC_* 变量

# 终端代理 (127.0.0.1:7890)
siti proxy on / off
siti proxy status

# 端口 / 网络 / 日志
siti port kill 3000 8080    # 释放端口（支持 --dev/--db/--web 预设）
siti net                    # ping baidu/google/github
siti ip                     # 内网 + 公网 IP
siti logs clean             # 清理当前目录 *.log

# Homebrew & 自身
siti brew up                # brew update + upgrade + cleanup 一键
siti upgrade                # 升级 siti-cli 自身
siti init zsh|bash|fish     # 输出 shell wrapper
```

## AI 服务商配置

`siti ai` 从 `~/.zshenv` 和 `~/.zshrc` 自动发现 `*_BASE_URL` 形式的环境变量。
约定：

```bash
# ~/.zshrc 或 ~/.zshenv 中按需添加
export MINIMAX_BASE_URL="https://api.minimaxi.com/anthropic"
export MINIMAX_API_KEY="sk-..."
export MINIMAX_MODEL="abab6.5"          # 可选：切换时同步设置 ANTHROPIC_MODEL

export ZHIPU_BASE_URL="https://open.bigmodel.cn/api/anthropic"
export ZHIPU_API_KEY="..."

# 不设 <PROVIDER>_API_KEY 时回退到 DEFAULT_AUTH_TOKEN
export DEFAULT_AUTH_TOKEN="..."

# 跳过列表（逗号分隔，大写名称）
export SITI_AI_SKIP="OPENAI,BAILIAN"
```

执行 `siti ai switch minimax` 时，wrapper 会把 `ANTHROPIC_BASE_URL` 等指向上面的引用变量，**仅当前 shell 生效**。要永久切换默认值，自行修改 zshrc。

## 工作机制

部分命令需要修改父 shell 环境变量（`ai switch` / `proxy on/off`），通过 **exit 10 协议**实现：

1. Go 命令把 shell 语句写入 stdout，以 exit 10 退出
2. `siti init zsh` 生成的 wrapper 检测到 exit 10，对 stdout `eval` 后返回 0
3. 当前 shell 环境变量被修改

详见 [CLAUDE.md](./CLAUDE.md) 的"架构"小节。

## 开发

```bash
make help          # 列出所有 make 目标
make build         # 构建到 ./siti
make test          # 跑单元测试
make tidy          # go mod tidy + 校验
make snapshot      # 本地 goreleaser dry-run（需要 brew install goreleaser）
```

完整开发指南、目录结构、新增命令规范见 [CLAUDE.md](./CLAUDE.md)。

## 链接

- [更新日志](CHANGELOG.md)
- [Issues](https://github.com/SeSiTing/siti-cli/issues) · [Pull Requests](https://github.com/SeSiTing/siti-cli/pulls)

MIT License — 见 [LICENSE](LICENSE)。
