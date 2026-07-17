# HWSCAN

**Herramienta de auditoría de hardware — booteable**

Desarrollado por Yafel Garcia · Go 1.21 · Solo stdlib · v1.1.2

---

## Descripción

HWSCAN detecta y reporta el hardware de cualquier máquina al arrancar desde una ISO de Alpine Linux. Muestra la información en consola, en una interfaz web y permite exportarla a JSON o CSV.

La ISO arranca en modo live, hace auto-login como `root` y ejecuta HWSCAN sin intervención. Requiere internet en el primer arranque para instalar `lspci` y `dmidecode` (detección completa de GPU y módulos RAM).

## Características

- Detección: CPU, RAM (módulos), placa madre (BIOS), GPU, discos
- Identificador único de máquina (`machine_id`)
- Consola formateada y servidor HTTP embebido en `:8080`
- Interfaz web con tema claro/oscuro y exportación JSON/CSV
- Exportación automática a JSON (USB o directorio actual)
- Binario 100% estático (`CGO_ENABLED=0`), sin dependencias Go externas
- ISO híbrida BIOS + UEFI: `hwscan-live-x86_64.iso`
- Multi-arquitectura del binario: `linux/amd64`, `linux/arm64`, `linux/armv7`

## Compilación

```bash
make build-amd64    # recomendado para la ISO → dist/hwscan
make build          # arquitectura actual
make build-all      # amd64 + arm64 + armv7
make clean
```

## Uso local

```bash
./bin/hwscan                        # consola + JSON + web :8080
./bin/hwscan -no-server -no-export  # solo consola
./bin/hwscan -port 9090
./bin/hwscan -output /tmp/reporte.json
./bin/hwscan -version
```

| Flag | Default | Descripción |
|------|---------|-------------|
| `-port` | `8080` | Puerto del servidor web |
| `-no-server` | `false` | Desactiva el servidor HTTP |
| `-no-export` | `false` | Desactiva la exportación automática a JSON |
| `-output` | `""` | Ruta específica del JSON |
| `-version` | — | Muestra la versión |
| `-help` | — | Ayuda |

## API REST

| Endpoint | Descripción |
|----------|-------------|
| `GET /api/hardware` | Hardware detectado (JSON) |
| `GET /api/health` | Estado del servicio |
| `GET /` | Dashboard web |

## Construir la ISO booteable

### 1. Compilar

```bash
make build-amd64
```

### 2. Colocar la ISO base de Alpine

Descarga una ISO **standard** desde [alpinelinux.org](https://alpinelinux.org/downloads/) y colócala en:

```text
build/alpine/base/
```

Ejemplo: `alpine-standard-3.24.1-x86_64.iso`

> Esta carpeta y las ISOs **no se versionan** (ver `.gitignore`).

### 3. Generar la ISO de HWSCAN

```bash
cd build/alpine
bash build.sh          # detecta ISOs en base/ y pregunta si hay varias
bash verify.sh         # comprueba boot, apkovl y parches UEFI/BIOS
```

Salida:

```text
build/alpine/output/hwscan-live-x86_64.iso
```

Formas no interactivas:

```bash
bash build.sh alpine-standard-3.24.1-x86_64.iso
ALPINE_ISO=base/mi-alpine.iso bash build.sh
```

### 4. Grabar en USB (recomendado: `dd`)

```bash
lsblk
sudo umount /dev/sdX* 2>/dev/null || true
sudo dd if=build/alpine/output/hwscan-live-x86_64.iso of=/dev/sdX bs=4M status=progress conv=fsync
sync
```

Sustituye `/dev/sdX` por tu USB real (nunca `sda` del disco del sistema).

En la PC objetivo (UEFI): desactiva **Secure Boot** y elige **UEFI: USB** en el menú de arranque.

Herramientas alternativas: Balena Etcher, Rufus (modo DD), Ventoy, GNOME Disks.

## Estructura del proyecto

```text
hwscan/
├── cmd/hwscan/           # Punto de entrada
├── internal/
│   ├── hardware/         # Detección y tipos
│   ├── server/           # API HTTP + web estática
│   ├── export/           # JSON y detección USB
│   └── version/          # Versión (fuente de verdad)
├── web/index.html        # Dashboard (tema claro/oscuro)
├── build/alpine/
│   ├── build.sh          # Genera hwscan-live-x86_64.iso
│   ├── verify.sh         # Verifica la ISO
│   ├── base/             # ISO Alpine de entrada (local, ignorada por git)
│   ├── work/             # Temporal de build (ignorado)
│   └── output/           # ISO generada (ignorada)
├── Makefile
└── go.mod
```

## Qué no subir a Git

El `.gitignore` excluye automáticamente:

| Ruta / patrón | Contenido |
|---------------|-----------|
| `build/alpine/base/` | ISOs Alpine descargadas |
| `build/alpine/output/` | `hwscan-live-x86_64.iso` generada |
| `build/alpine/work/` | Archivos temporales del builder |
| `*.iso` | Cualquier imagen ISO en el repo |
| `bin/`, `dist/` | Binarios compilados |
| `hwscan-*.json` | Reportes exportados localmente |

Los scripts `build.sh` y `verify.sh` **sí** van en el repositorio.

## Requisitos

**Compilar HWSCAN:** Go 1.21+

**Construir la ISO:** Linux, `xorriso`, `cpio`, `gzip`, `tar`

## Autor

**Yafel Garcia** — [github.com/Lexharden](https://github.com/Lexharden)

## Licencia

MIT License — Copyright (c) 2026 Yafel Garcia
