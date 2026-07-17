package hardware

// HardwareInfo contiene toda la información del hardware detectado.
// Los campos vacíos, listas vacías y estructuras sin datos se omiten en JSON.
type HardwareInfo struct {
	MachineID   string          `json:"machine_id,omitempty"`
	CPU         CPUInfo         `json:"cpu,omitempty"`
	Memory      MemoryInfo      `json:"memory,omitempty"`
	Motherboard MotherboardInfo `json:"motherboard,omitempty"`
	GPU         []GPUInfo       `json:"gpu,omitempty"`
	Disks       []DiskInfo      `json:"disks,omitempty"`
	Timestamp   string          `json:"timestamp,omitempty"`
}

// CPUInfo contiene información del procesador
type CPUInfo struct {
	Model     string   `json:"model,omitempty"`
	Vendor    string   `json:"vendor,omitempty"`
	Cores     int      `json:"cores,omitempty"`
	Threads   int      `json:"threads,omitempty"`
	Speed     float64  `json:"speed_mhz,omitempty"`
	CacheSize string   `json:"cache_size,omitempty"`
	Flags     []string `json:"flags,omitempty"`
}

// MemoryInfo contiene información de la memoria RAM
type MemoryInfo struct {
	TotalGB    float64        `json:"total_gb,omitempty"`
	TotalBytes uint64         `json:"total_bytes,omitempty"`
	Modules    []MemoryModule `json:"modules,omitempty"`
}

// MemoryModule representa un módulo individual de RAM
type MemoryModule struct {
	Size         string `json:"size,omitempty"`
	Type         string `json:"type,omitempty"`
	Speed        string `json:"speed,omitempty"`
	Locator      string `json:"locator,omitempty"`
	Manufacturer string `json:"manufacturer,omitempty"`
	PartNumber   string `json:"part_number,omitempty"`
}

// MotherboardInfo contiene información de la placa madre
type MotherboardInfo struct {
	Manufacturer string `json:"manufacturer,omitempty"`
	Product      string `json:"product,omitempty"`
	Version      string `json:"version,omitempty"`
	SerialNumber string `json:"serial_number,omitempty"`
	BIOSVendor   string `json:"bios_vendor,omitempty"`
	BIOSVersion  string `json:"bios_version,omitempty"`
	BIOSDate     string `json:"bios_date,omitempty"`
}

// GPUInfo contiene información de la tarjeta gráfica
type GPUInfo struct {
	Vendor     string `json:"vendor,omitempty"`
	Model      string `json:"model,omitempty"`
	PCIAddress string `json:"pci_address,omitempty"`
	Driver     string `json:"driver,omitempty"`
	MemorySize string `json:"memory_size,omitempty"`
}

// DiskInfo contiene información de un disco de almacenamiento
type DiskInfo struct {
	Name      string  `json:"name,omitempty"`
	Model     string  `json:"model,omitempty"`
	Vendor    string  `json:"vendor,omitempty"`
	SizeGB    float64 `json:"size_gb,omitempty"`
	SizeBytes uint64  `json:"size_bytes,omitempty"`
	Type      string  `json:"type,omitempty"`
}
