# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
