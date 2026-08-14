# Changelog

按 ISO 日期倒序排列。版本号在标题尾部以 `vX.Y.Z` 形式标注。
本项目遵循 [Semantic Versioning](https://semver.org/lang/zh-CN/spec/v2.0.0.html)。

---

## 2026-08-14 — v2.0.42

- `siti net reset` 不再依赖 active 状态，每次都强制恢复当前 Wi-Fi service 的 DHCP 和自动 DNS，并清除残留或损坏的 active 文件
- `blacklake-proxy` 对齐原始脚本：保留 DHCP 获取的本机 IP 和子网掩码，只替换网关与 DNS；应用前仍校验目标网关位于当前子网
- network YAML profile 新增 `ipv4.subnet_mask: current` 动态模式

## 2026-08-14 — v2.0.41

- network YAML profile 支持 `ssid` 和 `ipv4.address: current`，可以复用当前 DHCP 地址而不在配置文件中写死本机 IP
- 推荐将 Blacklake 的掩码、网关和 DNS 保存在 `~/.siti/network/blacklake-proxy.yaml`；内置 preset 继续作为无文件时的兼容默认值

## 2026-08-14 — v2.0.40

- 修复 `siti net apply` 在 macOS 尚未完成默认路由重建时过早判定失败的问题
- 配置读回成功后最多等待约 8 秒，确认目标 gateway/interface 真正进入内核路由表后再继续 DNS 和外网校验

## 2026-08-14 — v2.0.39

- `siti net apply` 等待 IPv4、网关和 DNS 连续读回一致后，再校验实际默认路由、指定 DNS 解析及绕过本机代理的 GitHub 连接
- 任一可用性检查失败时自动恢复 DHCP 和自动 DNS，不再把“配置已写入”误报为“网络可用”
- apply 成功信息改为展示系统实际读回值及 gateway、DNS、internet 校验结果

## 2026-08-14 — v2.0.38

- `siti net apply blacklake-proxy` 回归只管理 IPv4、网关和 DNS，不再尝试切换受保护的 Wi-Fi
- 已在 `blacklake` 目标网段时直接保留当前地址；不在目标网段时于提权和修改前提示手动连接

## 2026-08-13 — v2.0.37

- 明确 `siti proxy on/off` 用于 Clash Verge 与软路由的日常切换，终端代理同时覆盖 Git、Homebrew 和 curl
- 将 `siti proxy git on/off` 标记为仅影响 Git 的持久高级配置，开启时提示不会代理 Homebrew/curl
- Homebrew 更新前若检测到只有 Git 全局代理，分别提示 Clash Verge 和软路由的正确处理方式

## 2026-08-12 — v2.0.36

- 修复受保护 Wi-Fi 自动切换缺少管理员权限的问题，由 `sudo networksetup` 使用 macOS 已保存的网络凭据
- 识别 `networksetup` 以退出码 0 输出 `Failed to join network` 的异常行为，切换失败时立即返回真实错误，不再等待地址超时

## 2026-08-12 — v2.0.35

- `siti net apply blacklake-proxy` 自动切换到系统已保存的 `blacklake` Wi-Fi，并等待 DHCP 地址就绪
- 保留动态获取的本机 IPv4，按目标 `/21` 网段校验；允许 Blacklake 的 DHCP 路由器为 `172.16.40.1`，不再误要求它等于代理网关 `172.16.40.2`
- network apply/status 输出当前 preset 对应的 Wi-Fi SSID，切换失败时提示先在系统中连接并保存该网络

## 2026-08-12 — v2.0.34

- 内置 `blacklake-proxy` network preset，不再要求 `~/.siti/network/blacklake-proxy.yaml`
- 应用内置 preset 时动态读取当前 Wi-Fi IPv4，固定使用 `255.255.248.0` 掩码及 `172.16.40.2` 网关/DNS，避免在源码中写死本机地址
- 当前 Wi-Fi 网关不是 `172.16.40.2` 时在提权和修改前拒绝，避免在其他网络误应用后断网

## 2026-08-11 — v2.0.33

- 优化 tunnel 输出为“打开 / 转发 / 状态”结构，中文状态明确区分 SSH 转发就绪与应用健康
- 交互终端中的 Dashboard URL 使用 OSC 8 超链接，可直接点击；重定向和 CI 保持纯文本输出

## 2026-08-11 — v2.0.32

- 将内置 tunnel 名称简化为 `studio`，保持底层 SSH target 为 `mac-studio`，直接运行 `siti tunnel up studio`
- 同名用户覆盖路径调整为 `~/.siti/tunnels/studio.yaml`

## 2026-08-11 — v2.0.31

- 内置与 SSH Host alias 同名的 `mac-studio` tunnel preset，升级后可直接运行 `siti tunnel up mac-studio`
- 保留 `~/.siti/tunnels/mac-studio.yaml` 同名用户覆盖，源码不写入 IP、用户名或密钥路径

## 2026-08-11 — v2.0.30

- `siti upgrade` 与 `siti brew up` 在 `brew update` 遇到 DNS、连接超时、TLS syscall、HTTP 429/5xx 等明确瞬时网络错误时自动重试一次
- 保持有界且分类的重试策略，证书校验、权限和 tap 信任等永久性错误立即失败

## 2026-08-11 — v2.0.29

- 新增 `siti tunnel up/down/status/list`，通过 `~/.siti/tunnels/*.yaml` 管理后台 SSH 本地端口转发
- 使用 profile 专属 OpenSSH ControlSocket 实现幂等启停，启动前检查本地端口冲突，状态输出同时展示转发与 TCP 可达性
- 本地监听强制绑定 `127.0.0.1`，复用现有 SSH 配置和认证，不在 profile 中保存密码或私钥

## 2026-08-11 — v2.0.28

- 优化 `siti net reset` 输出，分行展示实际读回的 network service、DHCP 地址、网关和自动 DNS 状态

## 2026-08-10 — v2.0.27

- 用独立、对称的 `siti proxy git on/off` 替换 `proxy off --all`，将当前终端代理与持久 Git 全局代理明确分离
- `proxy status` 和 `siti upgrade` 按代理来源分别提示 `proxy off` 或 `proxy git off`

## 2026-08-10 — v2.0.26

- 新增 `siti proxy off --all`，在关闭当前终端代理的同时显式清理 Git 全局及 URL 级 HTTP/HTTPS 代理
- `siti proxy status` 分层展示终端环境、Git 全局配置和只读的 macOS 系统代理状态
- `siti upgrade` 在 Homebrew/Git 更新前检测失效的本地代理端口，失败时给出清理命令且不自动修改配置

## 2026-08-06 — v2.0.25

- 新增 `siti net apply <profile>`、`reset`、`status`、`list`，通过 `~/.siti/network/*.yaml` 管理 macOS 固定 IPv4、网关和 DNS
- 自动解析 Wi-Fi device 对应的真实 network service，支持 service 重命名；应用和恢复后读回验证，失败时尝试回滚
- 原 `siti net` ping 检测迁移为 `siti net check`

## 2026-08-06 — v2.0.24

- `siti upgrade` 新增 `--bl-ops` 目标，并在 `--all` 中纳入 `bl-ops`；优先识别 uv editable 安装，从本地源码刷新

## 2026-08-04 — v2.0.23

- 内置 DeepSeek Anthropic provider，统一使用 `DEEPSEEK_API_KEY`、`DEEPSEEK_MODEL` 与官方 Anthropic Base URL
- 自定义 provider 不再静默回退到跨服务商的 `DEFAULT_AUTH_TOKEN`，缺少专属 `<NAME>_API_KEY` 时明确诊断
- DeepSeek 明确限定为 Claude/Anthropic 客户端，不提供 Codex Responses 映射
- provider 地址按 Anthropic、Chat Completions、Responses 协议拆分，Grok 与 Codex 分别使用 `<NAME>_CHAT_COMPLETIONS_BASE_URL`、`<NAME>_RESPONSES_BASE_URL`
- 约定 `<NAME>_API_ORIGIN` 作为无路径根地址，仅用于组合协议 Base URL，不误注册为 provider
- `TI_BASE_URL` 作为 ti 客户端目标变量跳过，不误注册为 `ti` provider
- 内置 MiniMax、智谱与 OpenRouter 协议映射，并兼容旧 `<NAME>_BASE_URL`、`<NAME>_OPENAI_BASE_URL`、`<NAME>_GROK_BASE_URL`、`<NAME>_CODEX_BASE_URL` 命名

## 2026-08-03 — v2.0.22

- 修复新版 Homebrew 要求第三方 tap 信任时，`siti upgrade` 忽略错误并误报“已是最新版本”的问题
- 自升级统一使用完整 Formula 名称 `SeSiTing/tap/siti-cli`，并在 update、版本检查或 upgrade 失败时明确报错

## 2026-08-03 — v2.0.21

- `siti ai switch` 默认切换当前 shell 的 Claude + Grok，新增 `--client claude|grok|codex|all` 明确客户端作用域
- 内置 Ali Coding Plan / Bailian 地址和正式模型 `qwen3.8-max`；缺少 Key 时在修改前给出操作提示
- Grok 首次切换自动安装无密钥模型入口，shell wrapper 按当前 Shell 注入模型且尊重显式 `--model`
- 新增 Bailian Codex Responses 全局切换、配置备份/恢复和系统凭证库认证；Ali Coding Plan 因不支持 Responses API 会明确拒绝
- 新增 `siti ai doctor`、多客户端 `current` / `clear` 和 `siti ai credential` 凭证管理命令
- 修复根命令静默吞掉 Cobra 错误的问题，缺 Key、协议不支持等诊断现在会正常显示

## 2026-05-07 — v2.0.20

- fix: `siti upgrade` 检查 brew 版本前先 `brew update`，避免本地 tap 过期时误报"已是最新版本"

## 2026-05-07 — v2.0.19

- shell wrapper 提示改用 `siti init zsh --auto` 替代 `echo >> .zshrc`，自带防重复检测避免重复追加

## 2026-04-29 — v2.0.18

- `siti upgrade` 完全移除 gem 支持：macOS 系统 Ruby gem 无更新价值且因权限报错，删除 `upgrade_gem.go` 和 `--gem` 标志

## 2026-04-29 — v2.0.17

- `siti upgrade` 默认流程移除 gem：系统 Ruby 的 gem 无需更新且会因权限报错，保留 `--gem` 标志供手动使用

## 2026-04-29 — v2.0.16

- `siti upgrade` brew section 过滤逻辑完善：使用 `filteredBefore` 进行汇总和 diffScan，确保 siti-cli 不出现在 brew 的统计中
- brew upgrade 改用 `brew upgrade <names>` 逐个升级公式（排除 siti-cli），不再裸跑无参数

## 2026-04-29 — v2.0.15

- `siti upgrade` self 永远优先：默认先升 self → brew/npm/gem，self 更新后终止并提示 re-run
- brew section 跳过 siti-cli（避免 self 和 brew 重复），`--self` 独立控制自升级
- `sectionSelf` 改为返回 bool 表示是否有更新，dry-run 与真实执行显示一致过滤

## 2026-04-29 — v2.0.14

- `siti upgrade` 修复 brew/self 实时日志：裸 `exec.Command().Run()` 吞掉 TTY 输出导致用户看到"卡住"假象，统一改为 `runCmd()` 继承 stdout/stderr

## 2026-04-29 — v2.0.13

- `siti upgrade` 重构为多包管理器升级命令：默认升级 brew + npm global + gem，新增 `--self` / `--brew` / `--npm` / `--gem` / `--all` / `--dry-run` 标志
- `upgrade.go` 拆分为 5 个文件（upgrade.go + upgrade_self.go + upgrade_brew.go + upgrade_npm.go + upgrade_gem.go），每文件 <110 行
- `runCmd` 系列工具函数从 upgrade.go 迁移至 util.go，避免 brew.go 编译断裂

## 2026-04-29 — v2.0.12

- `siti brew up` 升级后新增二次扫描验证步骤，通过前后 outdated 差值精确汇总当次实际更新的 formula / cask 列表（含版本号和包名）
- 汇总区分"无需更新"与"无 package 被成功更新（跳过/失败）"，消除歧义

## 2026-04-29 — v2.0.11

- CLAUDE.md 强化版本号升级规则：每次提交必须升 z，AI 只允许升 z，禁止 minor/major

## 2026-04-29 — v2.0.10

- `siti brew up` 修复扫描始终为空导致跳过升级的问题：`brew outdated` 默认只输出包名，改为 `--verbose` 获取版本信息，同时解析 `!=` 操作符
- `runCmd` 打印原始命令（`$ brew upgrade`），方便追踪执行过程

## 2026-04-28 — v2.0.9 · `siti brew up` 预览与汇总

- `siti brew up` 升级前自动扫描并展示待更新的 formula / cask 清单，消除盲盒感
- 新增 `-n` / `--dry-run` 标志：仅预览不执行
- 末尾汇总显示更新数量、清理空间、总耗时
- 全部最新时自动跳过 upgrade/autoremove 步骤，节省时间
- `Makefile` 修复版本号 grep 误匹配注释

## 2026-04-28 — v2.0.8 · shell wrapper 集成优化

- `siti ai switch` / `siti proxy on` 等需要 eval 的命令在未加载 shell wrapper 时，stderr 给出清晰提示 + 一键可复制的修复命令，避免 export 语句打印到终端但不生效的静默失败
- `siti init zsh --auto` 自动检测配置文件（`.zshenv` / `.zshrc`），幂等追加 wrapper，无需手动编辑
- wrapper 模板新增 `SITI_WRAPPER=1` / `set -gx SITI_WRAPPER 1` 环境变量标记，供 Go 端检测 shell 集成状态

## 2026-04-28 — v2.0.7 · shell wrapper 直通 TTY

- 重写 shell wrapper：子进程直接继承终端（不再重定向 stdout），修复 `siti brew up` 无进度条的问题
- exit-10 协议改用 `$SITI_EVAL_FILE` 临时文件传递 shell 语句，stdout 不再被 wrapper 捕获

## 2026-04-28 — v2.0.6 · wrapper 不再捕获 stdout

- 重写 shell wrapper：stdout 重定向到临时文件而非 `$()` 捕获，子进程直接继承 TTY，修复 `siti brew up` 无进度条的问题
- Formula 迁移到 `SeSiTing/homebrew-tap` 统一 tap 仓库，安装命令简化为 `brew install SeSiTing/tap/siti-cli`

## 2026-04-27 — v2.0.4 · brew up 输出修复

- 修复 `siti brew up` 子命令无实时输出的问题：`runCmdIn` 补上 `c.Stdin = os.Stdin`，使 brew 子进程正确检测 TTY，恢复进度条和实时日志

## 2026-04-27 — v2.0.3 · 输出风格主流化（去 emoji）

对齐 gh / cargo / kubectl / pnpm / bun 的纯文本+符号风格：

- 全项目去 emoji。统一符号集：`✓ ✗ ! ? → ↷ •`
- 进程标签从 `🐳 Docker` 改为 `[docker]`，`🟢 Node.js` → `[node]` 等小写文字标签
- 区块分隔从 `━` 改为 `─`
- 时间从 `⏱ 总耗时` 改为 `took 1.2s`
- 区块标题从 `🌐 内网 IP:` 改为 `LAN:` / `WAN:`
- CLAUDE.md 新增「输出风格规范」章节，明确符号集和文案规则

理由：emoji 在不同终端宽度不一致、SSH/CI 日志会出锅、屏幕阅读器读不出。

## 2026-04-27 — v2.0.2 · CLI 命名规范化

> 注：本次包含命令重命名（严格 semver 应升 major），但按项目约定 AI 助手只允许自动升 patch 位，故记为 v2.0.2。

**命令重命名**（无兼容 alias，肌肉记忆需要重建）：

| 旧 | 新 |
|---|---|
| `siti ai unset` | `siti ai clear` |
| `siti proxy check` | `siti proxy status` |
| `siti ipshow` | `siti ip` |
| `siti netcheck` | `siti net` |
| `siti killports 3000` | `siti port kill 3000` |
| `siti cleanlogs` | `siti logs clean` |
| `siti brewup` | `siti brew up` |

理由：对齐 gh / kubectl / docker / brew 主流命名（namespace + verb），淘汰 bash 时代的复合词。

**Added**

- `siti version` 子命令（与 `--version` flag 并存，对齐 gh/kubectl/docker/git）
- `siti ip` 公网 IP 查询：尝试 ipify → ifconfig.me → ipinfo 三个 endpoint，修复旧版 64 字节截断把 HTML 错误页输出成乱码的 bug
- semver 升级门禁：CI 在版本号变化时校验，patch 自动放行；minor/major 必须 commit message 含 `[minor-bump]` / `[major-bump]`
- `CLAUDE.md` 明确 AI 助手默认只升 patch 位，minor/major 必须用户授权

**Fixed**

- `publish-on-version-bump.yml` 提取版本号的 grep 误匹配注释里的示例版本，导致 `Invalid format` 错误

---

## 2026-04-27 — v2.0.1 · Go + Cobra 全面重构

> 注：v2.0.0 因 CI 配置 bug 未实际发布，v2.0.1 为首个生效版本。包含 v2.0.0 全部内容 + workflow 修复（grep 误匹配注释示例 / 新增 semver 升级门禁）。



**Breaking changes**

- 整个项目从 bash 脚本集合迁移到 Go 单二进制
- 删除 `--persist` 选项：`siti ai switch` 仅修改当前 shell；要永久切换默认值，请手动编辑 `~/.zshrc`
- 删除独立安装脚本 `install.sh` 和 `~/.siti-cli/commands/` 自定义命令机制
- 唯一安装方式：`brew install SeSiTing/tap/siti-cli`

**Added**

- Go + Cobra 命令框架，自动生成 zsh / bash 补全
- `charmbracelet/huh` 交互式选择器（`siti ai switch` 无参数时启用）
- `goreleaser` 一键交叉编译 + Homebrew Formula 自动更新
- 单元测试 + golden 文件 snapshot（`internal/shell`、`internal/config`）
- `Makefile` 标准化开发命令（`make build/test/tidy/snapshot`）
- `AGENTS.md` 指向 `CLAUDE.md` 作为单一真源
- CI workflow（`.github/workflows/ci.yml`）：PR 必须 build/vet/test 通过

**Changed**

- 部分命令的 eval 协议：通过 cobra context 内的 `EvalBuffer` 收集 shell 语句，main 检测后 stdout 输出 + exit 10
- shell wrapper 不再 grep 白名单过滤——信任 Go 端 stdout 契约
- AI 服务商从 `~/.zshenv` 和 `~/.zshrc` 双文件发现（zshenv 优先）
- `ai switch` 自动管理 5 个 ANTHROPIC 模型变量

**Removed**

- `bin/siti`、`src/commands/*.sh`、`scripts/*.sh`、`docs/*.md`、`install.sh`
- 旧的 `EvalDirective` error-interface 模式

---

## 2026-03-22 — v1.2.6 / v1.2.7

- `siti ai unset`: 修复 "local: can only be used in a function" 错误
- `chore`: Formula 升级到 v1.2.27

## 2026-03-06 — v1.2.5

- `siti ai switch` 智能管理 ANTHROPIC 模型变量：
  - 检测 `<PROVIDER>_MODEL`，存在则同步设置 5 个 `ANTHROPIC_*_MODEL`，否则全部 unset
  - 支持临时切换和持久化（`--persist`）
- `siti ai unset` 同步清理这 5 个变量

## 2026-03-01 — v1.2.4

- 新增 `siti ai unset`，用于切换到 OAuth 登录模式
  - 临时清除 / `--persist` 持久化清除 ANTHROPIC_* 变量
  - shell wrapper 未配置时友好提示

## 2026-02-02 — v1.2.3

- `siti proxy`: 命令参数忽略大小写，环境变量同时设置/清理大小写两个版本
- `siti proxy check`: 显示 `no_proxy` / `NO_PROXY`
- `siti ai list`: 修复注释行被误识别为「当前」的问题

## 2026-02-01 — v1.0.0 → v1.2.2（同日多次发布）

**v1.2.2** — 重构 AI 跳过机制：从 `SKIP_` 前缀改为 `SITI_AI_SKIP` 变量

**v1.2.1** — 改进 `siti uninstall` 交互体验（Rust/Go 风格 `-y` / `--dry-run`）

**v1.2.0** — 目录统一为 `~/.siti-cli`，新增独立安装的 `siti uninstall`

**v1.1.0** — 修复 `siti ai switch` 误报 wrapper 未配置；Homebrew `post_install` 健壮性

**v1.0.9** — 修复重复追加 PATH、wrapper 检测改用 `declare -f`

**v1.0.8** — 修复独立安装克隆错误的仓库地址

**v1.0.7** — 支持包含数字的服务商名（LLMS8、LLMS9）

**v1.0.6** — 新增 `siti upgrade` 和 `siti init <shell>`，独立安装支持 `--unattended`

**v1.0.5** — 自动安装 shell wrapper，`siti ai switch` 和 `siti proxy` 开箱即用

**v1.0.4** — 修复 zsh / bash 补全在 Homebrew 安装时的路径检测

## 2026-01-31 — v1.0.2 / v1.0.3

- 改进 GitHub Actions 发布流程，自动更新 Formula

## 2024-初版 — v1.0.0 / v1.0.1

- 支持 Homebrew 安装、用户自定义命令、shell 补全、配置/日志/缓存目录
- 重构 `bin/siti` 支持多种安装路径
