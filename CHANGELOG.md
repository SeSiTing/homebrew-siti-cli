# Changelog

按 ISO 日期倒序排列。版本号在标题尾部以 `vX.Y.Z` 形式标注。
本项目遵循 [Semantic Versioning](https://semver.org/lang/zh-CN/spec/v2.0.0.html)。

---

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
