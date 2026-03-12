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

ClawRemove es un motor profesional y multiplataforma para desinstalar productos claw.

Su objetivo es muy claro: descubrir residuos reales, construir un plan de eliminación, ejecutar la limpieza y verificar el resultado. No intenta comportarse como un limpiador genérico que modifica el sistema sin control.

## Documentación

- English: [README.md](./README.md)
- 中文: [README.zh-CN.md](./README.zh-CN.md)
- Plan de desarrollo: [docs/PLAN.md](./docs/PLAN.md)

## Soporte actual

Proveedor incluido actualmente:

- `openclaw`

La arquitectura ya está preparada para más productos de la familia claw en el futuro, pero por ahora el enfoque está en OpenClaw.

## Estado del proyecto

ClawRemove está en desarrollo activo.

El objetivo inmediato es ofrecer un CLI de nivel profesional para eliminar OpenClaw, mientras el motor se mantiene listo para crecer hacia más proveedores y una futura interfaz de escritorio.

## Principios

- Descubrimiento basado en evidencia
- Auditoría antes de ejecución
- Acciones de alto riesgo siempre con opt-in explícito
- Sin servicios residentes ni base de datos oculta
- Arquitectura por proveedores para ampliar productos más adelante

## Por qué ClawRemove

- Está hecho para desinstalar, no para limpiar el sistema en general
- El comportamiento por defecto es conservador y revisable
- La evidencia pesa más que la heurística
- La salida JSON es útil para automatización y futuras interfaces
- La estructura del repositorio facilita iteración continua con agentes

## Comandos

```bash
claw-remove products
claw-remove audit --product openclaw --json
claw-remove plan --product openclaw --json
claw-remove apply --product openclaw --dry-run
claw-remove apply --product openclaw
claw-remove verify --product openclaw --json
```

### Resumen de comandos

- `products`
  Lista los proveedores compilados
- `audit`
  Auditoría de solo lectura
- `plan`
  Genera un plan sin aplicarlo
- `apply`
  Ejecuta el plan
- `verify`
  Ejecuta una verificación posterior a la eliminación

## Flags

- `--product`
  Id del proveedor. El valor por defecto actual es `openclaw`
- `--json`
  Salida estructurada en JSON
- `--dry-run`
  Muestra los cambios previstos sin aplicarlos
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

## Qué detecta ClawRemove

Según las reglas del proveedor y la plataforma, ClawRemove puede detectar:

- directorios de estado
- workspaces
- rutas temporales y logs
- bundles y datos de aplicación
- launchd, systemd y tareas programadas
- instalaciones con npm, pnpm, bun y Homebrew
- residuos en perfiles del shell
- procesos coincidentes
- puertos en escucha
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
internal/plan              planificación segura
internal/executor          ejecución de acciones
internal/output            salida humana y JSON
internal/products          registro de proveedores
internal/products/openclaw proveedor OpenClaw
internal/system            ejecución de comandos del sistema
docs                       hoja de ruta y plan de desarrollo
scripts                    scripts de build
dist                       artefactos locales
```

## Build

Local:

```bash
go test ./...
go build -o dist/claw-remove ./cmd/claw-remove
```

Multiplataforma:

```bash
./scripts/build.sh
```

PowerShell:

```powershell
./scripts/build.ps1
```

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

Verificar al final:

```bash
claw-remove verify --product openclaw --json
```

## Hoja de ruta

La hoja de ruta detallada vive en [docs/PLAN.md](./docs/PLAN.md).

Ese documento está en inglés a propósito para que contribuidores humanos y agentes puedan compartir la misma fuente de verdad.
