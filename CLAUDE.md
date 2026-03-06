# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

`siti-cli` 是一个 macOS/Linux 个人命令行工具集，通过 Homebrew 或一键脚本安装，支持 AI API 配置切换、代理管理、端口管理、网络检测等功能。

## 开发模式运行

在源码目录中直接运行（开发模式）：

```bash
./bin/siti --help
./bin/siti ai list
./bin/siti proxy check
```

无需构建步骤，直接修改 `src/commands/` 下的 `.sh` 文件即可立即生效。

## 版本发布流程

1. 修改 `bin/siti` 中的 `VERSION` 变量
2. 更新 `CHANGELOG.md`
3. 推送 git tag（如 `v1.2.5`），GitHub Actions 自动：
   - 计算 archive SHA256
   - 更新 `Formula/siti-cli.rb`（url + sha256）
   - 创建 GitHub Release

## 架构

### 核心机制：Shell Wrapper + 退出码协议

最关键的设计：部分命令（`ai switch`、`proxy on/off`）需要修改**调用方** shell 的环境变量。由于子进程无法修改父进程环境，采用以下协议：

- 命令脚本输出 `export VAR=value;` 形式的 shell 语句，然后以**退出码 10** 退出
- Shell wrapper（通过 `siti init zsh` 安装）检测到退出码 10 时，对输出执行 `eval`
- 用户无需手动 eval，wrapper 透明处理

```
用户运行: siti proxy on
→ wrapper 捕获输出 + 退出码
→ 退出码=10 → eval 输出 → 当前 shell 环境变量被修改
→ 退出码≠10 → 直接打印输出
```

### 目录结构

```
bin/siti              # 主入口：检测安装方式、路由到 src/commands/
src/commands/         # 各子命令实现（每个文件一个命令）
  ai.sh               # siti ai - AI 服务商切换
  proxy.sh            # siti proxy - 代理管理
  init.sh             # siti init - 生成 shell wrapper
  upgrade.sh          # siti upgrade - 版本升级
  ...
Formula/siti-cli.rb   # Homebrew Formula（由 CI 自动更新）
completions/          # zsh/_siti 和 bash/siti.bash 补全脚本
scripts/              # 安装/卸载后处理脚本
```

### 安装方式检测

`bin/siti` 通过 `SCRIPT_DIR` 自动检测安装方式，并设置 `COMMANDS_DIR`：
- `homebrew`：`/opt/homebrew/share/siti-cli/commands` 或 `/usr/local/share/siti-cli/commands`
- `standalone`：`~/.siti-cli/src/commands`（git clone 到本地）
- `source`：`../src/commands`（开发模式）

`INSTALL_METHOD` 和 `VERSION` 作为环境变量传递给子命令脚本。

### 用户自定义命令

用户可在 `~/.siti-cli/commands/` 放置自定义 `.sh` 文件，优先级高于系统命令。

### 新增命令规范

在 `src/commands/` 新建 `<命令名>.sh`，文件头注释格式：

```bash
#!/bin/bash

# 描述: 命令一行描述（显示在 siti --help 中）
# 补全:
#   subcommand1: 子命令1描述（用于 zsh 自动补全）
#   subcommand2: 子命令2描述
# 用法:
#   siti <cmd> subcommand1  说明
#   siti <cmd> subcommand2  说明
```

需要修改当前 shell 环境变量的命令：输出 shell 语句并以 `exit 10` 退出。

### AI 服务商配置

`siti ai` 命令从 `~/.zshrc` 读取 `*_BASE_URL` 变量来发现服务商，切换时修改 `ANTHROPIC_BASE_URL`、`ANTHROPIC_AUTH_TOKEN`（可选 `ANTHROPIC_MODEL`）。跳过列表通过 `SITI_AI_SKIP` 变量（逗号分隔大写名称）配置。
