package export

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lexharden/hwscan/internal/hardware"
)

// ExportToJSON exporta la información de hardware a un archivo JSON
func ExportToJSON(info *hardware.HardwareInfo, outputPath string) error {
	// Crear el contenido JSON con formato legible
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("error al serializar JSON: %w", err)
	}

	// Escribir el archivo
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("error al escribir archivo: %w", err)
	}

	return nil
}

// AutoExport intenta exportar automáticamente a un dispositivo USB
func AutoExport(info *hardware.HardwareInfo) (string, error) {
	// Buscar dispositivos USB montados
	usbPaths := findUSBMounts()

	if len(usbPaths) == 0 {
		// Si no hay USB, exportar al directorio actual
		filename := generateFilename()
		if err := ExportToJSON(info, filename); err != nil {
			return "", err
		}
		return filename, nil
	}

	// Exportar al primer USB encontrado
	usbPath := usbPaths[0]
	filename := filepath.Join(usbPath, generateFilename())

	if err := ExportToJSON(info, filename); err != nil {
		// Si falla en USB, intentar en directorio actual
		localFilename := generateFilename()
		if err := ExportToJSON(info, localFilename); err != nil {
			return "", err
		}
		return localFilename, nil
	}

	return filename, nil
}

// generateFilename genera un nombre de archivo con timestamp
func generateFilename() string {
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("hwscan-%s.json", timestamp)
}

// findUSBMounts busca dispositivos USB montados en el sistema
func findUSBMounts() []string {
	var usbPaths []string

	// Directorios comunes donde se montan USBs en Linux
	mountDirs := []string{
		"/media",
		"/mnt",
		"/run/media",
	}

	for _, baseDir := range mountDirs {
		entries, err := os.ReadDir(baseDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				fullPath := filepath.Join(baseDir, entry.Name())

				// Verificar si el directorio no está vacío
				subEntries, err := os.ReadDir(fullPath)
				if err != nil {
					continue
				}

				// Si tiene subdirectorios, probablemente es un punto de montaje USB
				for _, subEntry := range subEntries {
					if subEntry.IsDir() {
						subPath := filepath.Join(fullPath, subEntry.Name())

						// Verificar que podamos escribir
						testFile := filepath.Join(subPath, ".hwscan_test")
						if err := os.WriteFile(testFile, []byte("test"), 0644); err == nil {
							os.Remove(testFile)
							usbPaths = append(usbPaths, subPath)
						}
					}
				}
			}
		}
	}

	// También buscar en /media/$USER/
	if username := os.Getenv("USER"); username != "" {
		userMediaPath := filepath.Join("/media", username)
		entries, err := os.ReadDir(userMediaPath)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					fullPath := filepath.Join(userMediaPath, entry.Name())

					// Verificar escritura
					testFile := filepath.Join(fullPath, ".hwscan_test")
					if err := os.WriteFile(testFile, []byte("test"), 0644); err == nil {
						os.Remove(testFile)

						// Evitar duplicados
						found := false
						for _, existing := range usbPaths {
							if existing == fullPath {
								found = true
								break
							}
						}
						if !found {
							usbPaths = append(usbPaths, fullPath)
						}
					}
				}
			}
		}
	}

	return usbPaths
}

// GetExportLocation determina la mejor ubicación para exportar
func GetExportLocation() (string, bool) {
	usbPaths := findUSBMounts()

	if len(usbPaths) > 0 {
		return usbPaths[0], true
	}

	// Si no hay USB, usar directorio actual
	currentDir, err := os.Getwd()
	if err != nil {
		return ".", false
	}

	return currentDir, false
}

// FormatExportMessage genera un mensaje formateado sobre la exportación
func FormatExportMessage(filepath string, isUSB bool) string {
	var sb strings.Builder

	sb.WriteString("\n┌─ EXPORTACIÓN ────────────────────────────────────────────────┐\n")

	if isUSB {
		sb.WriteString("│ ✓ Archivo exportado a dispositivo USB\n")
	} else {
		sb.WriteString("│ ✓ Archivo exportado al directorio actual\n")
	}

	sb.WriteString(fmt.Sprintf("│ Ruta: %s\n", filepath))
	sb.WriteString("└──────────────────────────────────────────────────────────────┘\n")

	return sb.String()
}
