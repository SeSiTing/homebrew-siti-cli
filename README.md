# siti-cli

个人命令行工具集，用于简化日常开发操作。

## 功能特点

- 统一的命令入口：通过 `siti` 命令访问所有工具
- 简洁易用：简单直观的命令结构
- 易于扩展：轻松添加新功能
- 自动补全：支持命令和参数的自动补全

## 安装

```bash
# 克隆仓库
git clone https://github.com/yourusername/siti-cli.git
cd siti-cli

# 运行配置脚本
./setup.sh

# 使配置生效
source ~/.zshrc  # 或 ~/.bashrc
```

## 使用方法

### 代理控制

```bash
# 开启终端代理
siti proxy on

# 关闭终端代理
siti proxy off

# 检查代理状态
siti proxy check
```

### 网络工具

```bash
# 检查网络连接
siti netcheck

# 显示IP地址
siti ipshow
```

### 端口管理

```bash
# 释放默认端口范围 (2024-2030, 8000-8010, 8080-8090, 9000-9010)
siti killports

# 释放指定端口
siti killports 3000 5000

# 释放端口范围
siti killports 3000-3010

# 混合方式指定端口
siti killports 8080 3000-3010 9000

# 仅检查端口占用情况，不释放
siti killports check

# 显示帮助信息
siti killports help
```

### 配置备份

```bash
# 备份zshrc配置
siti backup-zshrc
```

### 日志清理

```bash
# 清理日志文件
siti cleanlogs
```

## 扩展

要添加新命令，只需在 `src/commands` 目录中创建新的 `.sh` 脚本文件。脚本将自动被 `siti` 工具识别。

## 帮助

```bash
# 显示帮助信息
siti --help

# 显示版本信息
siti --version
```
