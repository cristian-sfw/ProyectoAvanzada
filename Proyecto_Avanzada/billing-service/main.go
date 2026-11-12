package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// ==========================================
// 1. ESTRUCTURAS DE DATOS
// ==========================================

type Cliente struct {
	ID        int    `json:"id"`
	Nombre    string `json:"nombre"`
	CC        string `json:"cc"`
	Email     string `json:"email"`
	Direccion string `json:"direccion"`
}

type Producto struct {
	ID     int     `json:"id"`
	Nombre string  `json:"nombre"`
	Precio float64 `json:"precio"`
	Stock  int     `json:"stock"`
}

type Usuario struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
	Email  string `json:"email"`
	Rol    string `json:"rol"`
}

type ItemFactura struct {
	ID             int     `json:"id"`
	Nombre         string  `json:"nombre"`
	Cantidad       int     `json:"cantidad"`
	PrecioUnitario float64 `json:"precio_unitario"`
	Subtotal       float64 `json:"subtotal"`
}

type Factura struct {
	FacturaID int           `json:"factura_id"`
	Cliente   Cliente       `json:"cliente"`
	Usuario   Usuario       `json:"usuario"`
	Productos []ItemFactura `json:"productos"`
	Total     float64       `json:"total"`
	Fecha     string        `json:"fecha"`
}

type PeticionFactura struct {
	ClienteID int `json:"cliente_id"`
	UsuarioID int `json:"usuario_id"`
	Productos []struct {
		ID       int `json:"id"`
		Cantidad int `json:"cantidad"`
	} `json:"productos"`
}

// ==========================================
// 2. VARIABLES GLOBALES
// ==========================================

var (
	mu       sync.Mutex
	filePath = "data/facturas.json"
)

// ==========================================
// 3. FUNCIONES AUXILIARES
// ==========================================

func leerFacturas() ([]Factura, error) {
	var facturas []Factura
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(file, &facturas)
	return facturas, err
}

func guardarFacturas(facturas []Factura) error {
	data, err := json.MarshalIndent(facturas, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

func enableCORS(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// ==========================================
// 4. MANEJADORES HTTP
// ==========================================

func generarFacturaHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	var req PeticionFactura
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// 1. Consultar Cliente (Puerto 8083)
	respCliente, err := http.Get(fmt.Sprintf("http://localhost:8083/clientes/%d", req.ClienteID))
	if err != nil || respCliente.StatusCode != 200 {
		http.Error(w, "Error al obtener cliente o cliente no existe", http.StatusNotFound)
		return
	}
	defer respCliente.Body.Close()

	var cliente Cliente
	bodyCliente, _ := io.ReadAll(respCliente.Body)
	json.Unmarshal(bodyCliente, &cliente)

	// 2. Usuario (Simulado para no modificar Auth)
	usuario := Usuario{
		ID:     req.UsuarioID,
		Nombre: "Ricardo Gómez",
		Email:  "ricardo@example.com",
		Rol:    "admin",
	}

	// 3. Consultar Productos (Puerto 8082)
	var productosFactura []ItemFactura
	var total float64 = 0

	for _, pReq := range req.Productos {
		respProd, err := http.Get(fmt.Sprintf("http://localhost:8082/productos/%d", pReq.ID))
		if err != nil || respProd.StatusCode != 200 {
			http.Error(w, fmt.Sprintf("Producto con ID %d no encontrado", pReq.ID), http.StatusNotFound)
			return
		}

		var producto Producto
		bodyProd, _ := io.ReadAll(respProd.Body)
		json.Unmarshal(bodyProd, &producto)
		respProd.Body.Close()

		if producto.Stock < pReq.Cantidad {
			http.Error(w, fmt.Sprintf("No hay suficiente stock para: %s", producto.Nombre), http.StatusBadRequest)
			return
		}

		subtotal := float64(pReq.Cantidad) * producto.Precio
		total += subtotal

		item := ItemFactura{
			ID:             producto.ID,
			Nombre:         producto.Nombre,
			Cantidad:       pReq.Cantidad,
			PrecioUnitario: producto.Precio,
			Subtotal:       subtotal,
		}
		productosFactura = append(productosFactura, item)
	}

	// 4. Armar Factura y guardar
	facturas, _ := leerFacturas()
	maxID := 5000
	for _, f := range facturas {
		if f.FacturaID > maxID {
			maxID = f.FacturaID
		}
	}

	nuevaFactura := Factura{
		FacturaID: maxID + 1,
		Cliente:   cliente,
		Usuario:   usuario,
		Productos: productosFactura,
		Total:     total,
		Fecha:     time.Now().Format("2006-01-02"),
	}

	facturas = append(facturas, nuevaFactura)
	guardarFacturas(facturas)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(nuevaFactura)
}

func listarFacturasHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	if r.Method != "GET" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	facturas, err := leerFacturas()
	if err != nil {
		http.Error(w, "Error al leer facturas", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(facturas)
}

// ==========================================
// 5. FUNCIÓN PRINCIPAL
// ==========================================

func main() {
	http.HandleFunc("/facturas", func(w http.ResponseWriter, r *http.Request) {
		// --- SOLUCIÓN CORS ---
		// Si el navegador envía la petición de reconocimiento (OPTIONS), le damos luz verde.
		if r.Method == "OPTIONS" {
			enableCORS(&w)
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == "POST" {
			generarFacturaHandler(w, r)
		} else if r.Method == "GET" {
			listarFacturasHandler(w, r)
		} else {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("🧾 Servicio de Facturación corriendo en http://localhost:8084")
	log.Fatal(http.ListenAndServe(":8084", nil))
}
