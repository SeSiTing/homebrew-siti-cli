# siti-cli 快速开始

## 🎯 5 分钟上手

### 1. 安装

```bash
curl -fsSL https://raw.githubusercontent.com/roooooowing/siti-cli/main/install.sh | bash
source ~/.zshrc
```

### 2. 验证

```bash
siti --version
siti --help
```

### 3. 开始使用

```bash
# AI 配置管理
siti ai list              # 列出所有 AI 服务商
siti ai switch minimax    # 切换到 MiniMax
siti ai current           # 查看当前配置

# 代理管理
siti proxy on             # 开启代理
siti proxy off            # 关闭代理
siti proxy check          # 查看状态

# 端口管理
siti killports 3000       # 释放 3000 端口
siti killports 3000-3010  # 释放端口范围
```

---

## 📋 核心功能

### AI 配置管理

快速切换 AI 服务商（MiniMax、智谱、OpenRouter 等）：

```bash
# 1. 列出所有可用服务商
$ siti ai list
可用的 AI 服务商:
  • minimax         https://api.minimaxi.com/anthropic
  • zhipu           https://open.bigmodel.cn/api/paas/v4
  • openrouter      https://openrouter.ai/api/v1

# 2. 切换服务商
$ siti ai switch zhipu
✅ 已切换到 zhipu

# 3. 查看当前配置
$ siti ai current
当前 AI API 配置:
  服务商: zhipu
  BASE_URL: https://open.bigmodel.cn/api/paas/v4

# 4. 测试配置
$ siti ai test
🔍 测试 AI API 配置...
  ✅ BASE_URL: https://open.bigmodel.cn/api/paas/v4
  ✅ AUTH_TOKEN: sk-xxx...
```

**前提条件：** 需要在 `~/.zshrc` 中配置环境变量：

```bash
# AI 服务商配置
export MINIMAX_BASE_URL="https://api.minimaxi.com/anthropic"
export MINIMAX_API_KEY="your-api-key"

export ZHIPU_BASE_URL="https://open.bigmodel.cn/api/paas/v4"
export ZHIPU_API_KEY="your-api-key"

# 当前使用的配置
export ANTHROPIC_BASE_URL="$MINIMAX_BASE_URL"
export ANTHROPIC_AUTH_TOKEN="$MINIMAX_API_KEY"
```

---

### 代理管理

一键开关终端代理：

```bash
# 开启代理
$ siti proxy on
✅ 终端代理已开启 (127.0.0.1:7890)

# 验证
$ echo $http_proxy
http://127.0.0.1:7890

# 关闭代理
$ siti proxy off
🚫 终端代理已关闭

# 查看状态
$ siti proxy check
当前代理状态:
  ✅ 代理已开启
  http_proxy:  http://127.0.0.1:7890
```

---

### 端口管理

快速释放被占用的端口：

```bash
# 释放单个端口
siti killports 3000

# 释放多个端口
siti killports 3000 5000 8080

# 释放端口范围
siti killports 3000-3010

# 仅检查，不释放
siti killports check
```

---

### 其他工具

```bash
# 网络检测
siti netcheck

# 显示 IP
siti ipshow

# 清理日志
siti cleanlogs

# 备份配置
siti backup-zshrc
```

---

## 🔧 自定义命令

### 添加你自己的命令

在 `~/.siti/commands/` 创建脚本：

```bash
cat > ~/.siti/commands/hello.sh << 'EOF'
#!/bin/bash
# 描述: 打招呼
echo "Hello, $(whoami)!"
EOF

chmod +x ~/.siti/commands/hello.sh
siti hello
```

### 创建需要修改环境变量的命令

使用 `exit 10` 标记：

```bash
cat > ~/.siti/commands/myenv.sh << 'EOF'
#!/bin/bash
# 描述: 设置我的环境

echo "export MY_VAR='my_value';"
echo "echo '✅ 环境变量已设置';"
exit 10  # 告诉包装函数需要 eval
EOF

chmod +x ~/.siti/commands/myenv.sh
siti myenv  # 自动在当前 shell 生效
```

---

## 💡 使用技巧

### 1. Shell 补全

按 `Tab` 键自动补全命令和参数：

```bash
siti <Tab>          # 显示所有命令
siti ai <Tab>       # 显示 ai 子命令
siti ai switch <Tab> # 显示所有服务商
```

### 2. 查看帮助

```bash
siti --help         # 查看所有命令
siti ai --help      # 查看 ai 命令帮助
siti proxy --help   # 查看 proxy 命令帮助
```

### 3. 快速切换工作流

```bash
# 开发环境设置
siti proxy on
siti ai switch minimax

# 生产环境设置
siti proxy off
siti ai switch zhipu
```

---

## 🎓 下一步

- 查看 [完整文档](../README.md)
- 了解 [安装详情](INSTALL.md)
- 查看 [更新日志](../CHANGELOG.md)

---

## 🆘 获取帮助

- **GitHub Issues**: https://github.com/roooooowing/siti-cli/issues
- **查看帮助**: `siti --help`
- **查看版本**: `siti --version`
