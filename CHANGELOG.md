# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
