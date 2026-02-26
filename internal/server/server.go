package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Lexharden/hwscan/internal/hardware"
	"github.com/Lexharden/hwscan/internal/utils"
)

// Server representa el servidor HTTP embebido
type Server struct {
	hardwareInfo *hardware.HardwareInfo
	port         int
}

// New crea una nueva instancia del servidor
func New(info *hardware.HardwareInfo, port int) *Server {
	return &Server{
		hardwareInfo: info,
		port:         port,
	}
}

// Start inicia el servidor HTTP en modo background
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Endpoint API para obtener información de hardware
	mux.HandleFunc("/api/hardware", s.handleHardwareAPI)

	// Endpoint de salud
	mux.HandleFunc("/api/health", s.handleHealth)

	// Servir archivos estáticos desde el directorio web/
	// Busca en múltiples ubicaciones: ./web (desarrollo) y /usr/share/hwscan/web (producción)
	webDir := "web"
	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		webDir = "/usr/share/hwscan/web"
	}
	fs := http.FileServer(http.Dir(webDir))
	mux.Handle("/", fs)

	addr := fmt.Sprintf(":%d", s.port)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Obtener IP local para mostrar al usuario
	localIP := utils.GetLocalIP()
	if localIP != "" {
		log.Printf("Servidor web iniciado en http://%s:%d\n", localIP, s.port)
	} else {
		log.Printf("Servidor web iniciado en http://0.0.0.0:%d\n", s.port)
	}

	// Iniciar servidor en goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Error en servidor HTTP: %v\n", err)
		}
	}()

	return nil
}

// handleHardwareAPI maneja las peticiones al endpoint /api/hardware
func (s *Server) handleHardwareAPI(w http.ResponseWriter, r *http.Request) {
	// Solo permitir GET
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Configurar headers CORS (por si se accede desde otro origen)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// Serializar y enviar la información de hardware
	if err := json.NewEncoder(w).Encode(s.hardwareInfo); err != nil {
		http.Error(w, "Error al serializar datos", http.StatusInternalServerError)
		log.Printf("Error al serializar hardware info: %v\n", err)
		return
	}
}

// handleHealth maneja las peticiones al endpoint /api/health
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "hwscan",
		"version":   "1.0",
	}

	json.NewEncoder(w).Encode(response)
}
