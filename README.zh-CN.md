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

### AI 能做什么

- 用简单语言解释发现了什么
- 建议需要关注或清理的内容
- 帮助分类不确定的项目

### AI 不能做什么

- ❌ 不能执行破坏性命令
- ❌ 不能绕过安全检查
- ❌ 不能修改你的系统

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

### 所有 LLM 配置选项

| 变量 | 说明 | 示例 |
|------|------|------|
| `CLAWREMOVE_LLM_PROVIDER` | 提供商类型 | `openai`, `anthropic`, `openai-compatible` |
| `CLAWREMOVE_LLM_API_KEY` | API 密钥（通用） | `sk-xxx` |
| `OPENAI_API_KEY` | OpenAI 密钥 | `sk-xxx` |
| `ANTHROPIC_API_KEY` | Anthropic 密钥 | `sk-ant-xxx` |
| `CLAWREMOVE_LLM_BASE_URL` | 自定义 API 地址 | `https://api.example.com/v1` |
| `CLAWREMOVE_LLM_MODEL` | 模型覆盖 | `gpt-4`, `claude-3-sonnet` |
| `CLAWREMOVE_LLM_TIMEOUT_SECONDS` | 请求超时 | `60` |

### 带 AI 的命令

```bash
# 审计并获取 AI 解释
claw-remove audit --product openclaw --ai

# 获取 AI 分析报告
claw-remove explain --product openclaw --ai

# JSON 输出
claw-remove explain --product openclaw --ai --json
```

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

- `openai`
- `anthropic`
- `openai-compatible`

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
export CLAWREMOVE_LLM_MODEL="gpt-4.1-mini"
claw-remove explain --product openclaw --ai --json
```

Anthropic 示例：

```bash
export CLAWREMOVE_LLM_PROVIDER="anthropic"
export ANTHROPIC_API_KEY="..."
export CLAWREMOVE_LLM_MODEL="claude-3-5-sonnet-latest"
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

## 构建

本地：

```bash
go test ./...
go build -o dist/claw-remove ./cmd/claw-remove
```

多平台：

```bash
./scripts/build.sh
```

PowerShell：

```powershell
./scripts/build.ps1
```

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

## 路线图

完整开发路线图见 [docs/PLAN.md](./docs/PLAN.md)。

该文档使用英文维护，目的是让人类开发者和 agent 都能围绕同一份计划持续迭代。
