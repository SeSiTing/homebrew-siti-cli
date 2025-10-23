# siti-cli

个人命令行工具集，用于简化日常开发操作。

## 安装

### Homebrew 安装（推荐）

```bash
brew install roooooowing/tap/siti-cli
```

### 手动安装（开发用）

```bash
git clone https://github.com/roooooowing/siti-cli.git
cd siti-cli
./scripts/post-install.sh
export PATH="$(pwd)/bin:$PATH"
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

### 添加用户自定义命令

在 `~/.siti/commands/` 目录中创建脚本：

```bash
cat > ~/.siti/commands/mycmd.sh << 'EOF'
#!/bin/bash
# 描述: 我的命令
echo "Hello from custom command!"
EOF
chmod +x ~/.siti/commands/mycmd.sh
siti mycmd
```

## 帮助

```bash
# 显示帮助信息
siti --help

# 显示版本信息
siti --version
```
