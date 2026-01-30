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

## 快速开始

### Shell 包装函数（推荐）

为了让命令在当前终端立即生效，建议安装 shell 包装函数：

```bash
# 一次性安装（只需执行一次）
./scripts/setup-shell-wrapper.sh install
source ~/.zshrc

# 之后所有命令都能在当前终端立即生效
```

不安装也可以使用，但需要手动 eval：
```bash
eval "$(siti proxy on)"
eval "$(siti ai switch minimax)"
```

## 使用方法

### AI 配置管理

快速切换 AI 服务商配置（MiniMax、智谱、OpenRouter 等）：

```bash
# 列出所有可用的 AI 服务商
siti ai list

# 切换到 MiniMax
siti ai switch minimax

# 切换到智谱 AI
siti ai switch zhipu

# 查看当前配置
siti ai current

# 测试当前配置
siti ai test
```

**前提条件：** 需要在 `~/.zshrc` 中配置相应的环境变量：
```bash
export MINIMAX_BASE_URL="https://api.minimaxi.com/anthropic"
export MINIMAX_API_KEY="your-api-key"

export ZHIPU_BASE_URL="https://open.bigmodel.cn/api/paas/v4"
export ZHIPU_API_KEY="your-api-key"

# 当前使用的配置
export ANTHROPIC_BASE_URL="$MINIMAX_BASE_URL"
export ANTHROPIC_AUTH_TOKEN="$MINIMAX_API_KEY"
```

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

## Shell 包装函数说明

### 工作原理

siti-cli 采用退出码约定（借鉴 asdf/mise 模式）：
- `exit 0` - 正常输出，不需要 eval
- `exit 1` - 执行失败
- `exit 10` - 成功，但需要 eval（修改当前 shell 环境）

安装 shell 包装函数后，会自动检测退出码并执行相应操作，用户完全无感知。

### 管理包装函数

```bash
# 查看安装状态
./scripts/setup-shell-wrapper.sh status

# 安装
./scripts/setup-shell-wrapper.sh install

# 卸载
./scripts/setup-shell-wrapper.sh uninstall
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

### 创建需要修改环境变量的命令

如果你的命令需要修改当前 shell 环境变量，使用 `exit 10`：

```bash
cat > ~/.siti/commands/myenv.sh << 'EOF'
#!/bin/bash
# 描述: 设置我的环境变量

echo "export MY_VAR='my_value';"
echo "echo '✅ 环境变量已设置';"
exit 10  # 告诉包装函数需要 eval
EOF
chmod +x ~/.siti/commands/myenv.sh
siti myenv  # 自动在当前 shell 生效
```

## 帮助

```bash
# 显示帮助信息
siti --help

# 显示版本信息
siti --version
```
