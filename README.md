# HWSCAN

**Herramienta de auditoría de hardware — booteable**

Desarrollado por Yafel Garcia · Go 1.21 · Solo stdlib · v1.0.0

---

## Descripción

HWSCAN detecta y reporta el hardware de cualquier máquina con solo arrancar desde una ISO de Alpine Linux. Presenta la información en consola, la sirve mediante una interfaz web minimalista y la exporta automáticamente a un JSON en USB.

Requiere internet para instalar dependencias de GPU.

## Características

- Detección completa: CPU, RAM (módulos individuales), Placa Madre (BIOS incluido), GPU
- Velocidad del CPU leída desde `/sys/devices/.../cpufreq/cpuinfo_max_freq` (frecuencia máxima real, no idle)
- Consola formateada con datos al vuelo
- Servidor HTTP embebido en el puerto 8080 con dashboard web oscuro y responsive
- Exportación automática a JSON: detecta USB montado, si no hay exporta en el directorio actual
- Identificador único de máquina (`machine_id`)
- Binario 100% estático (`CGO_ENABLED=0`), sin dependencias externas
- Multi-arquitectura: `linux/amd64`, `linux/arm64`, `linux/armv7`

## Compilación

```bash
# Binario para arquitectura actual
make build

# Binario estático para Linux AMD64 (recomendado para Alpine ISO)
make build-amd64

# Otros targets
make build-arm64
make build-armv7
make build-all

# Limpiar artefactos
make clean
```

Los binarios se generan en `bin/`. El target `build-amd64` también copia el binario a `dist/hwscan`, que es la ruta que usa el script de construcción de la ISO.

## Uso

```bash
# Ejecución por defecto: detecta hardware, exporta JSON, levanta servidor web en :8080
./hwscan

# Puerto personalizado
./hwscan -port 9090

# Solo consola, sin servidor ni exportación
./hwscan -no-server -no-export

# Exportar a ruta específica
./hwscan -output /ruta/mi-reporte.json

# Sin exportación automática
./hwscan -no-export

# Info de versión
./hwscan -version
```

### Flags disponibles

| Flag | Default | Descripción |
|------|---------|-------------|
| `-port` | `8080` | Puerto del servidor web |
| `-no-server` | `false` | Deshabilita el servidor HTTP |
| `-no-export` | `false` | Deshabilita la exportación a JSON |
| `-output` | `""` | Ruta de salida específica para el JSON |
| `-version` | — | Muestra la versión y sale |
| `-help` | — | Muestra la ayuda y sale |

## API REST

Cuando el servidor está activo:

| Endpoint | Descripción |
|----------|-------------|
| `GET /api/hardware` | JSON completo con toda la info de hardware |
| `GET /api/health` | Estado del servidor (`{"status":"ok"}`) |
| `GET /` | Dashboard web |

### Ejemplo de respuesta `/api/hardware`

```json
{
  "machine_id": "abc123...",
  "cpu": {
    "model": "Intel(R) Core(TM) i7-12700",
    "vendor": "GenuineIntel",
    "cores": 12,
    "threads": 20,
    "speed_mhz": 4900.0,
    "cache_size": "25600 KB"
  },
  "memory": {
    "total_gb": 32.0,
    "modules": [
      { "size": "16GB", "type": "DDR4", "speed": "3200MHz", "locator": "DIMM1" }
    ]
  },
  "motherboard": {
    "manufacturer": "ASUSTeK",
    "product": "PRIME B660M-A",
    "bios_vendor": "American Megatrends",
    "bios_version": "1401",
    "bios_date": "11/14/2022"
  },
  "gpu": [
    { "vendor": "Intel", "model": "UHD Graphics 770", "pci_address": "0000:00:02.0" }
  ],
  "timestamp": "2026-02-26T10:30:00Z"
}
```

## Estructura del Proyecto

```
hwscan/
├── cmd/hwscan/             # Punto de entrada (main.go)
│   └── main.go             # Flags, orquestación, servidor, shutdown
├── internal/
│   ├── hardware/
│   │   ├── detector.go     # Lectura de /proc/cpuinfo, dmidecode paths, cpufreq, PCI
│   │   ├── formatter.go    # Salida formateada a consola
│   │   ├── machineid.go    # Identificador único de la máquina
│   │   └── types.go        # Structs: HardwareInfo, CPUInfo, MemoryInfo, etc.
│   ├── server/
│   │   └── server.go       # HTTP server: /api/hardware, /api/health, static web
│   ├── export/
│   │   └── export.go       # ExportToJSON, AutoExport (USB detection)
│   └── utils/
│       └── utils.go        # GetLocalIP()
├── web/
│   └── index.html          # Dashboard web (dark theme, responsive, vanilla JS)
├── build/
│   └── alpine/
│       ├── build.sh        # Genera alpine-hwscan-3.23.3-x86_64.iso
│       ├── verify.sh       # Verifica la ISO generada
│       ├── README.md       # Instrucciones del builder
│       └── base/           # ISO base de Alpine (no incluida en repo)
├── go.mod
├── Makefile
└── LICENSE
```

## Integración Alpine Linux

El script `build/alpine/build.sh` genera una ISO booteable de Alpine Linux con HWSCAN integrado:

```bash
# 1. Compilar el binario estático
make build-amd64
# → genera dist/hwscan

# 2. Colocar la ISO base de Alpine en build/alpine/base/
#    alpine-standard-3.23.3-x86_64.iso

# 3. Generar la ISO personalizada (requiere Linux + xorriso)
cd build/alpine
bash build.sh
# → genera build/alpine/output/alpine-hwscan-3.23.3-x86_64.iso
```

El proceso embebe el binario en `/usr/local/bin/hwscan`, los archivos web en `/usr/share/hwscan/web`, configura DHCP automático y lanza HWSCAN al iniciar sesión.

## Requisitos para compilar

- Go 1.21 o superior
- No requiere ninguna dependencia externa (solo stdlib)

## Requisitos para construir la ISO

- Linux (cualquier distro)
- `xorriso`, `cpio`, `gzip`, `tar`
- ISO base: `alpine-standard-3.23.3-x86_64.iso` (descargar desde [alpinelinux.org](https://alpinelinux.org/downloads/))

## Autor

**Yafel Garcia** — [github.com/Lexharden](https://github.com/Lexharden)

## Licencia

MIT License — Copyright (c) 2026 Yafel Garcia
