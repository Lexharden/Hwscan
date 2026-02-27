#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_DIR="$SCRIPT_DIR/base"
WORK_DIR="$SCRIPT_DIR/work"
OUTPUT_DIR="$SCRIPT_DIR/output"
ISO_NAME="alpine-standard-3.23.3-x86_64.iso"
ISO_PATH="$BASE_DIR/$ISO_NAME"
OUTPUT_ISO="$OUTPUT_DIR/alpine-hwscan-3.23.3-x86_64.iso"
BINARY_PATH="$SCRIPT_DIR/../../dist/hwscan"
WEB_DIR="$SCRIPT_DIR/../../web"

echo ""
echo "==================================================="
echo "  HWSCAN Alpine ISO Builder v1"
echo "  Alpine Linux 3.23.3 x86_64"
echo "==================================================="
echo ""

# [1/6] Verificar dependencias
echo "[1/6] Verificando dependencias..."
missing_deps=()
for cmd in xorriso cpio gzip find tar; do
    if ! command -v $cmd &> /dev/null; then
        missing_deps+=($cmd)
    fi
done

if [ ${#missing_deps[@]} -ne 0 ]; then
    echo "❌ Error: Faltan las siguientes dependencias: ${missing_deps[*]}"
    exit 1
fi
echo "   ✓ Todas las dependencias encontradas"

# [2/6] Verificar archivos
echo "[2/6] Verificando archivos necesarios..."
if [ ! -f "$ISO_PATH" ]; then
    echo "❌ Error: ISO de Alpine no encontrada en: $ISO_PATH"
    exit 1
fi

if [ ! -f "$BINARY_PATH" ]; then
    echo "❌ Error: binario hwscan no encontrado en: $BINARY_PATH"
    exit 1
fi

if [ ! -d "$WEB_DIR" ]; then
    echo "❌ Error: directorio web no encontrado en: $WEB_DIR"
    exit 1
fi
echo "   ✓ Archivos verificados"

# [3/6] Preparar entorno y extraer ISO base
echo "[3/6] Preparando espacio de trabajo y extrayendo ISO base..."
rm -rf "$WORK_DIR"
mkdir -p "$WORK_DIR/iso"
xorriso -osirrox on -indev "$ISO_PATH" -extract / "$WORK_DIR/iso" 2>&1 | head -1
chmod -R u+w "$WORK_DIR/iso"
echo "   ✓ ISO extraída correctamente"

# [4/6] Construir el sistema de archivos RAM (Overlay)
echo "[4/6] Construyendo overlay del sistema de archivos RAM..."
OVERLAY_DIR="$WORK_DIR/overlay"

# Crear estructura de directorios
mkdir -p "$OVERLAY_DIR/usr/bin"
mkdir -p "$OVERLAY_DIR/usr/share/hwscan/web" # <--- directorio web de la interfaz
mkdir -p "$OVERLAY_DIR/etc/network"
mkdir -p "$OVERLAY_DIR/etc/apk"
mkdir -p "$OVERLAY_DIR/etc/runlevels/sysinit"
mkdir -p "$OVERLAY_DIR/etc/runlevels/boot"
mkdir -p "$OVERLAY_DIR/etc/runlevels/default"
mkdir -p "$OVERLAY_DIR/etc/runlevels/shutdown"
mkdir -p "$OVERLAY_DIR/etc/profile.d"

# 4.1 Inyectar el binario
cp "$BINARY_PATH" "$OVERLAY_DIR/usr/bin/hwscan"
chmod +x "$OVERLAY_DIR/usr/bin/hwscan"

# 4.2 Copiar archivos de la interfaz web
cp -r "$WEB_DIR"/* "$OVERLAY_DIR/usr/share/hwscan/web/"
echo "   ✓ Binario y archivos web copiados"

# 4.3 Configurar nombre de máquina y red
echo "localhost" > "$OVERLAY_DIR/etc/hostname"
cat > "$OVERLAY_DIR/etc/network/interfaces" << 'EOFNET'
auto lo
iface lo inet loopback

auto eth0
iface eth0 inet dhcp
EOFNET

# 4.4 Activar repositorios online de Alpine
cat > "$OVERLAY_DIR/etc/apk/repositories" << 'EOFAPK'
https://dl-cdn.alpinelinux.org/alpine/v3.23/main
https://dl-cdn.alpinelinux.org/alpine/v3.23/community
EOFAPK

# 4.5 Script de auto-instalación de herramientas de hardware al iniciar sesión
cat > "$OVERLAY_DIR/etc/profile.d/install-hw-tools.sh" << 'EOFSH'
#!/bin/sh
if ! command -v lspci >/dev/null 2>&1 || ! command -v dmidecode >/dev/null 2>&1; then
    echo -e "\e[36m[*] Instalando herramientas de hardware (lspci/lsusb/dmidecode)...\e[0m"
    apk update -q >/dev/null 2>&1
    apk add -q pciutils pciutils-libs hwdata-pci usbutils hwdata-usb dmidecode >/dev/null 2>&1
    if command -v lspci >/dev/null 2>&1; then
        echo -e "\e[32m[✓] Herramientas instaladas correctamente.\e[0m"
    else
        echo -e "\e[31m[!] Error: No se pudo instalar las herramientas.\e[0m"
    fi
fi
EOFSH
chmod +x "$OVERLAY_DIR/etc/profile.d/install-hw-tools.sh"

# 4.6 Restaurar los servicios por defecto de Alpine
for svc in devfs dmesg mdev hwdrivers modloop; do ln -sf "/etc/init.d/$svc" "$OVERLAY_DIR/etc/runlevels/sysinit/$svc" 2>/dev/null || true; done
for svc in bootmisc hostname hwclock modules swap sysctl syslog termencoding urandom; do ln -sf "/etc/init.d/$svc" "$OVERLAY_DIR/etc/runlevels/boot/$svc" 2>/dev/null || true; done
for svc in networking acpid cron local; do ln -sf "/etc/init.d/$svc" "$OVERLAY_DIR/etc/runlevels/default/$svc" 2>/dev/null || true; done
for svc in killprocs mount-ro savecache; do ln -sf "/etc/init.d/$svc" "$OVERLAY_DIR/etc/runlevels/shutdown/$svc" 2>/dev/null || true; done

# Empaquetar el overlay
cd "$OVERLAY_DIR"
tar -czf "$WORK_DIR/iso/localhost.apkovl.tar.gz" *
cd "$SCRIPT_DIR"
echo "   ✓ Overlay construido con script de auto-instalación y repositorios"

# [5/6] Parchar boot configs (compatible con Ventoy, QEMU y boot directo)
echo "[5/6] Parcheando configuraciones de arranque..."
syslinux_cfg="$WORK_DIR/iso/boot/syslinux/syslinux.cfg"
grub_cfg="$WORK_DIR/iso/boot/grub/grub.cfg"
# El label DEBE coincidir con -V en el comando xorriso del paso 6
BOOTLABEL="ALPINE_HWSCAN"

if [ -f "$syslinux_cfg" ]; then
    cp "$syslinux_cfg" "$syslinux_cfg.bak"
    # ① Quitar alpine_dev=cdrom:iso9660 — solo busca /dev/sr0, cuelga con Ventoy
    sed -i 's/ alpine_dev=cdrom:iso9660//g' "$syslinux_cfg"
    # ② Inyectar detección por label de volumen: funciona con Ventoy (loop), QEMU y CD físico
    sed -i "s|\(APPEND.*\)|\1 alpine_dev=LABEL:${BOOTLABEL} modloop=/boot/modloop-lts modloop_verify=no console=tty0|" "$syslinux_cfg"
    echo "   ✓ syslinux.cfg parcheado"
fi

if [ -f "$grub_cfg" ]; then
    cp "$grub_cfg" "$grub_cfg.bak"
    # ① Quitar alpine_dev=cdrom:iso9660
    sed -i 's/ alpine_dev=cdrom:iso9660//g' "$grub_cfg"
    # ② Inyectar detección por label (líneas que comienzan con 'linux /boot')
    sed -i "s|\(linux /boot.*\)|\1 alpine_dev=LABEL:${BOOTLABEL} modloop=/boot/modloop-lts modloop_verify=no console=tty0|" "$grub_cfg"
    echo "   ✓ grub.cfg parcheado"
fi
echo "   ✓ Parámetros de arranque actualizados (compatible con Ventoy)"

# [6/6] Reconstruir ISO
echo "[6/6] Reconstruyendo ISO..."
mkdir -p "$OUTPUT_DIR"
rm -f "$OUTPUT_ISO"

mbr_bin="$WORK_DIR/iso/boot/syslinux/isohdpfx.bin"
if [ ! -f "$mbr_bin" ]; then
    # Extraer el sector MBR directamente desde la ISO base
    dd if="$ISO_PATH" bs=432 count=1 of="$WORK_DIR/isohdpfx.bin" 2>/dev/null
    mbr_bin="$WORK_DIR/isohdpfx.bin"
fi

xorriso -as mkisofs \
    -o "$OUTPUT_ISO" \
    -isohybrid-mbr "$mbr_bin" \
    -c boot/syslinux/boot.cat \
    -b boot/syslinux/isolinux.bin \
    -no-emul-boot \
    -boot-load-size 4 \
    -boot-info-table \
    -eltorito-alt-boot \
    -e boot/grub/efi.img \
    -no-emul-boot \
    -isohybrid-gpt-basdat \
    -V "ALPINE_HWSCAN" \
    -joliet -joliet-long -rational-rock \
    "$WORK_DIR/iso" 2>&1 | grep -i "writing" | head -1

# Limpiar directorio de trabajo temporal
rm -rf "$WORK_DIR"

ISO_SIZE=$(du -h "$OUTPUT_ISO" | cut -f1)

echo ""
echo "==================================================="
echo "  ✓ Construcción completada exitosamente"
echo "==================================================="
echo "Salida: $OUTPUT_ISO"
echo "Tamaño: $ISO_SIZE"
echo ""
echo "Prueba en QEMU:"
echo "  qemu-system-x86_64 -cdrom \"$OUTPUT_ISO\" -m 2048 -nic user,model=e1000"
echo ""