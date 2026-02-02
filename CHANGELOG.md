# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.3] - 2026-02-02

### Fixed
- `siti proxy`: 命令参数忽略大小写（on/off/check/status），环境变量同时设置/清理大小写两个版本（http_proxy/HTTP_PROXY 等）
- `siti proxy check`: 显示 no_proxy/NO_PROXY（有则打印，两个都有则都打印）
- `siti ai list`: 修复注释行被误识别为「当前」导致显示多个「← 当前」的问题

## [1.2.2] - 2026-02-01

### Changed
- 重构 AI 跳过机制：从 `SKIP_` 前缀改为 `SITI_AI_SKIP` 变量
  - 使用 `export SITI_AI_SKIP="OPENAI,BAILIAN"` 标记要跳过的服务商
  - 移除 `SKIP_` 前缀要求，避免影响其他程序使用原变量名
  - 支持从环境变量或 `.zshrc` 读取跳过列表

### Technical
- `ai.sh` 的 `list_providers()` 和 `switch_provider()` 改用 `SITI_AI_SKIP` 检查
- 新增 `get_skip_list()` 辅助函数，统一读取跳过列表
- 跳过列表格式：逗号分隔的大写服务商名称（如 "OPENAI,BAILIAN"）

### Migration Guide

如果你的 `.zshrc` 中使用了 `SKIP_` 前缀：

```bash
# 旧方式（v1.2.1 及之前）
export SKIP_OPENAI_BASE_URL="https://api.openai.com/v1"
export SKIP_BAILIAN_BASE_URL="https://dashscope.aliyuncs.com/v1"

# 新方式（v1.2.2+）
export OPENAI_BASE_URL="https://api.openai.com/v1"
export BAILIAN_BASE_URL="https://dashscope.aliyuncs.com/v1"
export SITI_AI_SKIP="OPENAI,BAILIAN"
```

好处：其他程序仍然可以使用 `OPENAI_BASE_URL` 等变量。

## [1.2.1] - 2026-02-01

### Changed
- 改进 `siti uninstall` 交互体验（Rust/Go 风格）
  - 默认显示将删除的内容（文件列表、大小）
  - 需要 `-y` 或 `--yes` 标志确认卸载
  - 新增 `--dry-run` 仅预览，不执行
  - 符合主流 CLI 工具交互模式（rustup、cargo 等）
- 更新帮助文本和文档

### Technical
- `src/commands/uninstall.sh` 参数解析：支持 `-y/--yes`、`--dry-run`、`--help`
- 删除交互式 `read -p` 确认，改为显式标志确认

## [1.2.0] - 2026-02-01

### Added
- ✨ 新增 `siti uninstall` 命令（独立安装）
  - 交互式确认后清理 `.zshrc` 中的 wrapper、补全、PATH 配置
  - 删除符号链接 `~/.local/bin/siti` 和安装目录 `~/.siti-cli`
  - Homebrew 安装时提示使用 `brew uninstall siti-cli`
- 📦 新增迁移脚本 `scripts/migrate-to-unified.sh`
  - 将旧目录 `~/.siti` 安全迁移到 `~/.siti-cli`（备份后迁移）

### Changed
- **目录统一**：用户数据与程序统一使用 `~/.siti-cli`
  - 原 `~/.siti`（commands、config、logs、cache）合并到 `~/.siti-cli`
  - 独立安装：程序与用户数据均在 `~/.siti-cli`
  - Homebrew：用户数据目录为 `~/.siti-cli`（程序在 Cellar）
- 🔄 安装/升级时自动检测并迁移旧目录 `~/.siti` 到 `~/.siti-cli`
  - 迁移前备份到 `~/.siti.backup.TIMESTAMP`
  - `post-install.sh` 与 `install.sh` 均支持迁移
- 📝 `post-uninstall.sh` 提示用户数据保留在 `~/.siti-cli`
- 📝 README 增加卸载说明和目录结构说明

### Technical
- `bin/siti` 中用户命令目录从 `$HOME/.siti/commands` 改为 `$HOME/.siti-cli/commands`
- 配置文件 `siti.conf` 中路径统一为 `~/.siti-cli/*`
## [1.1.0] - 2026-02-01

### Fixed
- 🐛 彻底修复 `siti ai switch` 误报 wrapper 未配置的问题
  - 改用检查 `~/.zshrc` 文件内容，而非当前 shell 函数状态
  - 解决子进程（bash）无法检测到父 shell（zsh）函数的问题
- 🔒 修复 Homebrew `post_install` 因权限问题导致安装失败
  - 移除 `set -e`，允许权限错误时继续执行
  - shell wrapper 和补全配置失败时静默跳过，显示友好提示
  - 不再导致 `brew upgrade` 报错

### Changed
- ✨ 添加 Homebrew `caveats` 提示
  - 安装后清晰告知用户如何手动配置 shell wrapper
  - 说明哪些命令需要 wrapper 才能生效
- 📝 改进 README 安装说明
  - 明确 Homebrew 安装需要手动配置 wrapper
  - 强调自动配置可能因权限失败
- 🛡️ 增强 `post-install.sh` 健壮性
  - 写入失败时提供明确的错误提示
  - 提供手动配置的后备方案

### Technical
- 修改 `ai.sh` 的 wrapper 检测逻辑：从 `declare -f siti` 改为 `grep -q "# siti shell wrapper" ~/.zshrc`
- 修改 `post-install.sh` 所有 `cat >> ~/.zshrc` 操作为条件写入，失败时不中断脚本

## [1.0.9] - 2026-02-01

### Fixed
- 修复独立安装脚本重复追加 PATH 配置到 ~/.zshrc 的问题
- 修复 `siti ai switch` 错误检测 wrapper 未安装的问题
  - 改用 `declare -f` 检测函数是否存在
  - 改进错误提示，引导使用 `siti init` 命令
- 添加自动清理重复配置的逻辑

### Changed
- 改进 PATH 配置检测，使用唯一标记 "# siti-cli PATH configuration - auto-generated"
- 优化 shell wrapper 检测逻辑，更健壮

## [1.0.8] - 2026-02-01

### Fixed
- 修复独立安装脚本克隆错误的仓库地址
  - 从 `https://github.com/SeSiTing/siti-cli.git` 改为 `https://github.com/SeSiTing/homebrew-siti-cli.git`
  - 影响文件：install.sh, README.md
- 修复独立安装用户无法获取最新版本的问题
- 添加自动检测和修复旧仓库地址的逻辑（对已安装用户透明升级）

### Changed
- 更新独立安装脚本 URL 到正确的仓库地址
- 更新文档链接统一使用 `homebrew-siti-cli`

## [1.0.7] - 2026-02-01

### Fixed
- 修复 `siti ai list` 不显示包含数字的服务商名称（如 LLMS8、LLMS9）
- 修复 `siti ai switch` 无法切换到包含数字的服务商
- 修复 `siti ai current` 解析包含数字的服务商名称失败

### Technical
- 更新正则表达式从 `[A-Z_]+` 到 `[A-Z0-9_]+` 以支持数字
- 影响文件：src/commands/ai.sh (3 处修改)

## [1.0.6] - 2026-02-01

### Added
- ✨ 新增 `siti upgrade` 命令，智能检测安装方式并自动更新
  - Homebrew 安装：自动调用 `brew upgrade siti-cli`
  - 独立脚本安装：自动执行 `git pull` 更新
  - 显示当前版本和更新日志
- ✨ 新增 `siti init <shell>` 命令，生成 shell wrapper 配置
  - 支持 zsh、bash、sh
  - 输出可审查的配置代码
  - 便于手动添加到 shell 配置文件
- 📦 增强独立安装脚本 (`install.sh`)
  - 支持 `--unattended` 非交互模式，适合自动化脚本
  - 支持 `--skip-wrapper` 跳过 shell wrapper 安装
  - 添加安装方式标识文件 (`.install-source`)
  - 改进错误处理和用户提示

### Changed
- 🔧 优化安装方式检测逻辑
  - 新增 `INSTALL_METHOD` 环境变量（`homebrew`/`standalone`/`source`）
  - 支持独立安装脚本检测（检查 `~/.siti-cli` 和 `~/.local/bin/siti`）
  - 安装方式信息传递给子命令使用
- 📝 重写 README 文档
  - 明确两种安装方式的差异和适用场景
  - 添加安装方式对比表格
  - 补充 `siti upgrade` 使用说明
  - 改进快速开始示例
- 🎨 改进命令行输出格式
  - 统一颜色和图标使用
  - 更友好的错误提示信息

### Documentation
- 📚 添加双模式安装详细说明（Homebrew vs 独立脚本）
- 📚 添加安装方式对比表格，帮助用户选择
- 📚 完善 `siti upgrade` 和 `siti init` 命令文档

## [1.0.5] - 2026-02-01

### Added
- 自动安装 shell wrapper 功能，`siti ai switch` 和 `siti proxy` 命令开箱即用
- 添加 `post_uninstall` 钩子，卸载时自动清理 shell 配置
- 添加 shell wrapper 检测，未安装时友好提示用户

### Fixed
- 修复 `siti ai switch` 中文括号乱码问题（改用英文方括号）
- 修复 `siti ai switch` 切换后不生效的问题（自动安装 wrapper）

### Changed
- `post-install.sh` 现在会自动安装 shell wrapper 到 `~/.zshrc`
- `brew upgrade` 时自动检查并更新 shell wrapper
- `brew uninstall` 时自动清理 shell wrapper 和补全配置
- 优化用户体验，无需手动配置即可使用所有功能

## [1.0.4] - 2026-02-01

### Fixed
- 修复 zsh 补全脚本在 Homebrew 安装时的路径检测问题
- 修复 bash 补全脚本在 Homebrew 安装时的路径检测问题
- 补全脚本现在能正确识别 Homebrew 安装路径（`/opt/homebrew/share/siti-cli/commands`）

### Changed
- 补全脚本智能检测安装类型（Homebrew vs 源码开发模式）
- 统一 zsh 和 bash 补全的路径检测逻辑，与 `bin/siti` 保持一致

## [1.0.3] - 2026-01-31

### Changed
- 改进 GitHub Actions 发布流程
- 自动更新 Formula 文件

## [1.0.2] - 2026-01-31

### Changed
- 升级版本到 v1.0.2

## [1.0.1] - 2024-01-XX

### Changed
- 更新 Homebrew Formula 配置

## [1.0.0] - 2024-01-XX

### Added
- 支持 Homebrew 安装方式
- 用户自定义命令功能（`~/.siti/commands/`）
- 自动配置 shell 补全
- 配置文件管理（`~/.siti/config/`）
- 日志和缓存目录（`~/.siti/logs/`, `~/.siti/cache/`）
- GitHub Actions 自动化发布流程

### Changed
- 重构 `bin/siti` 支持多种安装路径
- 优化命令查找逻辑，优先使用用户自定义命令
- 更新安装说明，推荐使用 Homebrew

### Removed
- 删除 `setup.sh` 手动安装脚本
- 删除 `uninstall.sh` 卸载脚本
- 移除对项目目录的依赖

### Fixed
- 修复路径解析问题
- 改进错误处理机制
