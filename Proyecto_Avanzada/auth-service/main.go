package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type Usuario struct {
	ID       int    `json:"id"`
	Nombre   string `json:"nombre"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Rol      string `json:"rol"`
}

type Credenciales struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var (
	mu       sync.Mutex
	filePath = "data/usuarios.json"
)

func leerUsuarios() ([]Usuario, error) {
	var usuarios []Usuario
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(file, &usuarios)
	return usuarios, err
}

func guardarUsuarios(usuarios []Usuario) error {
	data, err := json.MarshalIndent(usuarios, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

func enableCORS(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
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

	var nuevoUsuario Usuario
	if err := json.NewDecoder(r.Body).Decode(&nuevoUsuario); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	nuevoUsuario.Email = strings.ToLower(nuevoUsuario.Email)

	usuarios, err := leerUsuarios()
	if err != nil {
		usuarios = []Usuario{}
	}

	for _, u := range usuarios {
		if strings.EqualFold(u.Email, nuevoUsuario.Email) {
			http.Error(w, "El correo ya está registrado", http.StatusConflict)
			return
		}
	}

	fmt.Println("REGISTRANDO:", nuevoUsuario.Email)

	hash, err := bcrypt.GenerateFromPassword([]byte(nuevoUsuario.Password), 10)
	if err != nil {
		http.Error(w, "Error al encriptar", http.StatusInternalServerError)
		return
	}
	nuevoUsuario.Password = string(hash)

	maxID := 0
	for _, u := range usuarios {
		if u.ID > maxID {
			maxID = u.ID
		}
	}
	nuevoUsuario.ID = maxID + 1

	if nuevoUsuario.Rol == "" {
		nuevoUsuario.Rol = "cliente"
	}

	usuarios = append(usuarios, nuevoUsuario)
	guardarUsuarios(usuarios)

	nuevoUsuario.Password = ""
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(nuevoUsuario)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var creds Credenciales
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	usuarios, err := leerUsuarios()
	if err != nil {
		http.Error(w, "Error al leer DB", http.StatusInternalServerError)
		return
	}

	var usuarioEncontrado Usuario
	encontrado := false

	for _, u := range usuarios {
		if strings.EqualFold(u.Email, creds.Email) {
			usuarioEncontrado = u
			encontrado = true
			break
		}
	}

	if !encontrado {
		http.Error(w, "Usuario no encontrado", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(usuarioEncontrado.Password), []byte(creds.Password))
	if err != nil {
		http.Error(w, "Contraseña incorrecta", http.StatusUnauthorized)
		return
	}

	usuarioEncontrado.Password = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usuarioEncontrado)
}

func main() {
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)

	fmt.Println("🔐 Auth Service en http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
