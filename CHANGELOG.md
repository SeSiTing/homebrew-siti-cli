# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
