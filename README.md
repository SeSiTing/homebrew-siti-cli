# siti-cli

个人命令行工具集：AI 服务商配置切换、代理管理、网络配置、端口管理等。
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
siti upgrade --bl-ops   # 刷新 bl-ops（优先识别 uv editable 安装）
siti upgrade --all      # self + brew + npm + bl-ops
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
siti proxy git on / off       # 持久开启或关闭 Git 全局代理
siti proxy status             # 查看终端、Git 和 macOS 系统代理

# 端口 / 网络 / 日志
siti port kill 3000 8080    # 释放端口（支持 --dev/--db/--web 预设）
siti net apply blacklake-proxy  # 应用固定网络配置（macOS）
siti net reset              # 恢复 DHCP 和自动 DNS
siti net status             # 查看当前 siti 管理的网络配置
siti net list               # 列出 ~/.siti/network/*.yaml
siti net check              # ping baidu/google/github
siti tunnel up studio       # 后台启动 Studio 本地转发
siti tunnel status studio   # 查看 tunnel 和本地端口状态
siti tunnel down studio     # 关闭 siti 管理的 SSH tunnel
siti tunnel list            # 列出 ~/.siti/tunnels/*.yaml
siti ip                     # 内网 + 公网 IP
siti logs clean             # 清理当前目录 *.log

# Homebrew & 自身
siti brew up                # brew update + upgrade + cleanup 一键
siti upgrade                # 默认升级 self + brew + npm
siti upgrade --self         # 仅升级 siti-cli 自身
siti upgrade --bl-ops       # 升级或刷新 bl-ops
siti upgrade --all          # 升级 self + brew + npm + bl-ops
siti init zsh|bash|fish     # 输出 shell wrapper
```

`siti upgrade` 会在运行 Homebrew 或 Git 更新前检查本地代理端口。若 Git 或终端代理指向未监听的 `localhost` / `127.0.0.1`，升级会在修改任何配置前停止，并根据来源提示运行 `siti proxy off` 或 `siti proxy git off`；不会自动删除代理配置。

`siti upgrade` 和 `siti brew up` 遇到 DNS、连接超时、TLS syscall、HTTP 429/5xx 等明确的临时网络错误时，会等待 1 秒后自动重试一次 `brew update`。证书校验、权限、tap 信任和其他永久性错误不会重试。

## 网络配置

`siti net` 在 macOS 上通过 `networksetup` 管理固定 IPv4。profile 存放在 `~/.siti/network/`，例如：

```yaml
# ~/.siti/network/blacklake-proxy.yaml
version: 1
interface: wifi

ipv4:
  address: 172.16.40.100
  subnet_mask: 255.255.255.0
  gateway: 172.16.40.2

dns:
  - 172.16.40.2
```

应用和恢复网络配置需要管理员权限，命令会通过 `sudo` 请求授权：

```bash
siti net apply blacklake-proxy
siti net status
siti net reset
```

程序会自动查找 Wi-Fi 对应的 device 和 network service，不依赖 service 名称必须为 `Wi-Fi`。`reset` 只处理 siti 记录的 active profile，恢复 DHCP 和自动 DNS；不会恢复应用前的手动网络参数。

## SSH tunnel

`siti tunnel` 使用系统 OpenSSH 和已有的 `~/.ssh/config`，在后台管理仅绑定本机回环地址的 SSH 本地端口转发。只要 `ssh mac-studio` 可以直接连接，就可以零配置运行：

```bash
siti tunnel up studio
siti tunnel status studio
siti tunnel down studio
```

内置 `studio` preset 的 SSH target 是 `mac-studio`，包含 OpenClaw `19010→9010` 和 Hermes `19119→9119`。需要覆盖目标或端口时，在 `~/.siti/tunnels/` 创建同名 profile，用户配置优先：

```yaml
# ~/.siti/tunnels/studio.yaml
version: 1
target: mac-studio

forwards:
  - name: openclaw
    local_port: 19010
    remote_port: 9010
    url: http://127.0.0.1:19010/

  - name: hermes
    local_port: 19119
    remote_port: 9119
    url: http://127.0.0.1:19119/
```

`remote_host` 默认为远端 `127.0.0.1`。`target` 使用平时可以直接传给 `ssh` 的别名或 `user@host`；端口、跳板机、密钥等连接细节继续放在 OpenSSH 配置中，不写入 tunnel profile。

`up` 使用 profile 专属 ControlSocket 后台运行，不占用当前终端；重复执行是幂等的。启动前会检查本地端口冲突，固定使用 `127.0.0.1` 本地绑定，并启用 SSH keepalive 与 `ExitOnForwardFailure`。`down` 只关闭对应 profile 的 SSH master，不影响其他 SSH 会话。

Tunnel 只负责转发，不会启动或重启远端服务。`status` 中 `reachable` 表示本地转发端口可建立 TCP 连接，不代表应用认证或业务请求一定成功。Profile URL 仅用于展示，必须指向本机 loopback，且不允许携带凭证或 query 参数。

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
