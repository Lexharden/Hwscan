// Package utils proporciona funciones utilitarias comunes para el proyecto
package utils

import (
	"net"
)

// GetLocalIP obtiene la dirección IP local de la máquina
// Retorna la primera IP no-loopback encontrada
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		// Verificar si es una dirección IP (no una red)
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			// Solo retornar IPv4
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return "localhost"
}
