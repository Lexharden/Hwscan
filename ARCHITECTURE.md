# HWSCAN - Arquitectura del Sistema

**Desarrollado por: Yafel Garcia**  
📅 2026 | Enterprise Hardware Detection Tool

---

## Visión General

HWSCAN es una herramienta booteable de detección de hardware diseñada para ejecutarse en Alpine Linux. Proporciona detección completa de componentes del sistema con múltiples interfaces de salida.

## Objetivo

Crear una herramienta que:
- Se ejecute antes del sistema operativo
- Detecte hardware completo (CPU, RAM, GPU, Placa Madre)
- Muestre información en consola y web
- Exporte resultados a JSON
- Funcione completamente offline
- Sea un binario estático sin dependencias

## Estructura del Proyecto

```
hwscan/
├── cmd/hwscan/              # Punto de entrada principal
│   └── main.go             # Inicialización y orquestación
│
├── internal/
│   ├── hardware/           # Detección de hardware
│   │   ├── types.go       # Estructuras de datos
│   │   ├── detector.go    # Lógica de detección
│   │   └── formatter.go   # Formato para consola
│   │
│   ├── server/            # Servidor HTTP embebido
│   │   └── server.go     # Endpoints API y archivos estáticos
│   │
│   └── export/           # Exportación de datos
│       └── export.go    # JSON y detección de USB
│
├── web/                  # Interfaz web
│   └── index.html       # SPA con diseño moderno
│
├── build/alpine/        # Integración con Alpine Linux
│   ├── prepare-alpine.sh    # Script de preparación
│   ├── hwscan.start         # Inicio automático
│   └── README.md            # Documentación de integración
│
├── go.mod              # Módulo Go (sin dependencias externas)
├── Makefile           # Automatización de compilación
└── README.md         # Documentación principal

```

## Componentes Principales

### 1. Detector de Hardware (`internal/hardware/`)

**Responsabilidad:** Detectar componentes del sistema leyendo archivos del kernel y ejecutando comandos del sistema.

**Fuentes de información:**
- `/proc/cpuinfo` - Información del CPU
- `/proc/meminfo` - Memoria total
- `/sys/class/dmi/id/` - Información de la placa madre
- `dmidecode` - Detalles de módulos RAM
- `lspci` - Detección de GPUs

**Estructuras principales:**
```go
type HardwareInfo struct {
    CPU         CPUInfo
    Memory      MemoryInfo
    Motherboard MotherboardInfo
    GPU         []GPUInfo
    Timestamp   string
}
```

### 2. Servidor Web (`internal/server/`)

**Responsabilidad:** Servir interfaz web y API REST para acceder a la información de hardware.

**Endpoints:**
- `GET /` - Interfaz web (HTML/CSS/JS)
- `GET /api/hardware` - Información completa en JSON
- `GET /api/health` - Estado del servicio

**Características:**
- Sin dependencias externas (solo `net/http`)
- Timeouts configurados
- CORS habilitado
- Ejecuta en background

### 3. Exportador (`internal/export/`)

**Responsabilidad:** Exportar información a archivos JSON, preferiblemente en dispositivos USB.

**Funcionalidades:**
- Detección automática de USBs montados
- Generación de nombres con timestamp
- Fallback a directorio actual
- Formato JSON legible

### 4. Interfaz Web (`web/`)

**Responsabilidad:** Proporcionar UI moderna y responsive para visualizar hardware.

**Características:**
- Diseño moderno con gradientes
- Responsive (mobile-friendly)
- Carga asíncrona desde API
- Descarga de JSON desde navegador
- Sin frameworks pesados (vanilla JS)

### 5. Integración Alpine (`build/alpine/`)

**Responsabilidad:** Scripts para integrar HWSCAN en una ISO booteable de Alpine Linux.

**Componentes:**
- Script de preparación de archivos
- Servicio OpenRC
- Script de inicio automático
- Documentación detallada

## Flujo de Ejecución

```
1. Inicio del programa
   ├─> Parseo de flags CLI
   ├─> Banner de bienvenida
   │
2. Detección de Hardware
   ├─> Lectura de /proc/cpuinfo
   ├─> Lectura de /proc/meminfo
   ├─> Lectura de /sys/class/dmi/id/
   ├─> Ejecución de dmidecode (si disponible)
   ├─> Ejecución de lspci
   └─> Construcción de HardwareInfo
   │
3. Presentación
   ├─> Formato y muestra en consola (TUI)
   └─> Timestamp de detección
   │
4. Exportación
   ├─> Detección de USBs montados
   ├─> Generación de nombre con timestamp
   └─> Escritura de JSON
   │
5. Servidor Web
   ├─> Iniciar HTTP server en puerto 8080
   ├─> Registrar endpoints
   └─> Ejecutar en goroutine
   │
6. Mantener vivo
   └─> Esperar señal de interrupción (Ctrl+C)
```

## Decisiones de Diseño

### Sin Dependencias Externas
- **Razón:** Binario completamente estático para Alpine Linux
- **Implementación:** Solo biblioteca estándar de Go
- **Ventaja:** Sin problemas de compatibilidad, tamaño pequeño

### CGO Deshabilitado
- **Razón:** Compatibilidad con musl libc de Alpine
- **Configuración:** `CGO_ENABLED=0` en compilación
- **Ventaja:** Portabilidad total entre distribuciones

### Arquitectura Modular
- **Razón:** Facilitar mantenimiento y expansión
- **Patrón:** Separación por responsabilidades
- **Ventaja:** Cada módulo es testeable independientemente

### Servidor Embebido
- **Razón:** No requiere nginx/apache externo
- **Implementación:** `net/http` de Go
- **Ventaja:** Autocontenido, fácil de desplegar

### Detección Robusta
- **Razón:** Hardware varía mucho entre sistemas
- **Implementación:** Múltiples fuentes, fallbacks
- **Ventaja:** Funciona en la mayoría de configuraciones

## Compilación Multi-Arquitectura

El proyecto soporta compilación para:
- `linux/amd64` - Servidores y PCs x86-64
- `linux/arm64` - Raspberry Pi 3/4, servidores ARM
- `linux/arm/v7` - Raspberry Pi 2, dispositivos ARM32

Configuración de compilación:
```makefile
CGO_ENABLED=0
GOOS=linux
GOARCH=amd64|arm64|arm
LDFLAGS="-s -w"  # Strip y reducción de tamaño
```

## Manejo de Errores

Estrategia de errores por componente:

| Componente | Error Crítico | Error No Crítico |
|-----------|---------------|------------------|
| CPU Detection | Termina programa | N/A |
| Memory Detection | Termina programa | N/A |
| Motherboard | Continúa | Log warning |
| GPU Detection | Continúa | Log warning |
| USB Export | Continúa | Exporta local |
| Web Server | Continúa | Log warning |

## Seguridad

- Sin credenciales hardcoded
- Sin escritura en directorios críticos
- Permisos mínimos requeridos
- Sin ejecución de código arbitrario
- Validación de rutas de archivos

## Performance

- Detección completa: < 2 segundos
- Tamaño binario: ~5-8 MB (stripped)
- Memoria RAM: ~10-20 MB en ejecución
- Sin polling continuo (bajo CPU)

## Extensibilidad

El diseño permite agregar fácilmente:
- Nuevos tipos de hardware (NVMe, USB, etc.)
- Formatos de exportación (XML, CSV)
- Protocolos de servidor (gRPC, WebSocket)
- Plugins de detección personalizados

## Testing

Áreas de prueba recomendadas:
- Unit tests para parsing de /proc y /sys
- Integration tests en contenedor Alpine
- Tests de API HTTP
- Tests de exportación JSON

## Despliegue

### Desarrollo
```bash
make build
./bin/hwscan
```

### Producción Alpine
```bash
make build-amd64
cd build/alpine
bash prepare-alpine.sh
# Seguir README_INTEGRATION.md
```

## Roadmap Futuro

Posibles mejoras:
1. Detección de discos y particiones
2. Información de red (NICs, velocidades)
3. Temperaturas y sensores
4. Benchmarks básicos
5. Comparación con especificaciones conocidas
6. Base de datos de hardware compatible
7. Reportes PDF
8. API para integración con sistemas de inventario

## Referencias

- Alpine Linux: https://alpinelinux.org/
- Go Standard Library: https://pkg.go.dev/std
- Linux /proc filesystem: https://www.kernel.org/doc/Documentation/filesystems/proc.txt
- DMI/SMBIOS: https://www.dmtf.org/standards/smbios
