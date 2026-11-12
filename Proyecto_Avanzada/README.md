# Sistema de Facturación Digital

**Realizado por:** John Galeano y Esteban Martinez

## Descripción del proyecto

Este proyecto consiste en un sistema de facturación digital basado en una arquitectura de microservicios. El backend está desarrollado en Go y el frontend utiliza JavaScript, HTML y Bootstrap.

El sistema permite gestionar usuarios, clientes y productos, además de generar facturas en formato JSON para su visualización e impresión en PDF.

## Arquitectura del sistema

El sistema está compuesto por cuatro microservicios independientes que se comunican mediante HTTP REST. El frontend consume los servicios para realizar las diferentes operaciones.

Los servicios y sus respectivos puertos son:

- Auth Service: puerto 8081. Gestiona el registro y el inicio de sesión de los usuarios.
- Products Service: puerto 8082. Gestiona la información de los productos.
- Customers Service: puerto 8083. Gestiona la información de los clientes.
- Billing Service: puerto 8084. Genera y consulta las facturas.

El frontend se ejecuta desde el puerto 5500 cuando se utiliza Live Server.

La información de cada servicio se almacena en archivos JSON:

- Auth Service: `usuarios.json`
- Products Service: `productos.json`
- Customers Service: `clientes.json`
- Billing Service: `facturas.json`

## Endpoints

### 1. Servicio de autenticación (puerto 8081)

- `POST /register`: registra un nuevo usuario. La contraseña se encripta utilizando bcrypt.
- `POST /login`: valida las credenciales del usuario y devuelve sus datos sin incluir la contraseña.

### 2. Servicio de productos (puerto 8082)

- `GET /productos`: lista todos los productos.
- `GET /productos/{id}`: obtiene un producto específico.
- `POST /productos`: crea un nuevo producto.
- `PUT /productos/{id}`: actualiza un producto.
- `DELETE /productos/{id}`: elimina un producto.

### 3. Servicio de clientes (puerto 8083)

- `GET /clientes`: lista todos los clientes.
- `GET /clientes/{id}`: obtiene un cliente específico.
- `POST /clientes`: crea un nuevo cliente.
- `PUT /clientes/{id}`: actualiza un cliente.
- `DELETE /clientes/{id}`: elimina un cliente.

### 4. Servicio de facturación (puerto 8084)

- `POST /facturas`: recibe el ID del cliente, el ID del usuario y una lista de productos con sus cantidades. Consulta los otros microservicios para obtener la información necesaria y calcular los subtotales y el total de la factura.
- `GET /facturas`: muestra el historial de facturas generadas.

## Flujo de generación de una factura

1. El usuario inicia sesión desde el frontend. La solicitud se envía al Auth Service.
2. El usuario selecciona un cliente. El frontend consulta la información al Customers Service.
3. El usuario selecciona los productos y las cantidades. El frontend consulta al Products Service.
4. Al seleccionar la opción para generar la factura, el frontend envía los datos al Billing Service.
5. El Billing Service consulta la información de los productos en el puerto 8082 y los datos del cliente en el puerto 8083.
6. El servicio realiza los cálculos, guarda la factura en `facturas.json` y devuelve el resultado en formato JSON al navegador.

## Ejemplo de factura generada

La factura contiene la información del cliente, del usuario que realizó la atención y el detalle de los productos incluidos.

```json
{
  "factura_id": 5001,
  "cliente": {
    "id": 201,
    "nombre": "Carlos Pérez",
    "cc": "1234567",
    "email": "carlos@example.com",
    "direccion": "Calle 123 #45-67, Bogotá"
  },
  "usuario": {
    "id": 1,
    "nombre": "John Galeano",
    "email": "john@example.com",
    "rol": "admin"
  },
  "productos": [
    {
      "id": 101,
      "nombre": "Laptop Lenovo ThinkPad",
      "cantidad": 1,
      "precio_unitario": 3500000,
      "subtotal": 3500000
    },
    {
      "id": 102,
      "nombre": "Mouse inalámbrico Logitech",
      "cantidad": 2,
      "precio_unitario": 12000,
      "subtotal": 24000
    }
  ],
  "total": 3524000,
  "fecha": "2026-11-12"
}
```

## Pruebas

Para probar los endpoints sin utilizar el frontend, se puede instalar la extensión REST Client en Visual Studio Code y utilizar el archivo `pruebas.http` incluido en el proyecto. Desde este archivo se pueden ejecutar las solicitudes a los diferentes servicios.

## Cómo ejecutar el proyecto localmente

Para ejecutar el proyecto es necesario tener instalado Go y Visual Studio Code.

1. Clonar el repositorio.
2. Abrir el proyecto en Visual Studio Code.
3. Abrir cuatro terminales diferentes.
4. En la primera terminal, ejecutar el servicio de autenticación:

   ```bash
   cd auth-service
   go run main.go
   ```

5. En la segunda terminal, ejecutar el servicio de productos:

   ```bash
   cd products-service
   go run main.go
   ```

6. En la tercera terminal, ejecutar el servicio de clientes:

   ```bash
   cd customers-service
   go run main.go
   ```

7. En la cuarta terminal, ejecutar el servicio de facturación:

   ```bash
   cd billing-service
   go run main.go
   ```

8. Finalmente, abrir el archivo `frontend/index.html` en el navegador. También se puede utilizar la extensión Live Server de Visual Studio Code.

## Autores

Proyecto realizado por John Galeano y Esteban Martinez.