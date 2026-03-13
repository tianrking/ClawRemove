<div align=”center”>
  <h1>ClawRemove</h1>
  <p><strong>一个克制、专业、跨平台的 AI Agent 卸载引擎。</strong></p>
  <p>
    <a href=”https://github.com/tianrking/ClawRemove/actions/workflows/ci.yml”><img src=”https://github.com/tianrking/ClawRemove/actions/workflows/ci.yml/badge.svg” alt=”CI”></a>
    <a href=”./LICENSE”><img src=”https://img.shields.io/badge/license-MIT-1f6feb” alt=”MIT License”></a>
    <img src=”https://img.shields.io/badge/go-1.25%2B-00ADD8?logo=go” alt=”Go 1.25+”>
    <img src=”https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-111827” alt=”Platform support”>
    <a href=”https://github.com/tianrking/ClawRemove/releases”><img src=”https://img.shields.io/github/v/release/tianrking/ClawRemove” alt=”Latest release”></a>
  </p>
  <p><a href=”./README.md”>English</a> | 中文 | <a href=”./README.es.md”>Español</a></p>
</div>

ClawRemove 是一个使用 Go 编写的专业跨平台 AI Agent 卸载引擎。

它的目标非常单纯：基于证据发现 AI Agent（如 OpenClaw、NanoBot、PicoClaw 等）的残留，生成卸载计划，执行清理，并在最后做残留验证。它不是一个泛用垃圾清理器，也不会为了表现”智能”而到处修改系统。

## 文档

- English: [README.md](./README.md)
- Español: [README.es.md](./README.es.md)
- 开发计划: [docs/PLAN.md](./docs/PLAN.md)
- 架构说明: [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md)

## 支持的 AI Agent

| Agent | 描述 | 状态目录 |
|-------|------|---------|
| OpenClaw | AI 助手平台 | ~/.openclaw |
| NanoBot | Python AI Agent | ~/.nanobot |
| PicoClaw | Go AI Agent | ~/.picoclaw |
| OpenFang | AI Agent 框架 | ~/.openfang |
| ZeroClaw | 轻量级 AI Agent | ~/.zeroclaw |
| NanoClaw | 小型 AI Agent | ~/.nanoclaw |
| Cursor | AI 驱动 IDE | ~/.cursor |
| Windsurf | Codeium AI IDE | ~/.windsurf |
| Aider | AI 结对编程 CLI | ~/.aider |

## 当前状态

ClawRemove 目前处于持续建设阶段。

当前主目标是把 AI Agent 卸载能力打磨到发布级，同时把底层抽象成可扩展的 provider 引擎，为后续更多 AI Agent 扩展和上位机接入打基础。

现在每个 provider 不只描述事实规则，也可以声明自己的 skills 和只读 tools，用于后续持续增强分析与智能能力。

## 设计原则

- 基于证据发现，不靠模糊猜测乱删
- 先审计，再计划，再执行
- 高风险操作显式开关，不默认执行
- 不安装常驻服务，不写额外数据库，不乱留状态
- 核心引擎与产品规则分离，便于后续扩展
- 每个 provider 可以声明自己的 skills 和只读 tools
- 未来如果接入 LLM，也只能做分析辅助，不能越权直接删除系统内容

## 为什么是 ClawRemove

- 不是泛用清理器，而是专门做卸载
- 默认保守，先看计划再执行
- 证据优先，不靠模糊猜测乱删
- 输出既适合人看，也适合自动化系统消费
- 项目结构适合持续用 agent 演进

## AI 分析（可选）

ClawRemove 可以使用 AI 来解释发现结果。这是**可选功能** - 所有核心功能无需 AI 也能工作。

### 核心功能（无需 LLM）

| 命令 | 说明 |
|------|------|
| `claw-remove environment` | 完整环境检测 |
| `claw-remove inventory` | AI 运行时和 Agent 清单 |
| `claw-remove security` | API Key 泄露审计 |
| `claw-remove hygiene` | 存储使用分析 |
| `claw-remove audit --product X` | 发现残留 |
| `claw-remove plan --product X` | 生成删除计划 |
| `claw-remove apply --product X` | 执行清理 |
| `claw-remove verify --product X` | 验证清理结果 |

### AI 增强功能（需要 LLM）

| 命令 | 说明 |
|------|------|
| `claw-remove explain --ai` | 用自然语言解释发现结果 |
| `claw-remove audit --ai` | 审计 + AI 解释 |
| `claw-remove verify --ai` | 验证 + AI 解释 |

### AI 能做什么 / 不能做什么

**AI 能做：**
- ✅ 用简单语言解释发现了什么
- ✅ 建议需要关注或清理的内容
- ✅ 帮助分类不确定的项目

**AI 不能做：**
- ❌ 执行破坏性命令
- ❌ 绕过安全检查
- ❌ 修改你的系统

### 快速配置

**方式一：OpenAI**
```bash
# 设置 API Key
export OPENAI_API_KEY="sk-你的密钥"

# 使用 AI 分析
claw-remove explain --product openclaw --ai
```

**方式二：Anthropic Claude**
```bash
# 设置 API Key
export ANTHROPIC_API_KEY="sk-ant-你的密钥"
export CLAWREMOVE_LLM_PROVIDER="anthropic"

# 使用 AI 分析
claw-remove explain --product openclaw --ai
```

**方式三：OpenAI 兼容接口（任意提供商）**
```bash
# 配置你的提供商
export CLAWREMOVE_LLM_PROVIDER="openai-compatible"
export CLAWREMOVE_LLM_BASE_URL="https://你的提供商.com/v1"
export CLAWREMOVE_LLM_API_KEY="你的密钥"

# 使用 AI 分析
claw-remove explain --product openclaw --ai
```

**方式四：Anthropic 兼容接口（如智谱AI等）**
```bash
# 配置 Anthropic 格式的提供商
export CLAWREMOVE_LLM_PROVIDER="anthropic-compatible"
export CLAWREMOVE_LLM_BASE_URL="https://open.bigmodel.cn/api/paas/v4"
export CLAWREMOVE_LLM_API_KEY="你的智谱API密钥"
export CLAWREMOVE_LLM_MODEL="glm-4-plus"

# 使用 AI 分析
claw-remove explain --product openclaw --ai
```

### 支持的提供商类型

| 类型 | API 格式 | 适用场景 |
|------|----------|----------|
| `openai` | OpenAI 原生 | OpenAI 官方 API |
| `anthropic` | Anthropic 原生 | Claude 官方 API |
| `openai-compatible` | `/chat/completions` | 第三方 OpenAI 兼容服务 |
| `anthropic-compatible` | `/messages` | 第三方 Anthropic 兼容服务（如智谱AI） |

### 所有 LLM 配置选项

| 变量 | 说明 | 示例 |
|------|------|------|
| `CLAWREMOVE_LLM_PROVIDER` | 提供商类型 | `openai`, `anthropic`, `openai-compatible` |
| `CLAWREMOVE_LLM_API_KEY` | API 密钥（通用） | `sk-xxx` |
| `OPENAI_API_KEY` | OpenAI 密钥 | `sk-xxx` |
| `ANTHROPIC_API_KEY` | Anthropic 密钥 | `sk-ant-xxx` |
| `CLAWREMOVE_LLM_BASE_URL` | 自定义 API 地址 | `https://api.example.com/v1` |
| `CLAWREMOVE_LLM_MODEL` | 模型覆盖 | `GPT-5.4`, `claude-opus-4-6` |
| `CLAWREMOVE_LLM_TIMEOUT_SECONDS` | 请求超时 | `60` |

### 带 AI 的命令

```bash
# 审计并获取 AI 解释
claw-remove audit --product openclaw --ai

# 获取 AI 分析报告
claw-remove explain --product openclaw --ai

# JSON 输出
claw-remove explain --product openclaw --ai --json

# 完整环境分析
claw-remove environment --ai
```

### ReAct 深度分析

启用 `--ai` 后，ClawRemove 使用 **ReAct（推理+行动）** 机制进行全面分析：

**分析能力：**
| 工具 | 平台 | 说明 |
|------|------|------|
| `deep_analysis` | 全平台 | 所有发现的 artifact 综合概览 |
| `registry_probe` | Windows | 注册表启动项、卸载条目 |
| `env_probe` | 全平台 | PATH 修改、API keys、配置变量 |
| `hosts_probe` | 全平台 | Hosts 文件域名映射 |
| `autostart_probe` | 全平台 | launchd、systemd、cron、注册表自启动 |
| `shell_profile_probe` | 全平台 | Shell 别名、补全、PATH 修改 |

**AI 分析内容：**
- **文件系统**：状态目录、工作区、临时文件
- **服务**：系统服务、后台进程
- **环境变量**：PATH、API keys、自定义变量
- **网络**：Hosts 条目、端口监听
- **自启动**：launchd (macOS)、systemd (Linux)、注册表 (Windows)、cron

**示例输出：**
```
AI Analysis Summary:
- 发现 49 个确认残留项
- 检测到修改：filesystem, services, autostart
- 注册表启动项: HKCU\Software\OpenClaw
- PATH 在 ~/.zshrc 中被修改
- 建议: remove_confirmed_residue (风险=中等)
```

### 实时流式输出

使用 `--ai` 模式时，ClawRemove 会实时显示 AI 分析进度：

```bash
claw-remove explain --product openclaw --ai
```

**实时进度示例：**
```
🤖 AI Analysis Starting...
   Provider: openclaw
   Command: explain

🔄 ReAct Step 1/10...
   📤 Calling LLM...
   💭 Thought: 分析发现的 artifact 中 agent 的修改...
   🔧 Using tool: deep_analysis
   ✅ Tool result received

🔄 ReAct Step 2/10...
   📤 Calling LLM...
   💭 Thought: 调查 shell profile 修改...
   🔧 Using tool: shell_profile_probe
   ✅ Tool result received

✅ AI Analysis Complete!
   📝 Summary: 发现 49 个确认残留项。建议删除...
```

**流式输出特性：**
- 每个 ReAct 步骤实时显示进度
- 显示 AI 正在使用的工具
- 展示 AI 的思考过程摘要
- 仅在终端模式下流式输出（`--json` 或 `--quiet` 时禁用）

## 命令

### 环境检测

```bash
# 完整环境报告
claw-remove environment

# AI 清单
claw-remove inventory

# 安全审计
claw-remove security

# 存储分析
claw-remove hygiene

# JSON 输出
claw-remove environment --json
claw-remove security --json
```

### Agent 清理

```bash
claw-remove products
claw-remove audit --product openclaw --json
claw-remove plan --product openclaw --json
claw-remove apply --product openclaw --dry-run
claw-remove apply --product openclaw
claw-remove apply --product openclaw --yes
claw-remove verify --product openclaw --json
claw-remove explain --product openclaw --json
```

### 命令说明

**环境检测命令:**
- `environment`
  完整环境检测报告（运行时、Agent、产物、安全、存储）
- `inventory`
  AI 运行时和 Agent 清单
- `security`
  AI 工具安全审计（API 密钥泄露检测）
- `hygiene`
  AI 存储使用分析

**清理命令:**
- `products`
  列出当前编译进程序里的产品 provider
- `audit`
  只读审计，不执行删除
- `plan`
  生成卸载计划，不执行
- `apply`
  在交互确认后执行计划中的动作
- `verify`
  卸载后的验证扫描
- `explain`
  在确定性发现结果之上输出受控的 advisory 分析

## 常用参数

- `--product`
  指定产品 provider，当前默认是 `openclaw`
- `--json`
  输出结构化 JSON
- `--ai`
  在报告里附加受控的 advisory 分析
- `--dry-run`
  只预演，不真正执行
- `--yes`
  跳过交互确认，只适合已经审核过计划后的自动化场景
- `--keep-cli`
  保留包管理器卸载动作和 CLI wrapper 清理
- `--keep-app`
  保留桌面应用及其数据
- `--keep-workspace`
  保留工作区
- `--keep-shell`
  保留 shell completion 和 profile 清理
- `--kill-processes`
  显式允许终止匹配进程
- `--remove-docker`
  显式允许删除 Docker 或 Podman 容器和镜像

## LLM 配置

ClawRemove 可以在 `audit`、`verify` 和 `explain` 中挂载受控 advisor。

如果没有配置 LLM，ClawRemove 会自动回退到确定性的 advisory 输出。

支持的 provider：

| 类型 | API 格式 | 说明 |
|------|----------|------|
| `openai` | OpenAI 原生 | OpenAI 官方 API |
| `anthropic` | Anthropic 原生 | Claude 官方 API |
| `openai-compatible` | `/chat/completions` | 第三方 OpenAI 兼容服务 |
| `anthropic-compatible` | `/messages` | 第三方 Anthropic 兼容服务 |

环境变量：

- `CLAWREMOVE_LLM_PROVIDER`
  可选 `openai`、`anthropic`、`openai-compatible`
- `CLAWREMOVE_LLM_API_KEY`
  通用 API Key 覆盖
- `OPENAI_API_KEY`
  当 provider 为 `openai` 时的回退 key
- `ANTHROPIC_API_KEY`
  当 provider 为 `anthropic` 时的回退 key
- `CLAWREMOVE_LLM_BASE_URL`
  provider base URL 覆盖
- `CLAWREMOVE_LLM_MODEL`
  模型名覆盖
- `CLAWREMOVE_LLM_MAX_TOKENS`
  advisory 输出 token 上限
- `CLAWREMOVE_LLM_MAX_STEPS`
  受控 ReAct 最大步数
- `CLAWREMOVE_LLM_TIMEOUT_SECONDS`
  请求超时秒数

OpenAI 示例：

```bash
export CLAWREMOVE_LLM_PROVIDER="openai"
export OPENAI_API_KEY="..."
export CLAWREMOVE_LLM_MODEL="GPT-5.4"
claw-remove explain --product openclaw --ai --json
```

Anthropic 示例：

```bash
export CLAWREMOVE_LLM_PROVIDER="anthropic"
export ANTHROPIC_API_KEY="..."
export CLAWREMOVE_LLM_MODEL="claude-opus-4-6-20250205"
claw-remove explain --product openclaw --ai --json
```

其他 OpenAI-compatible 示例：

```bash
export CLAWREMOVE_LLM_PROVIDER="openai-compatible"
export CLAWREMOVE_LLM_BASE_URL="https://your-provider.example/v1"
export CLAWREMOVE_LLM_API_KEY="..."
export CLAWREMOVE_LLM_MODEL="your-model-name"
claw-remove explain --product openclaw --ai --json
```

## 安全删除流程

推荐使用方式：

1. `audit`
   先看发现结果
2. `verify`
   区分 confirmed residuals 和 investigate residuals
3. `explain --ai`
   让受控 advisor 总结重点
4. `apply`
   查看预演结果并输入确认短语
5. `apply --yes`
   只在已经完成审核后用于自动化

默认情况下，`apply` 不会静默全自动执行。

它会先展示预演结果，再要求用户输入确认短语，确保删除动作可控且安全。

如果自动化流程需要 JSON 输出，建议先用 `plan` 或 `verify` 做审核，再使用 `apply --yes`。

## 可发现的残留

根据 provider 和平台规则，ClawRemove 可检测：

- 状态目录
- 工作区目录（由 provider 声明的工作区子目录名称）
- 临时目录和日志目录
- 应用 bundle 与应用数据
- launchd / systemd / scheduled tasks
- npm / pnpm / bun / pip / Homebrew 安装
- shell profile / completion 残留（通过内容扫描验证实际包含 marker，而非仅路径匹配）
- 匹配进程
- 监听端口（由 provider 的事实规则声明，不硬编码）
- crontab 残留
- Docker / Podman 容器与镜像
- Windows 注册表键值
- 环境变量
- hosts 文件条目

## 风险分层

ClawRemove 把动作分成三类：

- 低风险
  明确属于目标产品的状态、临时目录、wrapper、App 残留
- 中风险
  服务停用或卸载、包管理器卸载
- 高风险
  进程终止、容器删除、镜像删除

高风险动作必须显式开关。

## 项目结构

```text
cmd/claw-remove            CLI 入口
internal/app               CLI 命令编排
internal/core              核心引擎
internal/discovery         发现层
internal/evidence          构建发现结果与计划间的证据桥梁
internal/plan              计划层
internal/executor          执行动作
internal/llm               advisor 控制器与编排
internal/llm/prompts       prompt 模板定义
internal/llm/providers     多模型客户端适配
internal/output            文本和 JSON 输出
internal/platform          跨平台底层适配支持
internal/products          provider 注册
internal/products/openclaw OpenClaw provider
internal/skills            provider 技能定义
internal/tools             provider 工具定义
internal/model             统一请求、发现及验证模型
internal/system            系统命令执行封装
internal/products/openclaw OpenClaw provider
internal/system            系统命令执行封装
docs                       路线图和开发计划
scripts                    构建脚本
dist                       本地构建产物
```

## 安装

### 预编译二进制

从 [GitHub Releases](https://github.com/tianrking/ClawRemove/releases) 下载。

**macOS:**
```bash
# DMG（推荐）
# 下载 claw-remove-VERSION-macOS.dmg，打开并拖到应用程序

# 或通过压缩包
curl -sL https://github.com/tianrking/ClawRemove/releases/latest/download/claw-remove-VERSION-darwin-arm64.tar.gz | tar xz
sudo mv claw-remove /usr/local/bin/
```

**Linux:**
```bash
# Debian/Ubuntu（deb 包）
wget https://github.com/tianrking/ClawRemove/releases/latest/download/claw-remove_VERSION_amd64.deb
sudo dpkg -i claw-remove_VERSION_amd64.deb

# RHEL/Fedora（rpm 包）
wget https://github.com/tianrking/ClawRemove/releases/latest/download/claw-remove-VERSION-1.x86_64.rpm
sudo rpm -i claw-remove-VERSION-1.x86_64.rpm

# Arch Linux
wget https://github.com/tianrking/ClawRemove/releases/latest/download/claw-remove-VERSION-1-x86_64.pkg.tar.zst
sudo pacman -U claw-remove-VERSION-1-x86_64.pkg.tar.zst

# AppImage（通用，无需安装）
wget https://github.com/tianrking/ClawRemove/releases/latest/download/claw-remove-VERSION-x86_64.AppImage
chmod +x claw-remove-VERSION-x86_64.AppImage
./claw-remove-VERSION-x86_64.AppImage

# 或通过压缩包
curl -sL https://github.com/tianrking/ClawRemove/releases/latest/download/claw-remove-VERSION-linux-amd64.tar.gz | tar xz
sudo mv claw-remove /usr/local/bin/
```

**Windows:**
```powershell
# 下载对应架构的 ZIP
# claw-remove-VERSION-windows-amd64.zip (x64)
# claw-remove-VERSION-windows-arm64.zip (ARM64)
# claw-remove-VERSION-windows-386.zip (32位)

# 解压并添加到 PATH
Expand-Archive claw-remove-VERSION-windows-amd64.zip -DestinationPath C:\Tools\claw-remove
$env:PATH += ";C:\Tools\claw-remove"
```

**BSD:**
```bash
# FreeBSD
curl -sL https://github.com/tianrking/ClawRemove/releases/latest/download/claw-remove-VERSION-freebsd-amd64.tar.gz | tar xz
sudo mv claw-remove /usr/local/bin/
```

### 从源码安装

```bash
go install github.com/tianrking/ClawRemove/cmd/claw-remove@latest
```

## 构建

### 本地构建

```bash
go test ./...
go build -o dist/claw-remove ./cmd/claw-remove
```

### 使用 GoReleaser 发布

```bash
# 安装 GoReleaser
go install github.com/goreleaser/goreleaser/v2@latest

# 本地快照构建
goreleaser build --snapshot --clean

# 完整发布（需要 tag）
goreleaser release --clean
```

### 发布产物

每个版本包含：

| 格式 | 平台 | 说明 |
|------|------|------|
| `.dmg` | macOS | 磁盘镜像安装器 |
| `.deb` | Linux (Debian/Ubuntu) | APT 包 |
| `.rpm` | Linux (RHEL/Fedora) | RPM 包 |
| `.pkg.tar.zst` | Linux (Arch) | Arch Linux 包 |
| `.AppImage` | Linux (通用) | 便携式可执行文件 |
| `.zip` | Windows | 压缩包 |
| `.tar.gz` | 所有平台 | 压缩包 |

支持的平台（共 22 个）：

**macOS:**
- `darwin-amd64` (Intel)
- `darwin-arm64` (Apple Silicon)

**Linux:**
- `linux-amd64` (x86_64)
- `linux-arm64` (ARM64)
- `linux-386` (32位)
- `linux-arm` (ARM v7, Raspberry Pi)
- `linux-riscv64` (RISC-V)
- `linux-ppc64le` (IBM Power)
- `linux-s390x` (IBM Z)
- `linux-mips64` (MIPS64 大端)
- `linux-mips64le` (MIPS64 小端)

**Windows:**
- `windows-amd64.exe` (x86_64)
- `windows-arm64.exe` (ARM64)
- `windows-386.exe` (32位)

**BSD:**
- `freebsd-amd64`, `freebsd-arm64`
- `netbsd-amd64`
- `openbsd-amd64`

## 推荐使用流程

先审计：

```bash
claw-remove audit --product openclaw --json
```

再生成计划：

```bash
claw-remove plan --product openclaw --json
```

先 dry-run：

```bash
claw-remove apply --product openclaw --dry-run
```

确认后再真实执行：

```bash
claw-remove apply --product openclaw
```

已经审核后再非交互执行：

```bash
claw-remove apply --product openclaw --yes
```

最后验证：

```bash
claw-remove verify --product openclaw --json
```

查看受控解释分析：

```bash
claw-remove explain --product openclaw --json
```

## 卸载 ClawRemove

完全从系统中移除 ClawRemove：

### macOS (DMG/tarball)

```bash
# 删除二进制文件
sudo rm -f /usr/local/bin/claw-remove
# 如果通过 DMG 安装
sudo rm -rf /Applications/claw-remove.app

# 删除 shell 补全（如果安装了）
rm -f ~/.zsh/completion/_claw-remove
rm -f ~/.bash_completion.d/claw-remove
```

### Linux (deb)

```bash
# 删除包
sudo dpkg --remove claw-remove

# 或完全清除（包括配置）
sudo dpkg --purge claw-remove

# 手动清理（如果从 tarball 安装）
sudo rm -f /usr/local/bin/claw-remove
sudo rm -f /usr/share/man/man1/claw-remove.1.gz
```

### Linux (rpm)

```bash
# 删除包
sudo rpm -e claw-remove

# 手动清理（如果从 tarball 安装）
sudo rm -f /usr/bin/claw-remove
```

### Linux (AppImage)

```bash
# 直接删除 AppImage 文件
rm -f claw-remove-*.AppImage

# 删除桌面集成（如果有）
rm -f ~/.local/share/applications/claw-remove.desktop
rm -f ~/.local/share/icons/claw-remove.png
```

### Windows

```powershell
# 从安装目录删除
Remove-Item -Recurse -Force "C:\Tools\claw-remove"

# 或仅删除 PATH 中的文件
Remove-Item "C:\Tools\claw-remove\claw-remove.exe"
```

### 从源码安装 (go install)

```bash
# 删除二进制
rm -f $(go env GOPATH)/bin/claw-remove

# 清理模块缓存（可选）
go clean -modcache
```

### 验证完全删除

```bash
# 应该返回 "command not found"
which claw-remove || echo "ClawRemove 已成功删除"
```

## 路线图

完整开发路线图见 [docs/PLAN.md](./docs/PLAN.md)。

该文档使用英文维护，目的是让人类开发者和 agent 都能围绕同一份计划持续迭代。
