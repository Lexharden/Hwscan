# HWSCAN - Inicio Rápido

**Desarrollado por: Yafel Garcia**  
🏆 Enterprise Hardware Detection & Auditing Tool

## 🚀 Compilar y Ejecutar

```bash
# Ver todos los comandos disponibles
make help

# Compilar para arquitectura actual
make build

# Compilar para todas las arquitecturas
make build-all

# Ejecutar directamente
make run
```

## 📦 Estructura Generada

```
hwscan/
├── cmd/hwscan/main.go         ✓ Punto de entrada principal
├── internal/
│   ├── hardware/              ✓ Detección de hardware
│   │   ├── types.go          ✓ Estructuras de datos
│   │   ├── detector.go       ✓ Lógica de detección
│   │   └── formatter.go      ✓ Formato consola
│   ├── server/                ✓ Servidor web
│   │   └── server.go         ✓ API REST y archivos estáticos
│   └── export/                ✓ Exportación JSON
│       └── export.go         ✓ Detección USB y exportación
├── web/
│   └── index.html            ✓ Interfaz web moderna
├── build/alpine/              ✓ Integración Alpine Linux
│   ├── prepare-alpine.sh     ✓ Script de preparación
│   ├── hwscan.start          ✓ Inicio automático
│   └── README.md             ✓ Guía de integración
├── go.mod                     ✓ Módulo Go
├── Makefile                   ✓ Automatización
├── README.md                  ✓ Documentación principal
└── ARCHITECTURE.md            ✓ Arquitectura detallada
```

## 🎯 Funcionalidades Implementadas

### ✅ Detección de Hardware
- CPU: modelo, velocidad, núcleos, threads
- RAM: capacidad total, módulos individuales con tipo y velocidad
- Placa Madre: fabricante, modelo, versión, BIOS
- GPU: tarjetas gráficas con vendor y modelo

### ✅ Interfaces
- **Consola**: TUI limpia con formato de tablas
- **Web**: Interfaz moderna en `http://localhost:8080`
- **JSON**: Exportación automática con timestamp

### ✅ Características
- Sin dependencias externas
- Binario completamente estático
- Multi-arquitectura (amd64, arm64, armv7)
- Detección automática de USB
- Servidor HTTP embebido
- Modo offline completo

## 📝 Próximos Pasos

### 1. Desarrollo Local

```bash
# Desde Windows (compilación cruzada)
make build-amd64

# El binario estará en: bin/hwscan-linux-amd64
```

### 2. Testing en Linux

Necesitarás un sistema Linux (VM, WSL2, o nativo) para probar:

```bash
# Copiar binario a Linux
scp bin/hwscan-linux-amd64 user@linux-machine:/tmp/hwscan

# En Linux, ejecutar
chmod +x /tmp/hwscan
sudo /tmp/hwscan
```

**Nota:** Se requiere `sudo` para acceder a dmidecode y obtener información completa de RAM y placa madre.

### 3. Integración con Alpine Linux

```bash
# En un sistema Linux:
make build-amd64
cd build/alpine
bash prepare-alpine.sh

# Seguir la documentación generada
cat README_INTEGRATION.md
```

## 🔧 Comandos Útiles

```bash
# Ver ayuda completa
make help

# Compilar solo AMD64
make build-amd64

# Compilar todas las arquitecturas
make build-all

# Limpiar binarios
make clean

# Ver tamaños de binarios
make size

# Ejecutar tests
make test

# Formatear código
make fmt

# Verificar código
make vet
```

## 🌐 Interfaz Web

Cuando ejecutes hwscan, podrás acceder a:

- `http://localhost:8080` - Interfaz web principal
- `http://localhost:8080/api/hardware` - JSON con toda la información
- `http://localhost:8080/api/health` - Estado del servicio

## 📤 Exportación JSON

El programa exporta automáticamente a:
1. **Primera opción**: USB montado (detecta `/media`, `/mnt`, etc.)
2. **Fallback**: Directorio actual

Formato del archivo: `hwscan-20260217-183045.json`

## 🎨 Características de la Interfaz Web

- Diseño moderno con gradientes
- Totalmente responsive (mobile-friendly)
- Tarjetas organizadas por componente
- Botón de descarga JSON
- Actualización en tiempo real
- Sin frameworks pesados (vanilla JS)

## 🐛 Solución de Problemas

### Error: "dmidecode not found"
```bash
# En Alpine Linux
apk add dmidecode

# En Ubuntu/Debian
sudo apt install dmidecode
```

### Error: "lspci not found"
```bash
# En Alpine Linux
apk add pciutils

# En Ubuntu/Debian
sudo apt install pciutils
```

### No detecta módulos de RAM
- Se requiere ejecutar con `sudo` para acceder a dmidecode
- Sin sudo solo mostrará memoria total

### Servidor web no inicia
- Verificar que el puerto 8080 esté libre
- Usar flag `-port` para cambiar: `hwscan -port 9090`

## 📚 Documentación Adicional

- [README.md](README.md) - Documentación completa
- [ARCHITECTURE.md](ARCHITECTURE.md) - Arquitectura del sistema
- [build/alpine/README.md](build/alpine/README.md) - Integración Alpine

## 🎓 Desarrollo

El proyecto está listo para evolucionar. Áreas de expansión:

1. **Hardware adicional**: Discos, USB, red
2. **Formatos**: XML, CSV, PDF
3. **Plugins**: Sistema de plugins para detección customizada
4. **Benchmarks**: Tests de rendimiento
5. **Base de datos**: Comparación con specs conocidas

## 💡 Tips

- El código está completamente documentado en español
- Cada módulo es independiente y testeable
- Sin dependencias = sin problemas de versiones
- Listo para uso empresarial

## 🤝 Contribuir

El proyecto está estructurado profesionalmente:
- Separación clara de responsabilidades
- Código idiomático Go
- Sin dependencias externas
- Listo para CI/CD

---

**¡El proyecto está completo y listo para usar!** 🎉
