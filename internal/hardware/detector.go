package hardware

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Detect realiza la detección completa del hardware del sistema
func Detect() (*HardwareInfo, error) {
	info := &HardwareInfo{
		Timestamp: time.Now().Format(time.RFC3339),
	}

	var err error

	// Detectar CPU
	info.CPU, err = detectCPU()
	if err != nil {
		return nil, fmt.Errorf("error detectando CPU: %w", err)
	}

	// Detectar Memoria
	info.Memory, err = detectMemory()
	if err != nil {
		return nil, fmt.Errorf("error detectando memoria: %w", err)
	}

	// Detectar Placa Madre
	info.Motherboard, err = detectMotherboard()
	if err != nil {
		// No es crítico, continuamos
		fmt.Printf("Advertencia: error detectando placa madre: %v\n", err)
	}

	// Detectar GPU
	info.GPU, err = detectGPU()
	if err != nil {
		// No es crítico, continuamos
		fmt.Printf("Advertencia: error detectando GPU: %v\n", err)
	}

	// Detectar Discos
	info.Disks, err = detectDisks()
	if err != nil {
		fmt.Printf("Advertencia: error detectando discos: %v\n", err)
	}

	// Generar Machine ID (debe ser al final para tener toda la info disponible)
	info.MachineID = GenerateMachineID(info)

	return info, nil
}

// detectCPU lee información del procesador desde /proc/cpuinfo
func detectCPU() (CPUInfo, error) {
	cpu := CPUInfo{
		Flags: make([]string, 0),
	}

	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return cpu, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	processorCount := 0
	coresMap := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "processor":
			processorCount++
		case "model name":
			if cpu.Model == "" {
				cpu.Model = value
			}
		case "vendor_id":
			if cpu.Vendor == "" {
				cpu.Vendor = value
			}
		case "cpu MHz":
			if cpu.Speed == 0 {
				speed, _ := strconv.ParseFloat(value, 64)
				cpu.Speed = speed
			}
		case "cache size":
			if cpu.CacheSize == "" {
				cpu.CacheSize = value
			}
		case "core id":
			coresMap[value] = true
		case "flags":
			if len(cpu.Flags) == 0 {
				cpu.Flags = strings.Fields(value)
			}
		}
	}

	cpu.Threads = processorCount
	cpu.Cores = len(coresMap)
	if cpu.Cores == 0 {
		cpu.Cores = processorCount
	}

	// Intentar obtener la velocidad máxima del CPU
	maxSpeed := getMaxCPUFrequency()
	if maxSpeed > 0 {
		cpu.Speed = maxSpeed
	}

	return cpu, scanner.Err()
}

// getMaxCPUFrequency intenta obtener la velocidad máxima del CPU en MHz
func getMaxCPUFrequency() float64 {
	// Intentar leer desde cpufreq (velocidad máxima del CPU)
	paths := []string{
		"/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_max_freq",
		"/sys/devices/system/cpu/cpu0/cpufreq/scaling_max_freq",
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err == nil {
			// El valor está en kHz, convertir a MHz
			khz := strings.TrimSpace(string(data))
			if freq, err := strconv.ParseFloat(khz, 64); err == nil {
				return freq / 1000 // kHz a MHz
			}
		}
	}

	return 0
}

// detectMemory lee información de memoria desde /proc/meminfo y dmidecode
func detectMemory() (MemoryInfo, error) {
	mem := MemoryInfo{
		Modules: make([]MemoryModule, 0),
	}

	// Leer memoria total desde /proc/meminfo
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return mem, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				totalKB, _ := strconv.ParseUint(parts[1], 10, 64)
				mem.TotalBytes = totalKB * 1024
				mem.TotalGB = float64(mem.TotalBytes) / (1024 * 1024 * 1024)
			}
			break
		}
	}

	// Intentar obtener información detallada con dmidecode
	modules := detectMemoryModules()
	if len(modules) > 0 {
		mem.Modules = modules
	} else {
		// Fallback: sintetizar entrada cuando dmidecode no está disponible
		mem.Modules = fallbackMemoryModules(mem.TotalGB)
	}

	return mem, nil
}

// fallbackMemoryModules crea una entrada sintética con la RAM total cuando
// dmidecode no está disponible o no devuelve información de módulos.
func fallbackMemoryModules(totalGB float64) []MemoryModule {
	if totalGB <= 0 {
		return nil
	}

	// Intentar obtener tipo de RAM desde /sys/firmware/dmi/tables (sin parsear)
	// y velocidad desde kernel para dar más información
	module := MemoryModule{
		Size: fmt.Sprintf("%.1f GB", totalGB),
	}

	// Intentar leer tipo de RAM desde dmidecode con timeout corto
	out, err := exec.Command("dmidecode", "-s", "memory-type").Output()
	if err == nil {
		t := strings.TrimSpace(string(out))
		if t != "" {
			module.Type = t
		}
	}

	return []MemoryModule{module}
}

// detectMemoryModules usa dmidecode para obtener información de módulos RAM
func detectMemoryModules() []MemoryModule {
	modules := make([]MemoryModule, 0)

	cmd := exec.Command("dmidecode", "-t", "memory")
	output, err := cmd.Output()
	if err != nil {
		// dmidecode puede no estar disponible o requiere privilegios
		return modules
	}

	var currentModule *MemoryModule
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Memory Device") {
			if currentModule != nil && currentModule.Size != "" && currentModule.Size != "No Module Installed" {
				modules = append(modules, *currentModule)
			}
			currentModule = &MemoryModule{}
		}

		if currentModule == nil {
			continue
		}

		if after, ok := strings.CutPrefix(line, "Size:"); ok {
			currentModule.Size = strings.TrimSpace(after)
		} else if after0, ok0 := strings.CutPrefix(line, "Type:"); ok0 {
			currentModule.Type = strings.TrimSpace(after0)
		} else if after1, ok1 := strings.CutPrefix(line, "Speed:"); ok1 {
			currentModule.Speed = strings.TrimSpace(after1)
		} else if after2, ok2 := strings.CutPrefix(line, "Locator:"); ok2 {
			currentModule.Locator = strings.TrimSpace(after2)
		} else if after3, ok3 := strings.CutPrefix(line, "Manufacturer:"); ok3 {
			currentModule.Manufacturer = strings.TrimSpace(after3)
		} else if after4, ok4 := strings.CutPrefix(line, "Part Number:"); ok4 {
			currentModule.PartNumber = strings.TrimSpace(after4)
		}
	}

	// Agregar el último módulo
	if currentModule != nil && currentModule.Size != "" && currentModule.Size != "No Module Installed" {
		modules = append(modules, *currentModule)
	}

	return modules
}

// detectDisks lee información de discos desde /sys/block
func detectDisks() ([]DiskInfo, error) {
	disks := make([]DiskInfo, 0)

	entries, err := os.ReadDir("/sys/block")
	if err != nil {
		return disks, err
	}

	for _, entry := range entries {
		name := entry.Name()
		// Ignorar dispositivos virtuales
		if strings.HasPrefix(name, "loop") ||
			strings.HasPrefix(name, "ram") ||
			strings.HasPrefix(name, "dm-") ||
			strings.HasPrefix(name, "zram") ||
			strings.HasPrefix(name, "sr") {
			continue
		}

		disk := DiskInfo{Name: name}
		basePath := "/sys/block/" + name

		// Tamaño del dispositivo (sectores de 512 bytes)
		if data, err := os.ReadFile(basePath + "/size"); err == nil {
			if sectors, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64); err == nil {
				disk.SizeBytes = sectors * 512
				disk.SizeGB = float64(disk.SizeBytes) / (1024 * 1024 * 1024)
			}
		}

		// Ignorar dispositivos muy pequeños (< 1 MB)
		if disk.SizeBytes < 1024*1024 {
			continue
		}

		// Modelo del disco
		for _, modelPath := range []string{
			basePath + "/device/model",
			basePath + "/device/name",
		} {
			if data, err := os.ReadFile(modelPath); err == nil {
				disk.Model = strings.TrimSpace(string(data))
				break
			}
		}

		// Fabricante
		if data, err := os.ReadFile(basePath + "/device/vendor"); err == nil {
			disk.Vendor = strings.TrimSpace(string(data))
		}

		// Tipo: NVMe, SSD o HDD
		if strings.HasPrefix(name, "nvme") {
			disk.Type = "NVMe SSD"
		} else if data, err := os.ReadFile(basePath + "/queue/rotational"); err == nil {
			if strings.TrimSpace(string(data)) == "0" {
				disk.Type = "SSD"
			} else {
				disk.Type = "HDD"
			}
		}

		disks = append(disks, disk)
	}

	return disks, nil
}

// detectMotherboard lee información de la placa madre desde /sys/class/dmi/id
func detectMotherboard() (MotherboardInfo, error) {
	mb := MotherboardInfo{}
	dmiPath := "/sys/class/dmi/id"

	files := map[string]*string{
		"board_vendor":  &mb.Manufacturer,
		"board_name":    &mb.Product,
		"board_version": &mb.Version,
		"board_serial":  &mb.SerialNumber,
		"bios_vendor":   &mb.BIOSVendor,
		"bios_version":  &mb.BIOSVersion,
		"bios_date":     &mb.BIOSDate,
	}

	for filename, target := range files {
		path := filepath.Join(dmiPath, filename)
		data, err := os.ReadFile(path)
		if err == nil {
			*target = strings.TrimSpace(string(data))
		}
	}

	return mb, nil
}

// detectGPU usa lspci para detectar tarjetas gráficas
func detectGPU() ([]GPUInfo, error) {
	gpus := make([]GPUInfo, 0)

	cmd := exec.Command("lspci")
	output, err := cmd.Output()
	if err != nil {
		return gpus, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// Buscar líneas que contengan VGA o 3D
		if strings.Contains(line, "VGA compatible controller") ||
			strings.Contains(line, "3D controller") ||
			strings.Contains(line, "Display controller") {

			parts := strings.SplitN(line, " ", 2)
			if len(parts) < 2 {
				continue
			}

			pciAddress := strings.TrimSpace(parts[0])
			description := strings.TrimSpace(parts[1])

			// Extraer información del GPU
			gpu := GPUInfo{
				PCIAddress: pciAddress,
			}

			// Detectar vendor
			descLower := strings.ToLower(description)
			if strings.Contains(descLower, "nvidia") {
				gpu.Vendor = "NVIDIA"
			} else if strings.Contains(descLower, "amd") || strings.Contains(descLower, "ati") {
				gpu.Vendor = "AMD"
			} else if strings.Contains(descLower, "intel") {
				gpu.Vendor = "Intel"
			}

			// El modelo es toda la descripción después del tipo
			if idx := strings.Index(description, ":"); idx > 0 {
				gpu.Model = strings.TrimSpace(description[idx+1:])
			} else {
				gpu.Model = description
			}

			// Detectar VRAM
			gpu.MemorySize = getVRAM(pciAddress)

			gpus = append(gpus, gpu)
		}
	}

	return gpus, nil
}

// getVRAM intenta detectar la VRAM de una GPU a partir de su dirección PCI.
//
//   - Estrategia 1: /proc/driver/nvidia/gpus/<addr>/information  (NVIDIA driver propietario)
//   - Estrategia 2: nvidia-smi --query-gpu=memory.total          (si nvidia-smi está presente)
//   - Estrategia 3: sysfs DRM mem_info_vram_total                (AMD / NVIDIA open)
//   - Estrategia 4: lspci -v BAR prefetchable >= 512 MB          (último recurso; filtra apertura 256MB)
func getVRAM(pciAddress string) string {
	// Normalizar dirección: lspci puede omitir el dominio "0000:"
	fullAddr := pciAddress
	if len(strings.Split(pciAddress, ":")) == 2 {
		fullAddr = "0000:" + pciAddress
	}

	// Estrategia 1: /proc/driver/nvidia/gpus/<addr>/information
	// El kernel NVIDIA escribe aquí "Video Memory: 4096 MB"
	for _, candidate := range []string{fullAddr, pciAddress} {
		infoPath := "/proc/driver/nvidia/gpus/" + candidate + "/information"
		if data, err := os.ReadFile(infoPath); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(line, "Video Memory:") {
					val := strings.TrimSpace(strings.TrimPrefix(line, "Video Memory:"))
					// formato: "4096 MB" o "4096MB"
					val = strings.ReplaceAll(val, " ", "")
					if b := parseVRAMSize(val); b > 0 {
						return formatVRAMBytes(b)
					}
				}
			}
		}
	}

	// Estrategia 2: nvidia-smi (devuelve MiB, ej: "4096 MiB" o solo "4096")
	if out, err := exec.Command("nvidia-smi",
		"--query-gpu=memory.total",
		"--format=csv,noheader,nounits",
		"--id="+pciAddress).Output(); err == nil {
		val := strings.TrimSpace(string(out))
		// nounits → valor en MiB como número entero
		if mib, err := strconv.ParseUint(val, 10, 64); err == nil && mib > 0 {
			return formatVRAMBytes(mib * 1024 * 1024)
		}
	}

	// Estrategia 3: sysfs DRM — mem_info_vram_total (AMD + módulo open NVIDIA)
	cards, _ := filepath.Glob("/sys/class/drm/card*/device")
	for _, cardDev := range cards {
		resolved, err := filepath.EvalSymlinks(cardDev)
		if err != nil {
			continue
		}
		if strings.HasSuffix(resolved, fullAddr) || strings.HasSuffix(resolved, pciAddress) {
			if data, err := os.ReadFile(cardDev + "/mem_info_vram_total"); err == nil {
				if b, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64); err == nil && b > 0 {
					return formatVRAMBytes(b)
				}
			}
		}
	}

	// Estrategia 4: lspci -v — BAR prefetchable mayor.
	// NOTA: NVIDIA expone solo una apertura de 256 MB como BAR prefetchable con el driver
	// propietario. Filtramos resultados < 512 MB para no reportar la apertura como VRAM.
	out, err := exec.Command("lspci", "-v", "-s", pciAddress).Output()
	if err != nil {
		return ""
	}

	var maxBytes uint64
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "prefetchable") || strings.Contains(line, "non-prefetchable") {
			continue
		}
		idx := strings.Index(line, "[size=")
		if idx < 0 {
			continue
		}
		sizeStr := line[idx+6:]
		if end := strings.Index(sizeStr, "]"); end > 0 {
			sizeStr = sizeStr[:end]
		}
		if b := parseVRAMSize(sizeStr); b > maxBytes {
			maxBytes = b
		}
	}
	// Solo reportar si el BAR es >= 512 MB para evitar falsos positivos (apertura NVIDIA)
	if maxBytes >= 512*1024*1024 {
		return formatVRAMBytes(maxBytes)
	}
	return ""
}

// parseVRAMSize convierte strings como "8G", "256M", "512K" a bytes.
func parseVRAMSize(s string) uint64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	multipliers := map[byte]uint64{'G': 1 << 30, 'M': 1 << 20, 'K': 1 << 10}
	if mult, ok := multipliers[s[len(s)-1]]; ok {
		if n, err := strconv.ParseFloat(s[:len(s)-1], 64); err == nil {
			return uint64(n * float64(mult))
		}
	}
	n, _ := strconv.ParseUint(s, 10, 64)
	return n
}

// formatVRAMBytes formatea bytes de VRAM en una cadena legible ("8 GB", "512 MB").
func formatVRAMBytes(b uint64) string {
	if gb := float64(b) / (1 << 30); gb >= 1 {
		return fmt.Sprintf("%.0f GB", gb)
	}
	return fmt.Sprintf("%.0f MB", float64(b)/(1<<20))
}
