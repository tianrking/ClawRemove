# ClawRemove

ClawRemove 是一个专业的跨平台 claw 卸载引擎。

它的目标非常单纯：基于证据发现目标产品残留，生成卸载计划，执行清理，并在最后做残留验证。它不是一个泛用“垃圾清理器”，也不会为了表现“智能”而到处修改系统。

## 文档

- English: [README.md](./README.md)
- Español: [README.es.md](./README.es.md)

## 当前支持

当前内置 provider：

- `openclaw`

架构已经支持后续增加更多 claw 系列产品，但当前重点仍然是先把 OpenClaw 做到专业和可靠。

## 设计原则

- 基于证据发现，不靠模糊猜测乱删
- 先审计，再计划，再执行
- 高风险操作显式开关，不默认执行
- 不安装常驻服务，不写额外数据库，不乱留状态
- 核心引擎与产品规则分离，便于后续扩展

## 命令

```bash
claw-remove products
claw-remove audit --product openclaw --json
claw-remove plan --product openclaw --json
claw-remove apply --product openclaw --dry-run
claw-remove apply --product openclaw
claw-remove verify --product openclaw --json
```

### 命令说明

- `products`
  列出当前编译进程序里的产品 provider。
- `audit`
  只读审计，不执行删除。
- `plan`
  生成卸载计划，不执行。
- `apply`
  执行计划中的动作。
- `verify`
  卸载后的验证扫描。

## 常用参数

- `--product`
  指定产品 provider，当前默认是 `openclaw`
- `--json`
  输出结构化 JSON
- `--dry-run`
  只预演，不真正执行
- `--keep-cli`
  保留包管理器卸载动作和 CLI wrapper 清理
- `--keep-app`
  保留桌面应用及其数据
- `--keep-workspace`
  保留工作区
- `--keep-shell`
  保留 shell completion/profile 清理
- `--kill-processes`
  显式允许终止匹配进程
- `--remove-docker`
  显式允许删除 Docker/Podman 容器和镜像

## 可发现的残留

根据 provider 和平台规则，ClawRemove 可检测：

- 状态目录
- 工作区目录
- 临时目录和日志目录
- 应用 bundle 与应用数据
- launchd / systemd / scheduled tasks
- npm / pnpm / bun / Homebrew 安装
- shell profile / completion 残留
- 匹配进程
- 监听端口
- crontab 残留
- Docker / Podman 容器与镜像

## 风险分层

ClawRemove 把动作分成三类：

- 低风险
  明确属于目标产品的状态、临时目录、wrapper、App 残留
- 中风险
  服务停用/卸载、包管理器卸载
- 高风险
  进程终止、容器删除、镜像删除

高风险动作必须显式开关。

## 项目结构

```text
cmd/claw-remove            CLI 入口
internal/app               CLI 命令编排
internal/core              核心引擎
internal/discovery         发现层
internal/plan              计划层
internal/executor          执行层
internal/output            文本/JSON 输出
internal/products          provider 注册
internal/products/openclaw OpenClaw provider
internal/system            系统命令执行封装
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

最后验证：

```bash
claw-remove verify --product openclaw --json
```

## 仓库说明

本仓库刻意把真正提交的产品代码放在 `ClawRemove/` 目录中。

外层工作区可用于：

- 分析其他 claw 产品
- 运行 agent 做研究
- 放参考仓库
- 做临时实验

这些外层内容默认不会提交到版本库。
