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
siti init zsh --auto
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
siti ai current             # 显示 Claude / Grok / Codex 当前配置
siti ai switch [name]       # 默认切换当前 shell 的 Claude + Grok
siti ai switch bailian --client codex  # 全局切换 Codex
siti ai doctor              # 检查环境变量、系统凭证和配置
siti ai test                # 测试当前 Claude API 连通性
siti ai clear               # 清除当前 shell 的 Claude + Grok 变量
siti ai clear --client codex            # 恢复 Codex 原有全局配置

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

Ali Coding Plan、百炼 Bailian 和 DeepSeek Anthropic 的非敏感地址与默认模型已内置，通常只需配置服务商级 API Key：

```bash
export ALI_API_KEY="sk-sp-..."     # Coding Plan
export BAILIAN_API_KEY="sk-..."    # 百炼按量付费 / Token Plan
export DEEPSEEK_API_KEY="sk-..."    # DeepSeek Anthropic

# 可选覆盖；不设置时使用 qwen3.8-max
export ALI_MODEL="qwen3.8-max"
export BAILIAN_MODEL="qwen3.8-max"
export DEEPSEEK_MODEL="deepseek-v4-pro"

# 跳过列表（逗号分隔，大写名称）
export SITI_AI_SKIP="OPENAI,OPENROUTER"
```

建议保护包含密钥的配置文件：

```bash
chmod 600 ~/.zshenv
```

默认命令只修改当前 shell，不碰 Codex：

```bash
siti ai switch ali                  # Claude + Grok
siti ai switch deepseek --client claude
siti ai switch ali --client claude
siti ai switch ali --client grok
```

Grok 的无密钥模型入口会在首次切换时自动写入 `~/.grok/config.toml`。shell wrapper 会让当前 Shell 中启动的 `grok` 自动使用 `siti-ali` 等对应入口；显式传入 `grok --model ...` 时以用户参数为准。协议地址使用 `ALI_CHAT_COMPLETIONS_BASE_URL`；只有 Key 或模型确实不同时，才使用 `ALI_GROK_API_KEY`、`ALI_GROK_MODEL` 等客户端专属覆盖。

Codex 必须显式全局切换。先把 Key 从环境变量导入 macOS Keychain / Linux Secret Service，再切换：

```bash
siti ai credential import bailian --from-env BAILIAN_API_KEY
siti ai switch bailian --client codex
```

该命令只管理 `~/.codex/config.toml` 中带 `siti-cli` 标记的区块，写入前备份，不修改 `auth.json`。它会影响 Codex CLI、Desktop 和 IDE，已运行的 Codex 需要重启。Ali Coding Plan 不支持 Codex 所需的 Responses API，因此 `siti ai switch ali --client codex` 会在修改前明确拒绝；Bailian Responses-compatible 端点可以直接使用。若已有 Responses 转换路由，可显式设置 `ALI_RESPONSES_BASE_URL` 启用 Ali 的 Codex 映射。

自定义服务商可用 `<NAME>_API_ORIGIN` 保存无路径的 `scheme + host + 可选端口`，再派生协议地址；`API_ORIGIN` 只用于组合，不会单独注册 provider。Claude 使用 `<NAME>_ANTHROPIC_BASE_URL`，Grok 使用 `<NAME>_CHAT_COMPLETIONS_BASE_URL`，Codex 只使用明确支持 Responses 的 `<NAME>_RESPONSES_BASE_URL`。Key 与模型统一使用 `<NAME>_API_KEY`、`<NAME>_MODEL`，不跨 provider 回退。旧的 `<NAME>_BASE_URL`、`<NAME>_OPENAI_BASE_URL`、`<NAME>_GROK_BASE_URL`、`<NAME>_CODEX_BASE_URL` 继续兼容读取。

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
