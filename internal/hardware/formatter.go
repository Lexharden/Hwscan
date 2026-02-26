package hardware

import (
	"fmt"
	"strings"

	"github.com/Lexharden/hwscan/internal/utils"
)

// FormatConsole formatea la información de hardware para mostrar en consola
func FormatConsole(info *HardwareInfo) string {
	var sb strings.Builder

	sb.WriteString("╔══════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                    HWSCAN v1.0                               ║\n")
	sb.WriteString("║              Hardware Detection Tool                         ║\n")
	sb.WriteString("║              Desarrollado por: Yafel Garcia                  ║\n")
	sb.WriteString("╚══════════════════════════════════════════════════════════════╝\n\n")

	// Machine ID
	sb.WriteString("┌─ IDENTIFICACIÓN ─────────────────────────────────────────────┐\n")
	sb.WriteString(fmt.Sprintf("│ Machine ID: %s\n", info.MachineID))
	sb.WriteString("└──────────────────────────────────────────────────────────────┘\n\n")

	// CPU
	sb.WriteString("┌─ CPU ────────────────────────────────────────────────────────┐\n")
	sb.WriteString(fmt.Sprintf("│ Modelo:    %s\n", info.CPU.Model))
	sb.WriteString(fmt.Sprintf("│ Vendor:    %s\n", info.CPU.Vendor))
	sb.WriteString(fmt.Sprintf("│ Cores:     %d físicos / %d hilos\n", info.CPU.Cores, info.CPU.Threads))
	sb.WriteString(fmt.Sprintf("│ Velocidad: %.2f GHz\n", info.CPU.Speed/1000))
	if info.CPU.CacheSize != "" {
		sb.WriteString(fmt.Sprintf("│ Caché:     %s\n", info.CPU.CacheSize))
	}
	sb.WriteString("└──────────────────────────────────────────────────────────────┘\n\n")

	// Memoria
	sb.WriteString("┌─ MEMORIA RAM ────────────────────────────────────────────────┐\n")
	sb.WriteString(fmt.Sprintf("│ Total:     %.2f GB (%.0f bytes)\n", info.Memory.TotalGB, float64(info.Memory.TotalBytes)))
	if len(info.Memory.Modules) > 0 {
		sb.WriteString("│\n│ Módulos instalados:\n")
		for i, mod := range info.Memory.Modules {
			sb.WriteString(fmt.Sprintf("│  [%d] %s %s %s", i+1, mod.Size, mod.Type, mod.Speed))
			if mod.Locator != "" {
				sb.WriteString(fmt.Sprintf(" (%s)", mod.Locator))
			}
			sb.WriteString("\n")
			if mod.Manufacturer != "" && mod.Manufacturer != "Unknown" {
				sb.WriteString(fmt.Sprintf("│      Fabricante: %s", mod.Manufacturer))
				if mod.PartNumber != "" {
					sb.WriteString(fmt.Sprintf(" | P/N: %s", mod.PartNumber))
				}
				sb.WriteString("\n")
			}
		}
	}
	sb.WriteString("└──────────────────────────────────────────────────────────────┘\n\n")

	// Placa Madre
	sb.WriteString("┌─ PLACA MADRE ────────────────────────────────────────────────┐\n")
	sb.WriteString(fmt.Sprintf("│ Fabricante: %s\n", info.Motherboard.Manufacturer))
	sb.WriteString(fmt.Sprintf("│ Modelo:     %s\n", info.Motherboard.Product))
	if info.Motherboard.Version != "" {
		sb.WriteString(fmt.Sprintf("│ Versión:    %s\n", info.Motherboard.Version))
	}
	if info.Motherboard.BIOSVendor != "" {
		sb.WriteString(fmt.Sprintf("│ BIOS:       %s v%s (%s)\n",
			info.Motherboard.BIOSVendor,
			info.Motherboard.BIOSVersion,
			info.Motherboard.BIOSDate))
	}
	sb.WriteString("└──────────────────────────────────────────────────────────────┘\n\n")

	// GPU
	if len(info.GPU) > 0 {
		sb.WriteString("┌─ GPU ────────────────────────────────────────────────────────┐\n")
		for i, gpu := range info.GPU {
			sb.WriteString(fmt.Sprintf("│ [%d] %s %s\n", i+1, gpu.Vendor, gpu.Model))
			sb.WriteString(fmt.Sprintf("│     PCI: %s\n", gpu.PCIAddress))
			if gpu.MemorySize != "" {
				sb.WriteString(fmt.Sprintf("│     VRAM: %s\n", gpu.MemorySize))
			}
			if i < len(info.GPU)-1 {
				sb.WriteString("│\n")
			}
		}
		sb.WriteString("└──────────────────────────────────────────────────────────────┘\n\n")
	}

	sb.WriteString("═══════════════════════════════════════════════════════════════\n")
	sb.WriteString(fmt.Sprintf(" Interfaz Web: http://%s:8080\n", utils.GetLocalIP()))
	sb.WriteString(fmt.Sprintf(" Fecha/Hora:   %s\n", info.Timestamp))
	sb.WriteString("═══════════════════════════════════════════════════════════════\n")

	return sb.String()
}
