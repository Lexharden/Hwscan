package hardware

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strings"
)

// GenerateMachineID genera un identificador único y estable para la máquina
// basado en características de hardware permanentes.
//
// Estrategia de generación (en orden de prioridad):
// 1. DMI Product UUID del sistema (/sys/class/dmi/id/product_uuid)
// 2. Hash SHA-256 de: board_serial + motherboard + cpu + ram
// 3. Hash SHA-256 de: mac_address + cpu + ram (último recurso)
//
// El ID resultante tiene el formato: HWSCAN-<UPPERCASE_HEX>
// Ejemplo: HWSCAN-4C4C4544003237108037B7C04F343132
//
// Este ID debe permanecer estable incluso después de:
// - Reinstalación del sistema operativo
// - Cambios de disco duro
// - Actualizaciones de BIOS (en la mayoría de casos)
//
// Puede cambiar si se reemplaza:
// - Placa madre
// - CPU (en sistemas sin UUID de hardware)
// - Toda la RAM (en sistemas sin UUID de hardware)
func GenerateMachineID(info *HardwareInfo) string {
	// Estrategia 1: Intentar usar DMI Product UUID
	if uuid := readDMIProductUUID(); isValidUUID(uuid) {
		// Limpiar y formatear el UUID
		cleanUUID := strings.ReplaceAll(uuid, "-", "")
		cleanUUID = strings.ReplaceAll(cleanUUID, " ", "")
		cleanUUID = strings.ToUpper(cleanUUID)
		return fmt.Sprintf("HWSCAN-%s", cleanUUID)
	}

	// Estrategia 2: Hash de información de hardware DMI
	if id := generateFromDMI(info); id != "" {
		return id
	}

	// Estrategia 3: Hash de MAC address + hardware básico
	if id := generateFromMAC(info); id != "" {
		return id
	}

	// Último recurso: hash del timestamp y datos disponibles
	// Esto no es ideal pero garantiza que siempre haya un ID
	return generateFallbackID(info)
}

// readDMIProductUUID lee el UUID del producto desde DMI/SMBIOS
// Este UUID es único por máquina y lo asigna el fabricante
func readDMIProductUUID() string {
	data, err := os.ReadFile("/sys/class/dmi/id/product_uuid")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// isValidUUID verifica que un UUID sea válido y no sea un valor placeholder
func isValidUUID(uuid string) bool {
	if uuid == "" {
		return false
	}

	// Normalizar para comparación
	normalized := strings.ToLower(strings.ReplaceAll(uuid, "-", ""))
	normalized = strings.TrimSpace(normalized)

	// Verificar longitud (debe ser 32 caracteres hex sin guiones)
	if len(normalized) != 32 {
		return false
	}

	// Verificar que no sea todo ceros
	if normalized == "00000000000000000000000000000000" {
		return false
	}

	// Verificar que no sea todo F's (otro valor común inválido)
	if normalized == "ffffffffffffffffffffffffffffffff" {
		return false
	}

	// Verificar que contenga solo caracteres hexadecimales
	for _, c := range normalized {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}

	return true
}

// generateFromDMI genera un ID basado en información DMI del sistema
// Usa: serial de placa + modelo de placa + modelo de CPU + RAM total
func generateFromDMI(info *HardwareInfo) string {
	var components []string

	// Agregar serial de la placa madre (más importante)
	if info.Motherboard.SerialNumber != "" &&
		info.Motherboard.SerialNumber != "Default string" &&
		info.Motherboard.SerialNumber != "To Be Filled By O.E.M." {
		components = append(components, info.Motherboard.SerialNumber)
	}

	// Agregar fabricante y modelo de placa madre
	if info.Motherboard.Manufacturer != "" {
		components = append(components, info.Motherboard.Manufacturer)
	}
	if info.Motherboard.Product != "" {
		components = append(components, info.Motherboard.Product)
	}

	// Agregar modelo de CPU
	if info.CPU.Model != "" {
		components = append(components, info.CPU.Model)
	}

	// Agregar RAM total (en MB para evitar pequeñas variaciones)
	ramMB := info.Memory.TotalBytes / (1024 * 1024)
	components = append(components, fmt.Sprintf("%d", ramMB))

	// Si no hay suficientes componentes, no es confiable
	if len(components) < 2 {
		return ""
	}

	// Generar hash SHA-256
	combined := strings.Join(components, "|")
	hash := sha256.Sum256([]byte(combined))

	// Tomar los primeros 16 bytes (32 caracteres hex)
	hexHash := strings.ToUpper(hex.EncodeToString(hash[:16]))

	return fmt.Sprintf("HWSCAN-%s", hexHash)
}

// generateFromMAC genera un ID basado en la dirección MAC principal
// Esto es menos ideal pero funciona cuando no hay información DMI
func generateFromMAC(info *HardwareInfo) string {
	// Obtener la primera MAC address válida (no loopback, no virtual)
	mac := getPrimaryMACAddress()
	if mac == "" {
		return ""
	}

	var components []string
	components = append(components, mac)

	// Agregar CPU si está disponible
	if info.CPU.Model != "" {
		components = append(components, info.CPU.Model)
	}

	// Agregar RAM total
	ramMB := info.Memory.TotalBytes / (1024 * 1024)
	components = append(components, fmt.Sprintf("%d", ramMB))

	// Generar hash
	combined := strings.Join(components, "|")
	hash := sha256.Sum256([]byte(combined))
	hexHash := strings.ToUpper(hex.EncodeToString(hash[:16]))

	return fmt.Sprintf("HWSCAN-%s", hexHash)
}

// getPrimaryMACAddress obtiene la dirección MAC de la interfaz de red principal
// Ignora interfaces loopback, virtuales y desconectadas
func getPrimaryMACAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	// Buscar la primera interfaz física válida
	for _, iface := range interfaces {
		// Ignorar loopback
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Ignorar interfaces sin dirección MAC
		if len(iface.HardwareAddr) == 0 {
			continue
		}

		// Ignorar interfaces que parecen virtuales
		name := strings.ToLower(iface.Name)
		if strings.Contains(name, "docker") ||
			strings.Contains(name, "veth") ||
			strings.Contains(name, "br-") ||
			strings.Contains(name, "vir") {
			continue
		}

		// Retornar la primera MAC válida
		return iface.HardwareAddr.String()
	}

	return ""
}

// generateFallbackID genera un ID de último recurso
// Usa timestamp + cualquier dato disponible
// NOTA: Este ID NO será estable entre reinicios
func generateFallbackID(info *HardwareInfo) string {
	var components []string

	// Agregar cualquier información disponible
	if info.CPU.Model != "" {
		components = append(components, info.CPU.Model)
	}
	if info.CPU.Vendor != "" {
		components = append(components, info.CPU.Vendor)
	}

	ramMB := info.Memory.TotalBytes / (1024 * 1024)
	components = append(components, fmt.Sprintf("%d", ramMB))

	if info.Motherboard.Manufacturer != "" {
		components = append(components, info.Motherboard.Manufacturer)
	}

	// Agregar timestamp para garantizar unicidad
	// ADVERTENCIA: Esto hace que el ID cambie en cada ejecución
	components = append(components, info.Timestamp)

	combined := strings.Join(components, "|")
	hash := sha256.Sum256([]byte(combined))
	hexHash := strings.ToUpper(hex.EncodeToString(hash[:16]))

	return fmt.Sprintf("HWSCAN-TEMP-%s", hexHash)
}
