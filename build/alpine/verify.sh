#!/bin/bash
# Verification script - checks if build was successful and patches applied

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORK_DIR="$SCRIPT_DIR/work"
OUTPUT_DIR="$SCRIPT_DIR/output"
OUTPUT_ISO="$OUTPUT_DIR/alpine-hwscan-3.23.3-x86_64.iso"

echo "=================================================="
echo "  HWSCAN Build Verification Script"
echo "=================================================="
echo ""

ERRORS=0
WARNINGS=0

# Check 1: Work directory should be cleaned
echo -n "[1/8] Checking work directory cleanup... "
if [ -d "$WORK_DIR" ]; then
    echo "⚠️  WARNING"
    echo "      Work directory still exists (should be cleaned)"
    WARNINGS=$((WARNINGS + 1))
else
    echo "✓ OK"
fi

# Check 2: Output ISO exists
echo -n "[2/8] Checking output ISO exists... "
if [ ! -f "$OUTPUT_ISO" ]; then
    echo "❌ FAIL"
    echo "      ISO not found at: $OUTPUT_ISO"
    ERRORS=$((ERRORS + 1))
else
    echo "✓ OK"
    ISO_SIZE=$(du -h "$OUTPUT_ISO" | cut -f1)
    echo "      Size: $ISO_SIZE"
fi

# Check 3: ISO is bootable (has MBR)
echo -n "[3/8] Checking ISO bootability (MBR)... "
if [ -f "$OUTPUT_ISO" ]; then
    # Check for MBR signature (0x55AA at offset 510)
    MBR_SIG=$(dd if="$OUTPUT_ISO" bs=1 skip=510 count=2 2>/dev/null | xxd -p)
    if [ "$MBR_SIG" = "55aa" ]; then
        echo "✓ OK"
    else
        echo "⚠️  WARNING"
        echo "      MBR signature not found (may not boot on BIOS)"
        WARNINGS=$((WARNINGS + 1))
    fi
else
    echo "⊘ SKIP (no ISO)"
fi

# Check 4: Extract and verify modloop modifications
echo -n "[4/8] Verifying modloop modifications... "
if [ -f "$OUTPUT_ISO" ]; then
    TMP_DIR=$(mktemp -d)
    
    # Extract modloop from ISO
    xorriso -osirrox on -indev "$OUTPUT_ISO" -extract /boot/modloop-lts "$TMP_DIR/modloop-lts" 2>&1 > /dev/null
    
    if [ -f "$TMP_DIR/modloop-lts" ]; then
        # Extract modloop
        unsquashfs -d "$TMP_DIR/modloop_check" "$TMP_DIR/modloop-lts" > /dev/null 2>&1
        
        # Check for hwscan binary
        if [ -f "$TMP_DIR/modloop_check/usr/local/bin/hwscan" ]; then
            echo "✓ OK"
            echo "      hwscan binary found in modloop"
            
            # Check if it's executable
            if [ -x "$TMP_DIR/modloop_check/usr/local/bin/hwscan" ]; then
                echo "      Binary is executable ✓"
            else
                echo "      ⚠️  Binary is NOT executable"
                WARNINGS=$((WARNINGS + 1))
            fi
        else
            echo "❌ FAIL"
            echo "      hwscan binary NOT found in modloop"
            ERRORS=$((ERRORS + 1))
        fi
        
        # Check for autostart scripts
        if [ -f "$TMP_DIR/modloop_check/etc/local.d/50-hwscan.start" ]; then
            echo "      Autostart script found ✓"
        else
            echo "      ⚠️  Autostart script NOT found"
            WARNINGS=$((WARNINGS + 1))
        fi
    else
        echo "❌ FAIL"
        echo "      Could not extract modloop from ISO"
        ERRORS=$((ERRORS + 1))
    fi
    
    rm -rf "$TMP_DIR"
else
    echo "⊘ SKIP (no ISO)"
fi

# Check 5: Verify SYSLINUX config patches
echo -n "[5/8] Verifying SYSLINUX boot config... "
if [ -f "$OUTPUT_ISO" ]; then
    TMP_DIR=$(mktemp -d)
    xorriso -osirrox on -indev "$OUTPUT_ISO" -extract /boot/syslinux/syslinux.cfg "$TMP_DIR/syslinux.cfg" 2>&1 > /dev/null
    
    if [ -f "$TMP_DIR/syslinux.cfg" ]; then
        if grep -q "modloop_verify=no" "$TMP_DIR/syslinux.cfg"; then
            echo "✓ OK"
            echo "      Found 'modloop_verify=no' in SYSLINUX config"
        else
            echo "❌ FAIL"
            echo "      'modloop_verify=no' NOT found in SYSLINUX config"
            echo "      This will cause 'Failed to verify signature' error!"
            ERRORS=$((ERRORS + 1))
        fi
    else
        echo "⚠️  WARNING"
        echo "      Could not extract syslinux.cfg"
        WARNINGS=$((WARNINGS + 1))
    fi
    
    rm -rf "$TMP_DIR"
else
    echo "⊘ SKIP (no ISO)"
fi

# Check 6: Verify GRUB config patches
echo -n "[6/8] Verifying GRUB boot config... "
if [ -f "$OUTPUT_ISO" ]; then
    TMP_DIR=$(mktemp -d)
    xorriso -osirrox on -indev "$OUTPUT_ISO" -extract /boot/grub/grub.cfg "$TMP_DIR/grub.cfg" 2>&1 > /dev/null
    
    if [ -f "$TMP_DIR/grub.cfg" ]; then
        if grep -q "modloop_verify=no" "$TMP_DIR/grub.cfg"; then
            echo "✓ OK"
            echo "      Found 'modloop_verify=no' in GRUB config"
        else
            echo "⚠️  WARNING"
            echo "      'modloop_verify=no' NOT found in GRUB config"
            echo "      UEFI boot may fail with signature error"
            WARNINGS=$((WARNINGS + 1))
        fi
    else
        echo "⚠️  WARNING"
        echo "      Could not extract grub.cfg (may be normal)"
        WARNINGS=$((WARNINGS + 1))
    fi
    
    rm -rf "$TMP_DIR"
else
    echo "⊘ SKIP (no ISO)"
fi

# Check 7: Verify ISO is hybrid (can be dd'd to USB)
echo -n "[7/8] Checking ISO hybrid capability... "
if [ -f "$OUTPUT_ISO" ]; then
    # Check for GPT header (for hybrid ISO)
    GPT_SIG=$(dd if="$OUTPUT_ISO" bs=1 skip=512 count=8 2>/dev/null | xxd -p)
    if [[ "$GPT_SIG" == "4546492050415254"* ]]; then
        echo "✓ OK"
        echo "      ISO is hybrid (BIOS+UEFI), can be dd'd to USB"
    else
        echo "⚠️  WARNING"
        echo "      ISO may not be properly hybrid"
        WARNINGS=$((WARNINGS + 1))
    fi
else
    echo "⊘ SKIP (no ISO)"
fi

# Check 8: Binary architecture check
# Check 8: Binary architecture + static check (robusto)
echo -n "[8/8] Verifying hwscan binary architecture... "
BINARY_PATH="$SCRIPT_DIR/../../dist/hwscan"

if [ -f "$BINARY_PATH" ]; then
    FILE_OUTPUT=$(file "$BINARY_PATH")

    if echo "$FILE_OUTPUT" | grep -q "x86-64"; then
        echo "✓ OK"
        echo "      Binary is x86-64 (correct)"

        # Verificación estática robusta (no dependiente de idioma)
        if command -v readelf >/dev/null 2>&1; then
            if readelf -d "$BINARY_PATH" 2>&1 | grep -q "NEEDED"; then
                echo "      ⚠️  Binary has dynamic dependencies (DT_NEEDED found)"
                WARNINGS=$((WARNINGS + 1))
            else
                echo "      Binary is fully static (no dynamic section) ✓"
            fi
        else
            echo "      ⚠️  readelf not available, falling back to ldd"
            if ldd "$BINARY_PATH" 2>&1 | grep -qiE "not a dynamic|no es un ejecutable dinámico|statically linked"; then
                echo "      Binary is statically linked ✓"
            else
                echo "      ⚠️  Binary may be dynamically linked"
                WARNINGS=$((WARNINGS + 1))
            fi
        fi
    else
        echo "❌ FAIL"
        echo "      Binary is NOT x86-64!"
        echo "      Output: $FILE_OUTPUT"
        ERRORS=$((ERRORS + 1))
    fi
else
    echo "⚠️  WARNING"
    echo "      Source binary not found at: $BINARY_PATH"
    WARNINGS=$((WARNINGS + 1))
fi

# Summary
echo ""
echo "=================================================="
echo "  Verification Summary"
echo "=================================================="

if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo "✅ ALL CHECKS PASSED!"
    echo ""
    echo "Your ISO is ready to use:"
    echo "  $OUTPUT_ISO"
    echo ""
    echo "Next steps:"
    echo "  1. Test in QEMU:"
    echo "     qemu-system-x86_64 -cdrom \"$OUTPUT_ISO\" -m 2048 -serial stdio"
    echo ""
    echo "  2. Burn to USB:"
    echo "     sudo dd if=\"$OUTPUT_ISO\" of=/dev/sdX bs=4M status=progress"
    echo ""
elif [ $ERRORS -eq 0 ]; then
    echo "⚠️  BUILD COMPLETED WITH $WARNINGS WARNING(S)"
    echo ""
    echo "The ISO may work but please review warnings above."
    echo "Test thoroughly in QEMU before using in production."
    echo ""
else
    echo "❌ BUILD HAS $ERRORS ERROR(S) AND $WARNINGS WARNING(S)"
    echo ""
    echo "The ISO will likely NOT work correctly."
    echo "Please fix the errors above and rebuild."
    echo ""
    echo "To rebuild:"
    echo "  ./build-v3.sh"
    echo ""
fi

echo "=================================================="

exit $ERRORS