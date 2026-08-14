# CLAUDE.md

本文件是项目对所有 AI 编码助手的**唯一真源**（Codex/Cursor/Aider/Copilot 通过 [AGENTS.md](./AGENTS.md) 指向这里）。
人类开发者也应优先读本文件再动手。

---

## 项目概述

`siti-cli` 是一个 macOS / Linux 个人命令行工具集，Go + Cobra 实现，通过 Homebrew tap 分发，由 goreleaser 自动发布。
功能：AI 服务商切换、终端代理开关、固定网络配置、SSH tunnel、端口清理、网络检测、IP 查看、日志清理、Homebrew 升级。

## 技术栈

| 领域 | 选型 |
|---|---|
| 语言 | Go 1.24+ |
| CLI 框架 | [spf13/cobra](https://github.com/spf13/cobra) |
| 交互 UI | [charmbracelet/huh](https://github.com/charmbracelet/huh)（Charm 系，活跃） |
| 配置发现 | AI provider 解析 `~/.zshenv` / `~/.zshrc`；network/tunnel profile 解析 `~/.siti/{network,tunnels}/*.yaml` |
| 测试 | 标准库 `testing` + golden 文件 snapshot |
| 发布 | [goreleaser](https://goreleaser.com/) — 交叉编译 + Formula 自动更新 |

## 开发流程

```bash
make help          # 列出所有命令
make build         # 构建到 ./siti
make test          # go test -race -count=1 ./...
make tidy          # go mod tidy + 校验未漂移
make snapshot      # 本地 goreleaser dry-run
go run . ai list   # 直接跑某个命令
```

源码改完直接 `go run .` 验证；不需要先构建。

## 版本发布流程

1. 修改 `version.go` 中的 `var version = "X.Y.Z"`
2. 更新 `CHANGELOG.md`，加一段 ISO 日期 + 版本号
3. push 到 `main` 分支

### 版本号升级规则（AI 助手必须遵守）

本项目是 main 分支即发版的一条流（trunk-based）流程：**每次提交都代表一次发版**。

#### 硬性规则

- **每次提交必须升级 patch 位（z）**：`2.0.9` → `2.0.10`，无例外。
- **AI 只允许升 z**：任何代码变更（bug 修复、重构、文档、测试）都直接 +1，不需要用户提醒或确认。
- **禁止升 minor / major**：除非用户明确说"升 minor"或"升 major"，否则 AI 不得触碰 x 和 y 位。

#### 提交流程

每次 `git commit` 时，AI 必须在同一批 staged 变更中包含：
1. `version.go` 中 z +1
2. `CHANGELOG.md` 中新增对应条目

提交完成后，在输出中追加一行 `→ 版本: vX.Y.Z`，方便用户确认。

#### CI 门禁

`publish-on-version-bump.yml` 会对比上一个 git tag 与新版本：
- 仅 patch 升级：直接放行
- minor 升级：要求 commit message 含 `[minor-bump]`
- major 升级：要求 commit message 含 `[major-bump]`
- 不满足则发版任务失败，强制人工确认

GitHub Actions 自动：
- 从 `version.go` 提取版本 → 创建 `vX.Y.Z` tag → push
- goreleaser 用该 tag 交叉编译（darwin/linux × amd64/arm64）
- 上传 GitHub Release + checksums + 自动更新 `SeSiTing/homebrew-tap` 仓库的 `Formula/siti-cli.rb`

补救：tag 已存在但 release 漏发 → Actions → "Release from Tag" → 手动跑。

**所需 secrets**（仓库 Settings → Secrets and variables → Actions）：
- `HOMEBREW_TAP_TOKEN`：fine-grained PAT，对本仓库 contents: write 权限（同仓库其实 GITHUB_TOKEN 也够，留 PAT 是为了将来分离 tap）

## 架构

### 核心机制：exit-10 协议

部分命令（`ai switch`、`ai clear`、`proxy on/off`）需要修改**调用方**父 shell 的环境变量。
子进程没法改父进程环境，所以约定：

```
用户运行: siti proxy on
  ↓
shell 函数 wrapper（由 `siti init zsh` 输出 + 用户 source）
  ↓ 调用真正的二进制并捕获 stdout / stderr / exit code
binary: cmd/proxy.go RunE
  ↓ 通过 cmd.Eval(c, shell.Export(...)) 把 shell 语句存入 context buffer
  ↓ return nil
cmd/root.go Execute()
  ↓ 检测 buffer 非空 → 打印到 stdout → return 10 → main.go os.Exit(10)
wrapper 看到 exit 10
  ↓ eval stdout（信任契约，不做白名单过滤）
  ↓ stderr 透传给用户
  ↓ return 0   ← 关键：让 `siti ai switch && echo ok` 正常工作
```

设计要点：
- `EvalBuffer` **不实现 error 接口**——eval 通道与 error 通道解耦
- wrapper **不 grep 过滤** stdout——Go 端是单一真源
- `os.Exit` 在全项目**只出现一次**（`main.go`）
- stderr 永远是给人看的，stdout 在 exit 10 时永远是 shell 代码
- `brew update` 只对明确的瞬时网络错误重试一次；证书、权限和 tap 信任错误不得重试
- 内置 `studio` tunnel preset 使用 SSH Host alias `mac-studio`，可被 `~/.siti/tunnels/studio.yaml` 同名覆盖
- 内置 `blacklake-proxy` network preset 要求用户先手动连接 `blacklake`，动态读取当前 DHCP IPv4，确认地址属于目标 `/21` 网段后固定应用该地址及 `172.16.40.2` 网关/DNS；不负责切换 Wi-Fi，可被同名 YAML 覆盖

### 目录结构

```
.
├── main.go                       # os.Exit(cmd.Execute(version))
├── version.go                    # var version = "dev"（CI 检测 + ldflags 注入双角色）
├── go.mod / go.sum
├── Makefile
├── README.md / CHANGELOG.md / CLAUDE.md / AGENTS.md / LICENSE
├── .editorconfig / .gitignore / .goreleaser.yml
│
├── .github/workflows/
│   ├── ci.yml                    # PR 检查：build/vet/test/tidy
│   ├── publish-on-version-bump.yml  # main 推送时自动发版
│   └── release-from-tag.yml      # 手动补发遗漏的 release
│
├── cmd/                          # 命令实现，每个文件一个 namespace
│   ├── root.go                   # rootCmd + Execute() + Eval(c, lines...)
│   ├── ai.go / ai_*.go           # AI 客户端切换、状态、凭证与诊断
│   ├── proxy.go                  # siti proxy on/off/git on|off/status + upgrade 代理预检
│   ├── initcmd.go                # siti init zsh|bash|fish
│   ├── version.go                # siti version
│   ├── net.go                    # siti net apply/reset/status/list/check
│   ├── tunnel.go                 # siti tunnel up/down/status/list
│   ├── ip.go / port.go / logs.go / brew.go / upgrade.go
│   └── util.go                   # 公用工具：lookupEnv / firstNonEmpty
│
├── internal/
│   ├── shell/
│   │   ├── eval.go               # Export/ExportRef/Unset/SourceIf 字符串 helper
│   │   ├── eval_test.go
│   │   ├── wrapper.go            # siti exit-10 + 当前 shell 的 Grok model 注入
│   │   ├── wrapper_test.go       # snapshot 测试（-update 刷新 golden）
│   │   └── testdata/
│   │       ├── wrapper_zsh.golden
│   │       └── wrapper_fish.golden
│   ├── config/
│   │   ├── zshrc.go              # Provider 发现、内置默认值与覆盖
│   │   ├── grok.go               # provider-specific Grok 模型入口
│   │   ├── codex.go              # Codex 全局 managed block + 备份/恢复
│   │   ├── credentials.go        # OS Keychain / Secret Service
│   │   └── *_test.go
│   ├── network/
│   │   ├── profile.go            # YAML profile 与 active 状态
│   │   ├── manager.go            # macOS networksetup 应用、恢复与验证
│   │   └── *_test.go
│   └── tunnel/
│       ├── profile.go            # SSH tunnel YAML profile 与严格校验
│       ├── manager.go            # OpenSSH ControlMaster 生命周期与端口探测
│       └── *_test.go
│
└── completions/                  # cobra 自动生成，goreleaser 也会重新生成
    ├── _siti
    └── siti.bash
```

故意**不存在**：`bin/`、`src/`、`scripts/`、`docs/`、`Formula/`（goreleaser 写到 tap 仓库的 Formula 目录）。

### 新增命令规范

在 `cmd/` 下新建 `<name>.go`，模板：

```go
package cmd

import (
    "fmt"
    "github.com/SeSiTing/siti-cli/internal/shell"
    "github.com/spf13/cobra"
)

var fooCmd = &cobra.Command{
    Use:   "foo",
    Short: "一行描述",
    Args:  cobra.NoArgs,
    RunE: func(c *cobra.Command, args []string) error {
        // 普通命令：直接 fmt.Println / printErr，return nil
        fmt.Println("结果")

        // 需要改父 shell 环境时：
        // Eval(c, shell.Export("FOO", "bar"))
        // return nil
        return nil
    },
}

func init() { rootCmd.AddCommand(fooCmd) }
```

约定：
- stdout 给机器读（exit 10 时是 shell 代码，否则是 `siti ai list` 这类纯输出）
- stderr 给人读（`printErr("✓ 已切换到 %s", name)`）
- `return nil` 表示成功；返回 error 由 cobra 自动打到 stderr 并以 exit 1 结束

### AI 服务商发现

`siti ai` 内置 Ali Coding Plan / Bailian / DeepSeek / MiniMax / 智谱 / OpenRouter 的非敏感协议地址与默认模型，并从 `~/.zshenv` 和 `~/.zshrc` 解析自定义覆盖：

- `export <NAME>_ANTHROPIC_BASE_URL=...`（推荐）或 `<NAME>_BASE_URL=...` → 注册 provider `<NAME>`
- `<NAME>_API_ORIGIN` 仅作为无路径 URL 根地址供 shell 组合，不单独注册 provider
- 同名 `<NAME>_API_KEY` → AuthTokenVar；每个 provider 必须显式使用自己的 Key，不跨 provider 回退
- 同名 `<NAME>_MODEL` → ModelVar（切换时同步设置 5 个 ANTHROPIC_*_MODEL）
- 可选 `<NAME>_CHAT_COMPLETIONS_BASE_URL` → 同一 provider 启用 Grok Build；兼容旧 `<NAME>_OPENAI_BASE_URL` / `<NAME>_GROK_BASE_URL`
- Grok 默认复用 `<NAME>_API_KEY` / `<NAME>_MODEL`，可由 `<NAME>_GROK_API_KEY` / `<NAME>_GROK_MODEL` 覆盖
- Grok 首次切换时在 `~/.grok/config.toml` 自动安装无密钥的 `siti-<provider>` 模型入口
- shell wrapper 读取 `SITI_GROK_MODEL_ID`，在未显式传 `--model` 时为当前 shell 的 `grok` 注入对应模型
- Codex 仅在显式 `--client codex` / `all` 时全局修改 `$CODEX_HOME/config.toml`；写入前备份且不修改 `auth.json`
- Codex Key 通过 `siti ai credential import` 存入系统凭证库，配置使用 command-backed auth，不写明文 Key
- 自定义 Codex 映射使用 `<NAME>_RESPONSES_BASE_URL`，兼容旧 `<NAME>_CODEX_BASE_URL`；端点必须支持 Responses API
- `SITI_AI_SKIP="A,B,C"` → 跳过列表（环境变量优先于 zshrc 解析）
- `ANTHROPIC` 前缀本身被忽略（避免循环引用）

默认 `siti ai switch` 只切当前 shell 的 Claude + 可用 Grok；Codex 必须显式选择。切换前完成全部 Key、协议和 wrapper 预检，失败不修改客户端配置。

## 测试规范

- 所有 `internal/` 包**应有测试**
- `internal/shell/wrapper_test.go` 用 golden 文件做 snapshot；改 wrapper 后跑 `go test ./internal/shell -update` 刷新
- `cmd/` 层不强求测试（cobra 部分主要靠手动验证）
- 提交前必跑 `make test && make tidy`

## 输出风格规范（CLI 文案）

完全对齐主流 CLI（gh / cargo / ripgrep / kubectl / pnpm / bun）的纯文本+符号风格。
**不允许使用 emoji**——emoji 在不同终端宽度不一致、SSH/CI 日志会出锅、屏幕阅读器读不出。

### 符号集（仅这些，不要扩展）

| 用途 | 符号 | 示例 |
|---|---|---|
| 成功 | `✓` (U+2713) | `✓ 已切换到 minimax` |
| 失败 / 关闭 | `✗` (U+2717) | `✗ 终端代理已关闭` |
| 警告 | `!` | `! 发现 3 个端口被占用` |
| 提问 | `?` | `? 是否清理? [y/N]` |
| 处理中 / 提示箭头 | `→` | `→ 扫描端口占用...` / `→ 提示: 运行 'siti ai switch'` |
| 跳过 | `↷` 或文字 `skip` | `↷ 端口 3000` |
| 项目符号 | `•` | `• 步骤 1 失败` |
| 区块分隔 | `─` × N | `────────────────────────` |
| 资源 / 进程标签 | `[xxx]` 小写 | `[java]` `[node]` `[docker]` `[pg]` `[mysql]` `[redis]` `[py]` `[other]` |

### 文案规则

- 区块标题用纯文本：`LAN:` `WAN:` `当前代理状态:`，不用 `🌐 内网 IP:`
- 不写"花式"装饰：删除 `🍺` `🧹` `📦` `💾` `🚀` 等表情
- 时间用纯文字：`took 1.2s` 不用 `⏱`
- stderr 走人类可读格式；stdout 走机器可解析格式（exit-10 时为 shell 代码）
- 所有现有文件已按此规范统一，新增命令必须沿用

## 编辑约束

- 不要写多行注释块；单行注释只在 *为什么* 不显然时加
- 不要为"未来需求"写抽象层；删掉就删掉
- 不要恢复 `EvalDirective` 的 error 接口模式
- 不要在 wrapper 里加 grep 白名单过滤
- 不要新增 `bin/` `src/` `scripts/` `docs/` 这些已被删除的目录

## 当前已知技术债

无。上一轮重构清理完毕。
