# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
