<div align="center">
  <h1>ClawRemove</h1>
  <p><strong>Un motor sobrio, preciso y multiplataforma para eliminar productos claw.</strong></p>
  <p>
    <a href="https://github.com/tianrking/ClawRemove/actions/workflows/ci.yml"><img src="https://github.com/tianrking/ClawRemove/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
    <a href="./LICENSE"><img src="https://img.shields.io/badge/license-MIT-1f6feb" alt="MIT License"></a>
    <img src="https://img.shields.io/badge/go-1.25%2B-00ADD8?logo=go" alt="Go 1.25+">
    <img src="https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-111827" alt="Platform support">
    <a href="https://github.com/tianrking/ClawRemove/releases"><img src="https://img.shields.io/github/v/release/tianrking/ClawRemove" alt="Latest release"></a>
  </p>
  <p><a href="./README.md">English</a> | <a href="./README.zh-CN.md">中文</a> | Español</p>
</div>

ClawRemove es un motor profesional y multiplataforma, escrito en Go, para eliminar productos claw.

Su objetivo es muy claro: descubrir residuos reales de OpenClaw y otros agentes de la familia claw, construir un plan de eliminación, ejecutar la limpieza y verificar el resultado. No intenta comportarse como un limpiador genérico que modifica el sistema sin control.

## Documentación

- English: [README.md](./README.md)
- 中文: [README.zh-CN.md](./README.zh-CN.md)
- Plan de desarrollo: [docs/PLAN.md](./docs/PLAN.md)
- Arquitectura: [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md)

## Soporte actual

Proveedor incluido actualmente:

- `openclaw`
- `nanobot`
- `picoclaw`
- `openfang`
- `zeroclaw`
- `nanoclaw`
- `cursor`
- `windsurf`
- `aider`

La arquitectura ya está preparada para más productos de la familia claw en el futuro.

## Estado del proyecto

ClawRemove está en desarrollo activo.

El objetivo inmediato es ofrecer un CLI de nivel profesional para eliminar OpenClaw, mientras el motor se mantiene listo para crecer hacia más proveedores y una futura interfaz de escritorio.

Ahora cada provider no solo describe hechos y reglas, sino tambien skills y tools de solo lectura para evolucionar el analisis de forma controlada.

## Principios

- Descubrimiento basado en evidencia
- Auditoría antes de ejecución
- Acciones de alto riesgo siempre con opt-in explícito
- Sin servicios residentes ni base de datos oculta
- Arquitectura por proveedores para ampliar productos más adelante
- Cada provider puede declarar sus propios skills y tools de solo lectura
- Si en el futuro integra LLM, ese modelo solo podrá asesorar; no podrá ejecutar acciones destructivas por su cuenta

## Por qué ClawRemove

- Está hecho para desinstalar, no para limpiar el sistema en general
- El comportamiento por defecto es conservador y revisable
- La evidencia pesa más que la heurística
- La salida JSON es útil para automatización y futuras interfaces
- La estructura del repositorio facilita iteración continua con agentes

## Análisis con IA (Opcional)

ClawRemove puede usar IA para explicar hallazgos en lenguaje sencillo. Esto es **opcional** - todas las funciones principales funcionan sin IA.

### Funciones Principales (Sin LLM)

| Comando | Descripción |
|---------|-------------|
| `claw-remove environment` | Inspección completa del entorno |
| `claw-remove inventory` | Inventario de runtime y agentes |
| `claw-remove security` | Auditoría de exposición de API keys |
| `claw-remove hygiene` | Análisis de uso de almacenamiento |
| `claw-remove audit --product X` | Descubrir residuos |
| `claw-remove plan --product X` | Generar plan de eliminación |
| `claw-remove apply --product X` | Ejecutar limpieza |
| `claw-remove verify --product X` | Verificar resultados |

### Funciones con IA (Requiere LLM)

| Comando | Descripción |
|---------|-------------|
| `claw-remove explain --ai` | Explicar hallazgos en lenguaje sencillo |
| `claw-remove audit --ai` | Auditoría + explicación IA |
| `claw-remove verify --ai` | Verificación + explicación IA |

### Lo que la IA Puede y No Puede Hacer

**La IA Puede:**
- ✅ Explicar lo que se descubrió en términos simples
- ✅ Sugerir qué revisar o limpiar
- ✅ Ayudar a clasificar items inciertos

**La IA No Puede:**
- ❌ Ejecutar comandos destructivos
- ❌ Saltar verificaciones de seguridad
- ❌ Modificar tu sistema

### Configuración Rápida

```bash
# Auditoría mejorada con análisis IA
claw-remove audit --ai

# Explicar hallazgos con asistencia IA
claw-remove explain --ai
```

## Comandos

### Inspección de Entorno

```bash
# Informe completo del entorno
claw-remove environment

# Inventario de IA
claw-remove inventory

# Auditoría de seguridad
claw-remove security

# Análisis de almacenamiento
claw-remove hygiene

# Salida JSON
claw-remove environment --json
claw-remove security --json
```

### Limpieza de Agent

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

### Resumen de comandos

**Comandos de Entorno:**
- `environment`
  Informe completo de inspección del entorno (runtime, agentes, artefactos, seguridad, almacenamiento).
- `inventory`
  Inventario de runtime y agentes de IA.
- `security`
  Auditoría de seguridad de herramientas IA (exposición de API keys).
- `hygiene`
  Análisis de uso de almacenamiento de IA.

**Comandos de Limpieza:**
- `products`
  Lista los proveedores compilados
- `audit`
  Auditoría de solo lectura
- `plan`
  Genera un plan sin aplicarlo
- `apply`
  Ejecuta el plan despues de una confirmacion interactiva de seguridad
- `verify`
  Ejecuta una verificación posterior a la eliminación
- `explain`
  Produce analisis asesorado y controlado sobre el descubrimiento determinista

## Flags

- `--product`
  Id del proveedor. El valor por defecto actual es `openclaw`
- `--json`
  Salida estructurada en JSON
- `--ai`
  Incluye analisis asesorado y controlado en el reporte
- `--dry-run`
  Muestra los cambios previstos sin aplicarlos
- `--yes`
  Omite la confirmacion interactiva solo despues de revisar el plan
- `--keep-cli`
  Conserva la desinstalación del CLI y los wrappers
- `--keep-app`
  Conserva la aplicación de escritorio y sus datos
- `--keep-workspace`
  Conserva los workspaces
- `--keep-shell`
  Conserva la limpieza de perfiles y completion del shell
- `--kill-processes`
  Permite terminar procesos coincidentes
- `--remove-docker`
  Permite eliminar contenedores e imágenes de Docker o Podman

## Configuracion LLM

ClawRemove puede adjuntar un advisor controlado a `audit`, `verify` y `explain`.

Si no hay configuracion LLM, ClawRemove vuelve automaticamente al modo determinista.

Proveedores soportados:

- `openai`
- `anthropic`
- `openai-compatible`

Variables de entorno:

- `CLAWREMOVE_LLM_PROVIDER`
  Uno de `openai`, `anthropic` u `openai-compatible`
- `CLAWREMOVE_LLM_API_KEY`
  API key generica
- `OPENAI_API_KEY`
  Fallback key cuando el provider es `openai`
- `ANTHROPIC_API_KEY`
  Fallback key cuando el provider es `anthropic`
- `CLAWREMOVE_LLM_BASE_URL`
  Override del base URL
- `CLAWREMOVE_LLM_MODEL`
  Override del modelo
- `CLAWREMOVE_LLM_MAX_TOKENS`
  Limite de tokens para respuestas asesoradas
- `CLAWREMOVE_LLM_MAX_STEPS`
  Maximo de pasos ReAct controlados
- `CLAWREMOVE_LLM_TIMEOUT_SECONDS`
  Timeout en segundos

Ejemplo con OpenAI:

```bash
export CLAWREMOVE_LLM_PROVIDER="openai"
export OPENAI_API_KEY="..."
export CLAWREMOVE_LLM_MODEL="gpt-4.1-mini"
claw-remove explain --product openclaw --ai --json
```

Ejemplo con Anthropic:

```bash
export CLAWREMOVE_LLM_PROVIDER="anthropic"
export ANTHROPIC_API_KEY="..."
export CLAWREMOVE_LLM_MODEL="claude-3-5-sonnet-latest"
claw-remove explain --product openclaw --ai --json
```

Ejemplo con otro proveedor OpenAI-compatible:

```bash
export CLAWREMOVE_LLM_PROVIDER="openai-compatible"
export CLAWREMOVE_LLM_BASE_URL="https://your-provider.example/v1"
export CLAWREMOVE_LLM_API_KEY="..."
export CLAWREMOVE_LLM_MODEL="your-model-name"
claw-remove explain --product openclaw --ai --json
```

## Flujo de eliminacion segura

Flujo recomendado:

1. `audit`
   Ver lo que encontro ClawRemove
2. `verify`
   Separar residuos confirmados de residuos para investigar
3. `explain --ai`
   Pedir al advisor controlado un resumen
4. `apply`
   Revisar el preview y confirmar con una frase interactiva
5. `apply --yes`
   Usar solo despues de la revision previa o en automatizacion controlada

Por defecto, `apply` no es silencioso ni totalmente automatico.

Primero muestra el preview y despues exige una frase de confirmacion antes de empezar a eliminar.

Si una automatizacion necesita JSON, conviene revisar primero con `plan` o `verify` y despues usar `apply --yes`.

## Qué detecta ClawRemove

Según las reglas del proveedor y la plataforma, ClawRemove puede detectar:

- directorios de estado
- workspaces (subdirectorios declarados por el proveedor)
- rutas temporales y logs
- bundles y datos de aplicación
- launchd, systemd y tareas programadas
- instalaciones con npm, pnpm, bun y Homebrew
- residuos en perfiles del shell (verificados por contenido real, no solo por ruta)
- procesos coincidentes
- puertos en escucha (declarados por el proveedor, sin valores codificados)
- referencias en crontab
- contenedores e imágenes Docker o Podman

## Modelo de seguridad

ClawRemove separa las acciones por nivel de riesgo:

- Bajo riesgo
  Rutas claramente propiedad del producto
- Riesgo medio
  Desactivación de servicios y desinstalación por gestor de paquetes
- Alto riesgo
  Terminación de procesos y eliminación de contenedores o imágenes

Las acciones de alto riesgo requieren opt-in explícito.

## Arquitectura

```text
cmd/claw-remove            entrada CLI
internal/app               orquestación CLI
internal/core              motor principal
internal/discovery         capa de descubrimiento
internal/evidence          capa de recolección de evidencia
internal/plan              planificación segura
internal/executor          ejecución de acciones
internal/llm               coordinación del advisor
internal/llm/prompts       plantillas de prompts
internal/llm/providers     clientes para múltiples modelos
internal/output            salida humana y JSON
internal/platform          adaptadores multiplataforma
internal/products          registro de proveedores
internal/products/openclaw proveedor OpenClaw
internal/skills            habilidades específicas de la plataforma
internal/tools             herramientas específicas de la plataforma
internal/model             modelos unificados
internal/system            ejecución de comandos del sistema
internal/system            ejecución de comandos del sistema
docs                       hoja de ruta y plan de desarrollo
scripts                    scripts de build
dist                       artefactos locales
```

## Instalación

### Binarios Precompilados

Descarga desde [GitHub Releases](https://github.com/tianrking/ClawRemove/releases).

**macOS:**
```bash
# DMG (recomendado)
# Descarga claw-remove-VERSION-macOS.dmg, ábrelo y arrastra a Aplicaciones

# O vía archivo comprimido
curl -sL https://github.com/tianrking/ClawRemove/releases/latest/download/claw-remove-VERSION-darwin-arm64.tar.gz | tar xz
sudo mv claw-remove /usr/local/bin/
```

**Linux:**
```bash
# Debian/Ubuntu (paquete deb)
wget https://github.com/tianrking/ClawRemove/releases/latest/download/claw-remove_VERSION_amd64.deb
sudo dpkg -i claw-remove_VERSION_amd64.deb

# RHEL/Fedora (paquete rpm)
wget https://github.com/tianrking/ClawRemove/releases/latest/download/claw-remove-VERSION-1.x86_64.rpm
sudo rpm -i claw-remove-VERSION-1.x86_64.rpm

# Arch Linux
wget https://github.com/tianrking/ClawRemove/releases/latest/download/claw-remove-VERSION-1-x86_64.pkg.tar.zst
sudo pacman -U claw-remove-VERSION-1-x86_64.pkg.tar.zst

# AppImage (universal, sin instalación)
wget https://github.com/tianrking/ClawRemove/releases/latest/download/claw-remove-VERSION-x86_64.AppImage
chmod +x claw-remove-VERSION-x86_64.AppImage
./claw-remove-VERSION-x86_64.AppImage

# O vía archivo comprimido
curl -sL https://github.com/tianrking/ClawRemove/releases/latest/download/claw-remove-VERSION-linux-amd64.tar.gz | tar xz
sudo mv claw-remove /usr/local/bin/
```

**Windows:**
```powershell
# Descarga el ZIP para tu arquitectura
# claw-remove-VERSION-windows-amd64.zip (x64)
# claw-remove-VERSION-windows-arm64.zip (ARM64)
# claw-remove-VERSION-windows-386.zip (32-bit)

# Extrae y añade al PATH
Expand-Archive claw-remove-VERSION-windows-amd64.zip -DestinationPath C:\Tools\claw-remove
$env:PATH += ";C:\Tools\claw-remove"
```

**BSD:**
```bash
# FreeBSD
curl -sL https://github.com/tianrking/ClawRemove/releases/latest/download/claw-remove-VERSION-freebsd-amd64.tar.gz | tar xz
sudo mv claw-remove /usr/local/bin/
```

### Desde Código Fuente

```bash
go install github.com/tianrking/ClawRemove/cmd/claw-remove@latest
```

## Build

### Local

```bash
go test ./...
go build -o dist/claw-remove ./cmd/claw-remove
```

### Usando GoReleaser

```bash
# Instalar GoReleaser
go install github.com/goreleaser/goreleaser/v2@latest

# Build snapshot local
goreleaser build --snapshot --clean

# Release completo (requiere tag)
goreleaser release --clean
```

### Artefactos de Release

Cada release incluye:

| Formato | Plataforma | Descripción |
|---------|------------|-------------|
| `.dmg` | macOS | Instalador de imagen de disco |
| `.deb` | Linux (Debian/Ubuntu) | Paquete APT |
| `.rpm` | Linux (RHEL/Fedora) | Paquete RPM |
| `.pkg.tar.zst` | Linux (Arch) | Paquete Arch Linux |
| `.AppImage` | Linux (universal) | Ejecutable portátil |
| `.zip` | Windows | Archivo comprimido |
| `.tar.gz` | Todas las plataformas | Archivo comprimido |

Plataformas soportadas (22 en total):

**macOS:**
- `darwin-amd64` (Intel)
- `darwin-arm64` (Apple Silicon)

**Linux:**
- `linux-amd64` (x86_64)
- `linux-arm64` (ARM64)
- `linux-386` (32-bit)
- `linux-arm` (ARM v7, Raspberry Pi)
- `linux-riscv64` (RISC-V)
- `linux-ppc64le` (IBM Power)
- `linux-s390x` (IBM Z)
- `linux-mips64` (MIPS64 big-endian)
- `linux-mips64le` (MIPS64 little-endian)

**Windows:**
- `windows-amd64.exe` (x86_64)
- `windows-arm64.exe` (ARM64)
- `windows-386.exe` (32-bit)

**BSD:**
- `freebsd-amd64`, `freebsd-arm64`
- `netbsd-amd64`
- `openbsd-amd64`

## Flujo recomendado

Auditar primero:

```bash
claw-remove audit --product openclaw --json
```

Generar el plan:

```bash
claw-remove plan --product openclaw --json
```

Probar con dry-run:

```bash
claw-remove apply --product openclaw --dry-run
```

Ejecutar de verdad:

```bash
claw-remove apply --product openclaw
```

Ejecutar sin interaccion solo despues de revisar:

```bash
claw-remove apply --product openclaw --yes
```

Verificar al final:

```bash
claw-remove verify --product openclaw --json
```

Pedir una explicacion controlada:

```bash
claw-remove explain --product openclaw --json
```

## Hoja de ruta

La hoja de ruta detallada vive en [docs/PLAN.md](./docs/PLAN.md).

Ese documento está en inglés a propósito para que contribuidores humanos y agentes puedan compartir la misma fuente de verdad.
