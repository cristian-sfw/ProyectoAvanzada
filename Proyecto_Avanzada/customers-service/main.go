package main

import (
	"encoding/json" // Para convertir de JSON a Go y viceversa
	"fmt"           // Para imprimir mensajes en la consola
	"log"           // Para manejar errores fatales
	"net/http"      // Para crear el servidor web
	"os"            // Para leer y escribir archivos
	"strconv"       // Para convertir strings a números
	"sync"          // Para usar Mutex y evitar choques de concurrencia
)

// ==========================================
// 1. ESTRUCTURAS DE DATOS
// ==========================================

// Cliente define cómo se ve un cliente en nuestro sistema.
// Observa que incluimos el campo "CC" con su tag `json:"cc"`.
type Cliente struct {
	ID        int    `json:"id"`
	Nombre    string `json:"nombre"`
	CC        string `json:"cc"`
	Email     string `json:"email"`
	Direccion string `json:"direccion"`
}

// ==========================================
// 2. VARIABLES GLOBALES
// ==========================================

var (
	mu       sync.Mutex // Candado para evitar choques al escribir el archivo
	filePath = "data/clientes.json"
)

// ==========================================
// 3. FUNCIONES DE LECTURA Y ESCRITURA
// ==========================================

// leerClientes lee el archivo JSON y lo convierte en una lista de clientes.
func leerClientes() ([]Cliente, error) {
	var clientes []Cliente
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(file, &clientes)
	return clientes, err
}

// guardarClientes toma la lista y la escribe en el archivo JSON.
func guardarClientes(clientes []Cliente) error {
	data, err := json.MarshalIndent(clientes, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

// enableCORS permite que el Frontend pueda hacer peticiones a esta API.
func enableCORS(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// ==========================================
// 4. MANEJADORES HTTP (ENDPOINTS)
// ==========================================

// clientesHandler maneja la ruta general "/clientes"
func clientesHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	switch r.Method {
	case "GET":
		// --- LEER TODOS LOS CLIENTES ---
		clientes, err := leerClientes()
		if err != nil {
			http.Error(w, "Error al leer clientes", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clientes)

	case "POST":
		// --- CREAR UN NUEVO CLIENTE ---
		var nuevoCliente Cliente
		if err := json.NewDecoder(r.Body).Decode(&nuevoCliente); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		clientes, _ := leerClientes()

		// Generamos un ID automático (buscamos el máximo y sumamos 1)
		maxID := 200
		for _, c := range clientes {
			if c.ID > maxID {
				maxID = c.ID
			}
		}
		nuevoCliente.ID = maxID + 1

		clientes = append(clientes, nuevoCliente)
		guardarClientes(clientes)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(nuevoCliente)

	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

// clientePorIDHandler maneja rutas como "/clientes/201"
func clientePorIDHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Extraemos el ID de la URL
	idStr := r.URL.Path[len("/clientes/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	clientes, _ := leerClientes()
	index := -1

	// Buscamos el cliente
	for i, c := range clientes {
		if c.ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		http.Error(w, "Cliente no encontrado", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		// --- OBTENER UN SOLO CLIENTE ---
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clientes[index])

	case "PUT":
		// --- ACTUALIZAR UN CLIENTE ---
		var clienteActualizado Cliente
		if err := json.NewDecoder(r.Body).Decode(&clienteActualizado); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		clienteActualizado.ID = id // Mantenemos el ID original
		clientes[index] = clienteActualizado
		guardarClientes(clientes)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clienteActualizado)

	case "DELETE":
		// --- ELIMINAR UN CLIENTE ---
		clientes = append(clientes[:index], clientes[index+1:]...)
		guardarClientes(clientes)
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

// ==========================================
// 5. FUNCIÓN PRINCIPAL
// ==========================================

func main() {
	http.HandleFunc("/clientes", clientesHandler)
	http.HandleFunc("/clientes/", clientePorIDHandler)

	fmt.Println("👥 Servicio de Clientes corriendo en http://localhost:8083")
	log.Fatal(http.ListenAndServe(":8083", nil))
}
