package hardware

// HardwareInfo contiene toda la información del hardware detectado
type HardwareInfo struct {
	MachineID   string          `json:"machine_id"` // Identificador único de la máquina
	CPU         CPUInfo         `json:"cpu"`
	Memory      MemoryInfo      `json:"memory"`
	Motherboard MotherboardInfo `json:"motherboard"`
	GPU         []GPUInfo       `json:"gpu"`
	Disks       []DiskInfo      `json:"disks"`
	Timestamp   string          `json:"timestamp"`
}

// CPUInfo contiene información del procesador
type CPUInfo struct {
	Model     string   `json:"model"`      // Modelo del procesador (ej: Intel Core i7-8700)
	Vendor    string   `json:"vendor"`     // Fabricante (Intel, AMD)
	Cores     int      `json:"cores"`      // Número de núcleos físicos
	Threads   int      `json:"threads"`    // Número de hilos lógicos
	Speed     float64  `json:"speed_mhz"`  // Velocidad en MHz
	CacheSize string   `json:"cache_size"` // Tamaño de caché
	Flags     []string `json:"flags"`      // Características del CPU
}

// MemoryInfo contiene información de la memoria RAM
type MemoryInfo struct {
	TotalGB    float64        `json:"total_gb"`    // Total de RAM en GB
	TotalBytes uint64         `json:"total_bytes"` // Total de RAM en bytes
	Modules    []MemoryModule `json:"modules"`     // Módulos de memoria individuales
}

// MemoryModule representa un módulo individual de RAM
type MemoryModule struct {
	Size         string `json:"size"`         // Tamaño (ej: 8GB)
	Type         string `json:"type"`         // Tipo (DDR4, DDR3, etc.)
	Speed        string `json:"speed"`        // Velocidad (ej: 2666MHz)
	Locator      string `json:"locator"`      // Ubicación física (DIMM1, etc.)
	Manufacturer string `json:"manufacturer"` // Fabricante
	PartNumber   string `json:"part_number"`  // Número de parte
}

// MotherboardInfo contiene información de la placa madre
type MotherboardInfo struct {
	Manufacturer string `json:"manufacturer"`  // Fabricante (ASUS, MSI, etc.)
	Product      string `json:"product"`       // Modelo (PRIME B360M-A, etc.)
	Version      string `json:"version"`       // Versión de la placa
	SerialNumber string `json:"serial_number"` // Número de serie
	BIOSVendor   string `json:"bios_vendor"`   // Fabricante del BIOS
	BIOSVersion  string `json:"bios_version"`  // Versión del BIOS
	BIOSDate     string `json:"bios_date"`     // Fecha del BIOS
}

// GPUInfo contiene información de la tarjeta gráfica
type GPUInfo struct {
	Vendor     string `json:"vendor"`      // Fabricante (NVIDIA, AMD, Intel)
	Model      string `json:"model"`       // Modelo (GTX 1060, RX 580, etc.)
	PCIAddress string `json:"pci_address"` // Dirección PCI
	Driver     string `json:"driver"`      // Driver en uso (si está disponible)
	MemorySize string `json:"memory_size"` // Tamaño de VRAM (si se puede detectar)
}

// DiskInfo contiene información de un disco de almacenamiento
type DiskInfo struct {
	Name      string  `json:"name"`       // Nombre del dispositivo (sda, nvme0n1)
	Model     string  `json:"model"`      // Modelo del disco
	Vendor    string  `json:"vendor"`     // Fabricante
	SizeGB    float64 `json:"size_gb"`    // Tamaño en GB
	SizeBytes uint64  `json:"size_bytes"` // Tamaño en bytes
	Type      string  `json:"type"`       // HDD, SSD, NVMe SSD
}
