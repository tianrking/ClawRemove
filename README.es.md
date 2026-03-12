# ClawRemove

ClawRemove es un motor profesional y multiplataforma para desinstalar productos claw.

Su objetivo es muy claro: descubrir residuos reales, construir un plan de eliminación, ejecutar la limpieza y verificar el resultado. No intenta comportarse como un “limpiador genérico” que modifica el sistema sin control.

## Documentación

- English: [README.md](./README.md)
- 中文: [README.zh-CN.md](./README.zh-CN.md)

## Soporte actual

Proveedor incluido actualmente:

- `openclaw`

La arquitectura ya está preparada para más productos de la familia claw en el futuro, pero por ahora el enfoque está en OpenClaw.

## Principios

- Descubrimiento basado en evidencia
- Auditoría antes de ejecución
- Acciones de alto riesgo siempre con opt-in explícito
- Sin servicios residentes ni base de datos oculta
- Arquitectura por proveedores para ampliar productos más adelante

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
  Lista los proveedores compilados.
- `audit`
  Auditoría de solo lectura.
- `plan`
  Genera un plan sin aplicarlo.
- `apply`
  Ejecuta el plan.
- `verify`
  Ejecuta una verificación posterior a la eliminación.

## Flags

- `--product`
  Id del proveedor. El valor por defecto actual es `openclaw`.
- `--json`
  Salida estructurada en JSON.
- `--dry-run`
  Muestra los cambios previstos sin aplicarlos.
- `--keep-cli`
  Conserva la desinstalación del CLI y los wrappers.
- `--keep-app`
  Conserva la aplicación de escritorio y sus datos.
- `--keep-workspace`
  Conserva los workspaces.
- `--keep-shell`
  Conserva la limpieza de perfiles/completion del shell.
- `--kill-processes`
  Permite terminar procesos coincidentes.
- `--remove-docker`
  Permite eliminar contenedores e imágenes de Docker/Podman.

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
- contenedores e imágenes Docker/Podman

## Modelo de seguridad

ClawRemove separa las acciones por nivel de riesgo:

- Bajo riesgo
  Rutas claramente propiedad del producto.
- Riesgo medio
  Desactivación de servicios y desinstalación por gestor de paquetes.
- Alto riesgo
  Terminación de procesos y eliminación de contenedores/imágenes.

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

## Nota sobre el repositorio

El código real del producto vive dentro de `ClawRemove/`.

La zona externa del workspace puede usarse para:

- análisis con agentes
- investigación de otros productos claw
- repositorios de referencia
- experimentos temporales

Ese material externo se excluye intencionalmente del control de versiones.
