# ClawRemove Development Plan

## The Killer Feature

> **One tool, under 10MB, solves all your OpenClaw-related headaches.**

- **Zero dependencies** - Single static binary
- **No installation** - Download and run
- **Cross-platform** - macOS, Linux, Windows, BSD
- **Safe by default** - Audit before any action
- **AI-powered analysis** - Optional LLM advisory with intelligent ReAct loop

## Core Positioning

**Agent Environment Inspector**

Inspect, audit and clean environments where AI agents run.

AI Agent 运行环境检测与治理工具

### What ClawRemove IS

- AI Agent runtime detection
- AI Agent tool detection
- AI Agent artifact detection
- AI Agent environment cleanup

### What ClawRemove IS NOT

- System cleaner (like CCleaner)
- Security scanner (like Nessus)
- AI model trainer
- AI data processor

## Agent Ecosystem Structure

ClawRemove scans 4 layers of the agent ecosystem:

```
┌─────────────────────────────┐
│ Agent Applications          │  Layer 4
│ openclaw / nanobot / cursor │
│ windsurf / aider / cline    │
└─────────────────────────────┘

┌─────────────────────────────┐
│ Agent Frameworks            │  Layer 3
│ LangChain / LangGraph       │
│ AutoGen / CrewAI            │
└─────────────────────────────┘

┌─────────────────────────────┐
│ AI Runtime                  │  Layer 2
│ Ollama / LocalAI / vLLM     │
│ llama.cpp / LM Studio       │
└─────────────────────────────┘

┌─────────────────────────────┐
│ AI Artifacts                │  Layer 1
│ models / embeddings         │
│ vector db / cache           │
└─────────────────────────────┘
```

## Core Capabilities

### 1. Agent Runtime Detection

Detect local LLM runtimes.

```bash
clawremove audit
```

Output:
```
AI Runtime
----------
Ollama detected
  Models: 42GB
  Running: yes
  Port: 11434

LM Studio detected
  Models: 12GB
  Running: no
```

Detection methods:
- Binary presence
- Service status
- Model paths
- Config files

### 2. Agent Tool Detection

Detect agent applications and frameworks.

```bash
clawremove audit
```

Output:
```
Agent Tools
-----------
Applications:
  OpenClaw installed at ~/.openclaw
  Cursor installed at ~/.cursor
  NanoBot installed at ~/.nanobot

Frameworks:
  LangChain installed (pip)
  OpenAI SDK installed (npm)
  Transformers installed (pip)
```

Detection methods:
- Repository paths
- pip/npm packages
- Docker containers
- State directories

### 3. Agent Artifact Detection

Detect AI-generated resources.

```bash
clawremove audit
```

Output:
```
AI Artifacts
------------
Models:
  Ollama: 64GB
  HuggingFace Cache: 21GB
  LM Studio: 12GB

Vector Databases:
  ChromaDB: 8GB at ~/.chromadb

Embedding Cache:
  Sentence Transformers: 3GB
```

Detection paths:
- ~/.ollama/models
- ~/.cache/huggingface
- ~/.chromadb
- ~/.local/share/milvus

### 4. Agent Cleanup

Deep clean agent environments.

```bash
clawremove cleanup --product openclaw
```

Removes:
- Configuration files
- Cache directories
- Background services
- Shell integrations
- Model artifacts (optional)

### 5. Agent Security Audit

Check for exposed API keys in agent configs.

```bash
clawremove security
```

Output:
```
Security Audit
--------------
⚠️  API keys detected in 3 locations:

  ~/.openclaw/.env
    - OPENAI_API_KEY

  ~/.cursor/config.json
    - ANTHROPIC_API_KEY

  Environment variable
    - GEMINI_API_KEY

Recommendation: Move keys to secure secret manager
```

Scope: ONLY AI tool configurations, not general file scanning.

### 6. Agent Machine Hygiene

Analyze AI storage usage.

```bash
clawremove hygiene
```

Output:
```
AI Storage Usage
----------------
Models:      86GB
Cache:       21GB
Vector DB:   12GB
Logs:        2GB
────────────────
Total:       121GB

Recommendations:
- 3 old model versions can be cleaned (save 24GB)
- Unused embedding cache: 8GB
```

## CLI Commands

```bash
# Full environment audit
clawremove audit

# Quick detection
clawremove detect

# Deep cleanup
clawremove cleanup --product openclaw

# Security check
clawremove security

# Storage analysis
clawremove hygiene
```

## Similar Tools

ClawRemove is similar to:

- `brew doctor` - Environment health check
- `docker system df` - Storage analysis
- `npm doctor` - Node environment check

NOT similar to:

- CCleaner (system cleaner)
- Nessus (security scanner)
- Malwarebytes (malware removal)

## Implementation Phases

### Phase 1: Detection Infrastructure ✅

- [x] Agent runtime detection (Ollama, LM Studio, GPT4All, LocalAI)
- [x] Agent framework detection (langchain, openai sdk, transformers)
- [x] Model cache detection (HuggingFace, Torch, TensorFlow)
- [x] Vector store detection (ChromaDB, Pinecone, Weaviate)
- [x] AI tool security scanner (API keys in configs)

### Phase 2: CLI Integration ✅

- [x] Add `environment` command with full environment report
- [x] Add `inventory` command for AI runtime/agent scan
- [x] Add `security` command for API key audit
- [x] Add `hygiene` command for storage analysis
- [x] Update output formatting with JSON support

### Phase 3: Provider Expansion

- [x] Cursor provider
- [x] Windsurf provider
- [x] Aider provider
- [x] Cline provider
- [x] Continue.dev provider

### Phase 4: Advanced Detection ✅

- [x] vLLM detection
- [x] llama.cpp detection
- [x] LangGraph detection
- [x] AutoGen detection
- [x] CrewAI detection

### Phase 4.5: ReAct Intelligence Enhancement ✅

- [x] Smart search tools (search_agent_traces, quick_scan)
- [x] Credential detection probe
- [x] Configuration file analysis
- [x] Batch tool execution support
- [x] Confidence tracking in ReAct loop
- [x] Progress indicators (starting, gathering, analyzing, finalizing)
- [x] Enhanced system prompts for deeper analysis

### Phase 5: Cleanup Enhancement ✅

- [x] Model version cleanup
- [x] Orphaned cache cleanup
- [x] Unused vector db cleanup
- [x] Log rotation

### Phase 6: Performance & UX Enhancement ✅

#### 6.1 Performance Optimization
- [x] Parallel scanning (Discovery, Inventory, Cleanup scanners)
- [x] dirSize optimization (with caching)
- [x] File system operation caching
- [ ] Lazy loading for Discovery fields
- [ ] Timeout limits for large directory traversal

#### 6.2 Output Format Enhancement
- [x] YAML output format support
- [x] HTML report generation
- [x] Formatter interface abstraction
- [x] Output format CLI flag (`--format text/json/yaml/html`)

#### 6.3 TUI Enhancement
- [x] Pagination support for large lists
- [x] Search/filter functionality
- [ ] Detail panel for selected items
- [ ] Progress bar with percentage
- [ ] Theme/color scheme support
- [ ] Extended coverage (environment, security commands)

#### 6.4 ReAct Enhancement (Vertical)
- [x] Parallel batch tool execution
- [ ] Tool result caching
- [ ] Improved error recovery
- [ ] Enhanced confidence scoring
- [ ] Better progress streaming

## Supported Agents

| Agent | Type | State Directory |
|-------|------|-----------------|
| OpenClaw | Application | ~/.openclaw |
| NanoBot | Application | ~/.nanobot |
| PicoClaw | Application | ~/.picoclaw |
| OpenFang | Application | ~/.openfang |
| ZeroClaw | Application | ~/.zeroclaw |
| NanoClaw | Application | ~/.nanoclaw |
| Cursor | IDE | ~/.cursor |
| Windsurf | IDE | ~/.windsurf |
| Aider | CLI | ~/.aider |
| Cline | Extension | ~/.cline |
| Continue | Extension | ~/.continue |

## Supported Runtimes

| Runtime | Detection | Model Path |
|---------|-----------|------------|
| Ollama | binary + service | ~/.ollama/models |
| LM Studio | directory | ~/.lmstudio/models |
| GPT4All | directory | ~/.cache/gpt4all |
| LocalAI | binary | ~/.localai/models |

## Important Boundaries

### Always Remember

1. **Focus on Agent Environment** - Not general system
2. **Best Effort Detection** - Not exhaustive scanning
3. **Safe Cleanup** - Always backup before delete
4. **Clear Scope** - AI tools only, not enterprise data

### Never Do

1. Don't become a system cleaner
2. Don't become a security scanner
3. Don't scan arbitrary files
4. Don't modify non-AI configurations

## Why This Matters

AI Agents are becoming "local infrastructure" like Docker, Node, Python.

But there's no tool to audit AI agent environments.

ClawRemove fills this gap:

- ✔ Direction is reasonable
- ✔ Technically achievable
- ✔ Won't over-bloat
- ✔ Clear value proposition

## LLM Provider Support

ClawRemove supports multiple LLM providers for AI-powered analysis:

### Supported Provider Types

| Type | API Format | Use Case |
|------|------------|----------|
| `openai` | OpenAI native | OpenAI official API |
| `anthropic` | Anthropic native | Claude official API |
| `openai-compatible` | `/chat/completions` | Third-party OpenAI-compatible services |
| `anthropic-compatible` | `/messages` | Third-party Anthropic-compatible services (e.g., ZhipuAI) |

### Configuration

```bash
# OpenAI
export OPENAI_API_KEY="sk-xxx"
export CLAWREMOVE_LLM_MODEL="GPT-5.4"

# Anthropic
export ANTHROPIC_API_KEY="sk-ant-xxx"
export CLAWREMOVE_LLM_PROVIDER="anthropic"
export CLAWREMOVE_LLM_MODEL="claude-opus-4-6-20250205"

# OpenAI-compatible (e.g., local LLM, custom provider)
export CLAWREMOVE_LLM_PROVIDER="openai-compatible"
export CLAWREMOVE_LLM_BASE_URL="https://your-provider.com/v1"
export CLAWREMOVE_LLM_API_KEY="your-key"

# Anthropic-compatible (e.g., ZhipuAI 智谱AI)
export CLAWREMOVE_LLM_PROVIDER="anthropic-compatible"
export CLAWREMOVE_LLM_BASE_URL="https://open.bigmodel.cn/api/paas/v4"
export CLAWREMOVE_LLM_API_KEY="your-zhipu-key"
export CLAWREMOVE_LLM_MODEL="glm-4-plus"
```

### API Format Details

- **OpenAI-compatible**: Uses `/chat/completions` endpoint with messages array
- **Anthropic-compatible**: Uses `/messages` endpoint with Anthropic-style request format

Both formats are supported for custom SaaS providers, allowing maximum flexibility.
