// HWSCAN - Hardware Detection Tool
// Developed by: Yafel Garcia
// Copyright (c) 2026

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Lexharden/hwscan/internal/export"
	"github.com/Lexharden/hwscan/internal/hardware"
	"github.com/Lexharden/hwscan/internal/server"
	"github.com/Lexharden/hwscan/internal/version"
)

func main() {
	// Flags de línea de comandos
	portFlag := flag.Int("port", 8080, "Puerto para el servidor web")
	noServerFlag := flag.Bool("no-server", false, "Desactivar servidor web")
	noExportFlag := flag.Bool("no-export", false, "Desactivar exportación automática")
	outputFlag := flag.String("output", "", "Ruta específica para exportar JSON")
	versionFlag := flag.Bool("version", false, "Mostrar versión")
	helpFlag := flag.Bool("help", false, "Mostrar ayuda")

	flag.Parse()

	// Mostrar versión
	if *versionFlag {
		fmt.Printf("HWSCAN v%s\n", version.Current)
		os.Exit(0)
	}

	// Mostrar ayuda
	if *helpFlag {
		showHelp()
		os.Exit(0)
	}

	// Paso 1: Detectar hardware
	fmt.Println("Detectando hardware del sistema...")
	fmt.Println()

	hwInfo, err := hardware.Detect()
	if err != nil {
		log.Fatalf("Error al detectar hardware: %v\n", err)
	}

	// Paso 2: Mostrar información en consola
	fmt.Print(hardware.FormatConsole(hwInfo))
	fmt.Println()

	// Paso 3: Exportar a JSON
	if !*noExportFlag {
		var exportPath string
		var isUSB bool

		if *outputFlag != "" {
			// Usar ruta especificada
			exportPath = *outputFlag
			isUSB = false
		} else {
			// Exportación automática
			exportPath, err = export.AutoExport(hwInfo)
			if err != nil {
				log.Printf("Advertencia: no se pudo exportar JSON: %v\n", err)
			} else {
				_, locationIsUSB := export.GetExportLocation()
				isUSB = locationIsUSB && exportPath != ""
			}
		}

		if exportPath != "" {
			fmt.Print(export.FormatExportMessage(exportPath, isUSB))
			fmt.Println()
		}
	}

	// Paso 4: Iniciar servidor web (si no está desactivado)
	if !*noServerFlag {
		srv := server.New(hwInfo, *portFlag)
		if err := srv.Start(); err != nil {
			log.Printf("Advertencia: no se pudo iniciar servidor web: %v\n", err)
		} else {
			fmt.Printf("Servidor web: http://localhost:%d\n", *portFlag)
			fmt.Println()
		}
	}

	// Paso 5: Mantener el programa ejecutándose
	fmt.Println("HWSCAN está ejecutándose. Presione Ctrl+C para salir.")
	fmt.Println()

	// Esperar señal de interrupción
	waitForShutdown()
}

// showHelp muestra la ayuda del programa
func showHelp() {
	help := `
HWSCAN - Hardware Detection Tool

USO:
    hwscan [opciones]

OPCIONES:
    -port <número>      Puerto para el servidor web (default: 8080)
    -no-server          Desactivar servidor web
    -no-export          Desactivar exportación automática a JSON
    -output <ruta>      Ruta específica para exportar JSON
    -version            Mostrar versión del programa
    -help               Mostrar esta ayuda

EJEMPLOS:
    # Ejecutar con configuración por defecto
    hwscan

    # Cambiar puerto del servidor web
    hwscan -port 9090

    # Solo detección sin servidor web
    hwscan -no-server

    # Exportar a ruta específica
    hwscan -output /tmp/mi-hardware.json

    # Solo mostrar en consola
    hwscan -no-server -no-export

DESCRIPCIÓN:
    HWSCAN detecta automáticamente el hardware del sistema incluyendo:
    - CPU (modelo, velocidad, núcleos)
    - Memoria RAM (capacidad, módulos, velocidades)
	- Disco(s) (modelo, capacidad, tipo)
    - Placa Madre (fabricante, modelo, BIOS)
    - GPU (tarjetas gráficas instaladas)

    La información se muestra en consola, se exporta a JSON y está
    disponible mediante una interfaz web en http://localhost:8080

INTEGRACIÓN:
    Este programa está diseñado para ejecutarse en Alpine Linux
    como parte de un sistema booteable de diagnóstico de hardware.

AUTOR:
    Yafel Garcia (Lexharden)
    Copyright © 2026

Para más información visite: https://github.com/Lexharden/Hwscan
`
	fmt.Println(help)
}

// waitForShutdown espera una señal de interrupción para cerrar el programa
func waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Bloquear hasta recibir señal
	<-sigChan

	fmt.Println()
	fmt.Println("Recibida señal de terminación. Cerrando HWSCAN...")

	// Dar tiempo para cerrar conexiones
	time.Sleep(500 * time.Millisecond)

	fmt.Println("¡Hasta pronto!")
}
