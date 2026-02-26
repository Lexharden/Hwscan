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
	}

	return mem, nil
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

		if strings.HasPrefix(line, "Size:") {
			currentModule.Size = strings.TrimSpace(strings.TrimPrefix(line, "Size:"))
		} else if strings.HasPrefix(line, "Type:") {
			currentModule.Type = strings.TrimSpace(strings.TrimPrefix(line, "Type:"))
		} else if strings.HasPrefix(line, "Speed:") {
			currentModule.Speed = strings.TrimSpace(strings.TrimPrefix(line, "Speed:"))
		} else if strings.HasPrefix(line, "Locator:") {
			currentModule.Locator = strings.TrimSpace(strings.TrimPrefix(line, "Locator:"))
		} else if strings.HasPrefix(line, "Manufacturer:") {
			currentModule.Manufacturer = strings.TrimSpace(strings.TrimPrefix(line, "Manufacturer:"))
		} else if strings.HasPrefix(line, "Part Number:") {
			currentModule.PartNumber = strings.TrimSpace(strings.TrimPrefix(line, "Part Number:"))
		}
	}

	// Agregar el último módulo
	if currentModule != nil && currentModule.Size != "" && currentModule.Size != "No Module Installed" {
		modules = append(modules, *currentModule)
	}

	return modules
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

			gpus = append(gpus, gpu)
		}
	}

	return gpus, nil
}
