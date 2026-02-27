package hardware

import (
	"fmt"
	"strings"

	"github.com/Lexharden/hwscan/internal/utils"
)

// FormatConsole formatea la información de hardware para mostrar en consola
func FormatConsole(info *HardwareInfo) string {
	var sb strings.Builder
	sb.Grow(4096) // Prealocación aproximada para evitar realocaciones

	// Header
	fmt.Fprintln(&sb, "╔══════════════════════════════════════════════════════════════╗")
	fmt.Fprintln(&sb, "║                      HWSCAN v1.0                             ║")
	fmt.Fprintln(&sb, "║                 Hardware Detection Tool                      ║")
	fmt.Fprintln(&sb, "║           Desarrollado por: Yafel Garcia (Lexharden)         ║")
	fmt.Fprintln(&sb, "╚══════════════════════════════════════════════════════════════╝")
	fmt.Fprintln(&sb)

	// Machine ID
	fmt.Fprintln(&sb, "┌─ IDENTIFICACIÓN ─────────────────────────────────────────────┐")
	fmt.Fprintf(&sb, "│ Machine ID: %s\n", info.MachineID)
	fmt.Fprintln(&sb, "└──────────────────────────────────────────────────────────────┘")
	fmt.Fprintln(&sb)

	// CPU
	fmt.Fprintln(&sb, "┌─ CPU ────────────────────────────────────────────────────────┐")
	fmt.Fprintf(&sb, "│ Modelo:    %s\n", info.CPU.Model)
	fmt.Fprintf(&sb, "│ Vendor:    %s\n", info.CPU.Vendor)
	fmt.Fprintf(&sb, "│ Cores:     %d físicos / %d hilos\n", info.CPU.Cores, info.CPU.Threads)
	fmt.Fprintf(&sb, "│ Velocidad: %.2f GHz\n", info.CPU.Speed/1000)
	if info.CPU.CacheSize != "" {
		fmt.Fprintf(&sb, "│ Caché:     %s\n", info.CPU.CacheSize)
	}
	fmt.Fprintln(&sb, "└──────────────────────────────────────────────────────────────┘")
	fmt.Fprintln(&sb)

	// Memoria
	fmt.Fprintln(&sb, "┌─ MEMORIA RAM ────────────────────────────────────────────────┐")
	fmt.Fprintf(&sb, "│ Total:     %.2f GB (%.0f bytes)\n",
		info.Memory.TotalGB,
		float64(info.Memory.TotalBytes),
	)

	if len(info.Memory.Modules) > 0 {
		fmt.Fprintln(&sb, "│")
		fmt.Fprintln(&sb, "│ Módulos instalados:")
		for i, mod := range info.Memory.Modules {
			fmt.Fprintf(&sb, "│  [%d] %s %s %s", i+1, mod.Size, mod.Type, mod.Speed)

			if mod.Locator != "" {
				fmt.Fprintf(&sb, " (%s)", mod.Locator)
			}
			fmt.Fprintln(&sb)

			if mod.Manufacturer != "" && mod.Manufacturer != "Unknown" {
				fmt.Fprintf(&sb, "│      Fabricante: %s", mod.Manufacturer)
				if mod.PartNumber != "" {
					fmt.Fprintf(&sb, " | P/N: %s", mod.PartNumber)
				}
				fmt.Fprintln(&sb)
			}
		}
	}
	fmt.Fprintln(&sb, "└──────────────────────────────────────────────────────────────┘")
	fmt.Fprintln(&sb)

	// Placa Madre
	fmt.Fprintln(&sb, "┌─ PLACA MADRE ────────────────────────────────────────────────┐")
	fmt.Fprintf(&sb, "│ Fabricante: %s\n", info.Motherboard.Manufacturer)
	fmt.Fprintf(&sb, "│ Modelo:     %s\n", info.Motherboard.Product)

	if info.Motherboard.Version != "" {
		fmt.Fprintf(&sb, "│ Versión:    %s\n", info.Motherboard.Version)
	}
	if info.Motherboard.BIOSVendor != "" {
		fmt.Fprintf(&sb, "│ BIOS:       %s v%s (%s)\n",
			info.Motherboard.BIOSVendor,
			info.Motherboard.BIOSVersion,
			info.Motherboard.BIOSDate,
		)
	}

	fmt.Fprintln(&sb, "└──────────────────────────────────────────────────────────────┘")
	fmt.Fprintln(&sb)

	// GPU
	if len(info.GPU) > 0 {
		fmt.Fprintln(&sb, "┌─ GPU ────────────────────────────────────────────────────────┐")
		for i, gpu := range info.GPU {
			fmt.Fprintf(&sb, "│ [%d] %s %s\n", i+1, gpu.Vendor, gpu.Model)
			fmt.Fprintf(&sb, "│     PCI: %s\n", gpu.PCIAddress)

			if gpu.MemorySize != "" {
				fmt.Fprintf(&sb, "│     VRAM: %s\n", gpu.MemorySize)
			}

			if i < len(info.GPU)-1 {
				fmt.Fprintln(&sb, "│")
			}
		}
		fmt.Fprintln(&sb, "└──────────────────────────────────────────────────────────────┘")
		fmt.Fprintln(&sb)
	}

	// Discos
	if len(info.Disks) > 0 {
		fmt.Fprintln(&sb, "┌─ ALMACENAMIENTO ────────────────────────────────────────────────┐")
		for i, disk := range info.Disks {
			model := disk.Model
			if model == "" {
				model = disk.Name
			}

			fmt.Fprintf(&sb, "│ [%d] %s", i+1, model)
			if disk.Vendor != "" {
				fmt.Fprintf(&sb, " (%s)", disk.Vendor)
			}
			fmt.Fprintln(&sb)

			fmt.Fprintf(&sb,
				"│     Capacidad: %.1f GB | Tipo: %s | Dev: /dev/%s\n",
				disk.SizeGB,
				disk.Type,
				disk.Name,
			)

			if i < len(info.Disks)-1 {
				fmt.Fprintln(&sb, "│")
			}
		}
		fmt.Fprintln(&sb, "└──────────────────────────────────────────────────────────────")
		fmt.Fprintln(&sb)
	}

	// Footer
	fmt.Fprintln(&sb, "═══════════════════════════════════════════════════════════════")
	fmt.Fprintf(&sb, " Interfaz Web: http://%s:8080\n", utils.GetLocalIP())
	fmt.Fprintf(&sb, " Fecha/Hora:   %s\n", info.Timestamp)
	fmt.Fprintln(&sb, "═══════════════════════════════════════════════════════════════")

	return sb.String()
}
