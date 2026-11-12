package main

import (
	"encoding/json" // Nos permite convertir datos de Go a JSON y viceversa
	"fmt"           // Nos permite imprimir mensajes en la consola
	"log"           // Nos ayuda a manejar errores fatales (como si el puerto está ocupado)
	"net/http"      // El paquete estándar de Go para crear servidores web
	"os"            // Nos permite leer y escribir archivos en el disco duro
	"strconv"       // Nos permite convertir strings a números (útil para leer IDs)
	"sync"          // Nos permite usar "Mutex" para evitar choques si hay dos peticiones al mismo tiempo
)

// ==========================================
// 1. ESTRUCTURAS DE DATOS
// ==========================================

// Producto define cómo se ve un producto en nuestro sistema.
// Los "tags" (ej: `json:"id"`) le dicen a Go cómo llamar a este campo en el JSON.
type Producto struct {
	ID     int     `json:"id"`
	Nombre string  `json:"nombre"`
	Precio float64 `json:"precio"`
	Stock  int     `json:"stock"`
}

// ==========================================
// 2. VARIABLES GLOBALES
// ==========================================

var (
	// mu es un "Mutex" (candado). Sirve para que si dos usuarios intentan
	// guardar un producto al mismo tiempo, no dañen el archivo JSON.
	mu sync.Mutex

	// Ruta donde está guardado nuestro archivo JSON.
	filePath = "data/productos.json"
)

// ==========================================
// 3. FUNCIONES DE LECTURA Y ESCRITURA DE DATOS
// ==========================================

// leerProductos abre el archivo JSON y lo convierte en una lista de productos de Go.
func leerProductos() ([]Producto, error) {
	var productos []Producto // Creamos una lista vacía

	// Leemos todo el contenido del archivo
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err // Si el archivo no existe o hay error, lo devolvemos
	}

	// Convertimos el texto JSON del archivo en nuestra lista de Go
	err = json.Unmarshal(file, &productos)
	return productos, err
}

// guardarProductos toma la lista de productos y la escribe en el archivo JSON.
func guardarProductos(productos []Producto) error {
	// Convertimos la lista de Go a texto JSON con formato bonito (indentado)
	data, err := json.MarshalIndent(productos, "", "  ")
	if err != nil {
		return err
	}

	// Escribimos el texto en el archivo.
	// 0644 son los permisos (lectura/escritura para ti, solo lectura para otros).
	return os.WriteFile(filePath, data, 0644)
}

// ==========================================
// 4. CONFIGURACIÓN DEL SERVIDOR Y CORS
// ==========================================

// enableCORS permite que nuestro futuro Frontend (en otro puerto) pueda hacer peticiones a esta API.
// Sin esto, los navegadores bloquean las peticiones por seguridad.
func enableCORS(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*") // Permite cualquier origen
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// productosHandler maneja las peticiones a la ruta general "/productos"
func productosHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)

	// Si el navegador envía una petición de verificación (OPTIONS), le decimos que sí y cortamos.
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Bloqueamos el candado para que esta petición tenga acceso exclusivo al archivo
	mu.Lock()
	defer mu.Unlock() // "defer" asegura que el candado se libere al terminar esta función

	switch r.Method {
	case "GET":
		// --- LEER TODOS LOS PRODUCTOS ---
		productos, err := leerProductos()
		if err != nil {
			http.Error(w, "Error al leer productos", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(productos) // Enviamos la lista como JSON

	case "POST":
		// --- CREAR UN NUEVO PRODUCTO ---
		var nuevoProducto Producto

		// Leemos el JSON que viene en el cuerpo de la petición
		if err := json.NewDecoder(r.Body).Decode(&nuevoProducto); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		productos, _ := leerProductos()

		// Generamos un ID automático (buscamos el más alto y le sumamos 1)
		maxID := 100
		for _, p := range productos {
			if p.ID > maxID {
				maxID = p.ID
			}
		}
		nuevoProducto.ID = maxID + 1

		// Agregamos el producto a la lista y guardamos
		productos = append(productos, nuevoProducto)
		guardarProductos(productos)

		// Respondemos con el producto creado y el código 201 (Creado)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(nuevoProducto)

	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

// productoPorIDHandler maneja peticiones a rutas con ID, como "/productos/101"
func productoPorIDHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Extraemos el ID de la URL.
	// r.URL.Path es algo como "/productos/101", así que cortamos el texto para quedarnos solo con "101".
	idStr := r.URL.Path[len("/productos/"):]
	id, err := strconv.Atoi(idStr) // Convertimos "101" (string) a 101 (número)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	productos, _ := leerProductos()
	index := -1 // -1 significa que no lo hemos encontrado aún

	// Buscamos el producto en la lista
	for i, p := range productos {
		if p.ID == id {
			index = i // Guardamos su posición en la lista
			break
		}
	}

	// Si no lo encontramos, devolvemos error 404
	if index == -1 {
		http.Error(w, "Producto no encontrado", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		// --- OBTENER UN SOLO PRODUCTO ---
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(productos[index])

	case "PUT":
		// --- ACTUALIZAR UN PRODUCTO ---
		var productoActualizado Producto
		if err := json.NewDecoder(r.Body).Decode(&productoActualizado); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		// Nos aseguramos de que el ID no se cambie por accidente
		productoActualizado.ID = id

		// Reemplazamos el producto viejo por el actualizado en la lista
		productos[index] = productoActualizado
		guardarProductos(productos)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(productoActualizado)

	case "DELETE":
		// --- ELIMINAR UN PRODUCTO ---
		// En Go, para eliminar un elemento de una lista, unimos la parte izquierda y derecha del índice
		productos = append(productos[:index], productos[index+1:]...)
		guardarProductos(productos)

		// 204 significa "Éxito, pero no hay contenido que devolver"
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

// ==========================================
// 5. FUNCIÓN PRINCIPAL
// ==========================================

func main() {
	// Asociamos la URL "/productos" a su función
	http.HandleFunc("/productos", productosHandler)
	// Asociamos la URL "/productos/" (con slash al final) para las rutas con ID
	http.HandleFunc("/productos/", productoPorIDHandler)

	fmt.Println("🛒 Servicio de Productos corriendo en http://localhost:8082")

	// Encendemos el servidor en el puerto 8082. Si falla, el programa se cierra (log.Fatal).
	log.Fatal(http.ListenAndServe(":8082", nil))
}
