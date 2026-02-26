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
echo "[1/6] Checking dependencies..."
missing_deps=()
for cmd in xorriso cpio gzip find tar; do
    if ! command -v $cmd &> /dev/null; then
        missing_deps+=($cmd)
    fi
done

if [ ${#missing_deps[@]} -ne 0 ]; then
    echo "❌ Error: Missing dependencies: ${missing_deps[*]}"
    exit 1
fi
echo "   ✓ All dependencies found"

# [2/6] Verificar archivos
echo "[2/6] Checking files..."
if [ ! -f "$ISO_PATH" ]; then
    echo "❌ Error: Alpine ISO not found at: $ISO_PATH"
    exit 1
fi

if [ ! -f "$BINARY_PATH" ]; then
    echo "❌ Error: hwscan binary not found at: $BINARY_PATH"
    exit 1
fi

if [ ! -d "$WEB_DIR" ]; then
    echo "❌ Error: web directory not found at: $WEB_DIR"
    exit 1
fi
echo "   ✓ Files verified"

# [3/6] Preparar entorno y extraer ISO base
echo "[3/6] Preparing workspace and extracting ISO..."
rm -rf "$WORK_DIR"
mkdir -p "$WORK_DIR/iso"
xorriso -osirrox on -indev "$ISO_PATH" -extract / "$WORK_DIR/iso" 2>&1 | head -1
chmod -R u+w "$WORK_DIR/iso"
echo "   ✓ ISO extracted"

# [4/6] Construir el sistema de archivos RAM (Overlay)
echo "[4/6] Building RAM filesystem overlay..."
OVERLAY_DIR="$WORK_DIR/overlay"

# Crear estructura
mkdir -p "$OVERLAY_DIR/usr/bin"
mkdir -p "$OVERLAY_DIR/usr/share/hwscan/web" # <--- CREAMOS LA CARPETA WEB AQUÍ
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

# 4.1.2 Copiar directorio web
cp -r "$WEB_DIR"/* "$OVERLAY_DIR/usr/share/hwscan/web/"
echo "   ✓ Binary and web files copied"

# 4.2 Configurar nombre de máquina y red
echo "localhost" > "$OVERLAY_DIR/etc/hostname"
cat > "$OVERLAY_DIR/etc/network/interfaces" << 'EOFNET'
auto lo
iface lo inet loopback

auto eth0
iface eth0 inet dhcp
EOFNET

# 4.3 Activar los repositorios online de Alpine
cat > "$OVERLAY_DIR/etc/apk/repositories" << 'EOFAPK'
https://dl-cdn.alpinelinux.org/alpine/v3.23/main
https://dl-cdn.alpinelinux.org/alpine/v3.23/community
EOFAPK

# 4.4 Script de auto-instalación de herramientas de hardware al hacer login
cat > "$OVERLAY_DIR/etc/profile.d/install-hw-tools.sh" << 'EOFSH'
#!/bin/sh
if ! command -v lspci >/dev/null 2>&1; then
    echo -e "\e[36m[*] Instalando herramientas de hardware (lspci/lsusb)...\e[0m"
    apk update -q >/dev/null 2>&1
    apk add -q pciutils pciutils-libs hwdata-pci usbutils hwdata-usb >/dev/null 2>&1
    if command -v lspci >/dev/null 2>&1; then
        echo -e "\e[32m[✓] Herramientas instaladas correctamente.\e[0m"
    else
        echo -e "\e[31m[!] Error: No se pudo instalar lspci.\e[0m"
    fi
fi
EOFSH
chmod +x "$OVERLAY_DIR/etc/profile.d/install-hw-tools.sh"

# 4.5 Restaurar los servicios por defecto de Alpine
for svc in devfs dmesg mdev hwdrivers modloop; do ln -sf "/etc/init.d/$svc" "$OVERLAY_DIR/etc/runlevels/sysinit/$svc" 2>/dev/null || true; done
for svc in bootmisc hostname hwclock modules swap sysctl syslog termencoding urandom; do ln -sf "/etc/init.d/$svc" "$OVERLAY_DIR/etc/runlevels/boot/$svc" 2>/dev/null || true; done
for svc in networking acpid cron local; do ln -sf "/etc/init.d/$svc" "$OVERLAY_DIR/etc/runlevels/default/$svc" 2>/dev/null || true; done
for svc in killprocs mount-ro savecache; do ln -sf "/etc/init.d/$svc" "$OVERLAY_DIR/etc/runlevels/shutdown/$svc" 2>/dev/null || true; done

# Empaquetar
cd "$OVERLAY_DIR"
tar -czf "$WORK_DIR/iso/localhost.apkovl.tar.gz" *
cd "$SCRIPT_DIR"
echo "   ✓ Overlay built with auto-install script and repositories"

# [5/6] Parchar boot configs
echo "[5/6] Patching boot configs..."
syslinux_cfg="$WORK_DIR/iso/boot/syslinux/syslinux.cfg"
grub_cfg="$WORK_DIR/iso/boot/grub/grub.cfg"

if [ -f "$syslinux_cfg" ]; then
    cp "$syslinux_cfg" "$syslinux_cfg.bak"
    sed -i 's/\(APPEND.*\)/\1 modloop=\/boot\/modloop-lts modloop_verify=no console=tty0/' "$syslinux_cfg"
fi

if [ -f "$grub_cfg" ]; then
    cp "$grub_cfg" "$grub_cfg.bak"
    sed -i 's/\(linux .*\)/\1 modloop=\/boot\/modloop-lts modloop_verify=no console=tty0/' "$grub_cfg"
fi
echo "   ✓ Boot parameters updated"

# [6/6] Reconstruir ISO
echo "[6/6] Rebuilding ISO..."
mkdir -p "$OUTPUT_DIR"
rm -f "$OUTPUT_ISO"

mbr_bin="$WORK_DIR/iso/boot/syslinux/isohdpfx.bin"
if [ ! -f "$mbr_bin" ]; then
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

# Cleanup
rm -rf "$WORK_DIR"

ISO_SIZE=$(du -h "$OUTPUT_ISO" | cut -f1)

echo ""
echo "==================================================="
echo "  ✓ Build complete!"
echo "==================================================="
echo "Output: $OUTPUT_ISO"
echo "Size:   $ISO_SIZE"
echo ""
echo "Prueba en QEMU:"
echo "  qemu-system-x86_64 -cdrom \"$OUTPUT_ISO\" -m 2048 -nic user,model=e1000"
echo ""