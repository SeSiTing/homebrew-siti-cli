# Changelog

按 ISO 日期倒序排列。版本号在标题尾部以 `vX.Y.Z` 形式标注。
本项目遵循 [Semantic Versioning](https://semver.org/lang/zh-CN/spec/v2.0.0.html)。

---

## 2026-04-27 — v2.0.1 · Go + Cobra 全面重构

> 注：v2.0.0 因 CI 配置 bug 未实际发布，v2.0.1 为首个生效版本。包含 v2.0.0 全部内容 + workflow 修复（grep 误匹配注释示例 / 新增 semver 升级门禁）。



**Breaking changes**

- 整个项目从 bash 脚本集合迁移到 Go 单二进制
- 删除 `--persist` 选项：`siti ai switch` 仅修改当前 shell；要永久切换默认值，请手动编辑 `~/.zshrc`
- 删除独立安装脚本 `install.sh` 和 `~/.siti-cli/commands/` 自定义命令机制
- 唯一安装方式：`brew install SeSiTing/siti-cli/siti-cli`

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
