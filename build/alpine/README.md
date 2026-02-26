# Integración con Alpine Linux

Este directorio contiene los scripts y archivos necesarios para integrar HWSCAN en una ISO booteable de Alpine Linux.

## Archivos

- `prepare-alpine.sh` - Script principal de preparación (ejecutar desde Linux)
- `hwscan.start` - Script de inicio automático para Alpine Linux
- `README_INTEGRATION.md` - Documentación detallada (generada por prepare-alpine.sh)
- `create-alpine-iso.sh` - Script automatizado para crear ISO (generado por prepare-alpine.sh)
- `rootfs/` - Estructura de archivos para integrar (generada por prepare-alpine.sh)

## Uso Rápido

Desde el directorio raíz del proyecto:

```bash
# 1. Compilar el binario para Linux AMD64
make build-amd64

# 2. Preparar integración Alpine
cd build/alpine
bash prepare-alpine.sh

# 3. Seguir instrucciones del README_INTEGRATION.md generado
```

## Notas

- Los scripts de creación de ISO deben ejecutarse en un sistema Linux
- Se requieren permisos de root para montar ISOs y crear nuevas imágenes
- El proceso requiere herramientas: `mkisofs`, `mount`, `wget`
