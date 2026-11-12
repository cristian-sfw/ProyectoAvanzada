// Variables globales
let usuarioLogueado = null;
let productosFactura = [];
let total = 0;

// ==========================================
// 0. FUNCIONES DE INTERFAZ
// ==========================================
function toggleForms() {
    const loginForm = document.getElementById('login-form');
    const registerForm = document.getElementById('register-form');
    
    // Usamos classList de Bootstrap (d-none)
    if (loginForm.classList.contains('d-none')) {
        loginForm.classList.remove('d-none');
        registerForm.classList.add('d-none');
    } else {
        loginForm.classList.add('d-none');
        registerForm.classList.remove('d-none');
    }
}

function cerrarSesion() {
    location.reload();
}

// ==========================================
// 1. LOGIN Y REGISTRO (Auth Service)
// ==========================================
async function login() {
    const email = document.getElementById('login-email').value.trim().toLowerCase();
    const password = document.getElementById('login-password').value.trim(); 
    const msg = document.getElementById('login-msg');

    try {
        const response = await fetch('http://localhost:8081/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });

        if (response.ok) {
            usuarioLogueado = await response.json();
            
            // Ocultar pantalla de login y mostrar la app
            document.getElementById('welcome-screen').classList.add('d-none');
            document.getElementById('app-section').classList.remove('d-none');
            
            document.getElementById('user-info').innerText = `👤 ${usuarioLogueado.nombre} (${usuarioLogueado.rol})`;
            
            cargarClientes();
            cargarProductos();
            cargarHistorial();
        } else {
            const errorMsg = await response.text();
            msg.innerHTML = `<span class="text-danger">Error: ${errorMsg}</span>`;
        }
    } catch (error) {
        msg.innerHTML = `<span class="text-danger">Error al conectar con Auth Service</span>`;
    }
}

async function registrar() {
    const nombre = document.getElementById('reg-nombre').value.trim();
    const email = document.getElementById('reg-email').value.trim().toLowerCase();
    const password = document.getElementById('reg-password').value.trim(); 
    const rol = document.getElementById('reg-rol').value;
    const msg = document.getElementById('reg-msg');

    if (!nombre || !email || !password) {
        return msg.innerHTML = `<span class="text-danger">Todos los campos son obligatorios</span>`;
    }

    try {
        const response = await fetch('http://localhost:8081/register', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ nombre, email, password, rol })
        });

        if (response.ok) {
            msg.innerHTML = `<span class="text-success">¡Usuario creado! Ya puedes iniciar sesión.</span>`;
            document.getElementById('reg-nombre').value = '';
            document.getElementById('reg-email').value = '';
            document.getElementById('reg-password').value = '';
            
            setTimeout(() => {
                toggleForms();
                msg.innerHTML = '';
            }, 2000);
        } else {
            const errorText = await response.text();
            msg.innerHTML = `<span class="text-danger">${errorText}</span>`;
        }
    } catch (error) {
        msg.innerHTML = `<span class="text-danger">Error al conectar con Auth Service</span>`;
    }
}

// ==========================================
// 2. CARGAR CLIENTES Y PRODUCTOS
// ==========================================
async function cargarClientes() {
    try {
        const response = await fetch('http://localhost:8083/clientes');
        const clientes = await response.json();
        
        const select = document.getElementById('cliente-select');
        select.innerHTML = '<option value="">Seleccione un cliente...</option>';
        
        clientes.forEach(c => {
            select.innerHTML += `<option value="${c.id}">${c.nombre} (CC: ${c.cc})</option>`;
        });
    } catch (error) {
        alert("Error al cargar clientes. Revisa el puerto 8083.");
    }
}

async function cargarProductos() {
    try {
        const response = await fetch('http://localhost:8082/productos');
        const productos = await response.json();
        
        const select = document.getElementById('producto-select');
        select.innerHTML = '<option value="">Seleccione un producto...</option>';
        
        productos.forEach(p => {
            select.innerHTML += `<option value="${p.id}">${p.nombre} - $${p.precio.toLocaleString()}</option>`;
        });
    } catch (error) {
        alert("Error al cargar productos. Revisa el puerto 8082.");
    }
}

// ==========================================
// 3. AGREGAR PRODUCTOS A LA TABLA
// ==========================================
async function agregarProducto() {
    const prodId = parseInt(document.getElementById('producto-select').value);
    const cantidad = parseInt(document.getElementById('cantidad-input').value);

    if (!prodId || cantidad < 1) return alert("Seleccione producto y cantidad válida");

    try {
        const response = await fetch(`http://localhost:8082/productos/${prodId}`);
        const producto = await response.json();

        const subtotal = producto.precio * cantidad;
        
        productosFactura.push({
            id: producto.id,
            nombre: producto.nombre,
            cantidad: cantidad,
            precio_unitario: producto.precio,
            subtotal: subtotal
        });

        total += subtotal;
        actualizarTabla();
        
        document.getElementById('cantidad-input').value = 1;
    } catch (error) {
        alert("Error al obtener datos del producto");
    }
}

function actualizarTabla() {
    const tbody = document.querySelector('#tabla-factura tbody');
    tbody.innerHTML = '';
    
    productosFactura.forEach(p => {
        tbody.innerHTML += `
            <tr>
                <td>${p.nombre}</td>
                <td>$${p.precio_unitario.toLocaleString()}</td>
                <td>${p.cantidad}</td>
                <td>$${p.subtotal.toLocaleString()}</td>
            </tr>
        `;
    });

    document.getElementById('total-factura').innerText = total.toLocaleString();
}

// ==========================================
// 4. GENERAR LA FACTURA VISUAL
// ==========================================
async function generarFactura() {
    const clienteId = document.getElementById('cliente-select').value;
    if (!clienteId) return alert("Debe seleccionar un cliente");
    if (productosFactura.length === 0) return alert("Debe agregar al menos un producto");

    const data = {
        cliente_id: parseInt(clienteId),
        usuario_id: usuarioLogueado.id,
        productos: productosFactura.map(p => ({ id: p.id, cantidad: p.cantidad }))
    };

    try {
        const response = await fetch('http://localhost:8084/facturas', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });

        if (response.ok) {
            const factura = await response.json();
            
            let filasProductos = '';
            factura.productos.forEach(p => {
                filasProductos += `
                    <tr>
                        <td>${p.nombre}</td>
                        <td class="text-center">${p.cantidad}</td>
                        <td class="text-end">$${p.precio_unitario.toLocaleString()}</td>
                        <td class="text-end">$${p.subtotal.toLocaleString()}</td>
                    </tr>
                `;
            });

            const htmlFactura = `
                <div class="row mb-4">
                    <div class="col-8">
                        <h2 class="text-primary fw-bold">FACTURA DE VENTA</h2>
                        <p class="mb-0"><strong>Factura N°:</strong> ${factura.factura_id}</p>
                        <p class="mb-0"><strong>Fecha:</strong> ${factura.fecha}</p>
                    </div>
                    <div class="col-4 text-end">
                        <h4 class="text-dark fw-bold">🏢 John y Esteban</h4>
                        <p class="mb-0">NIT: 900.123.456-7</p>
                        <p class="mb-0">Bogotá, Colombia</p>
                    </div>
                </div>
                <hr>
                <div class="row mb-4">
                    <div class="col-6">
                        <h6 class="text-muted fw-bold">DATOS DEL CLIENTE</h6>
                        <p class="mb-0"><strong>Nombre:</strong> ${factura.cliente.nombre}</p>
                        <p class="mb-0"><strong>CC:</strong> ${factura.cliente.cc}</p>
                        <p class="mb-0"><strong>Email:</strong> ${factura.cliente.email}</p>
                        <p class="mb-0"><strong>Dirección:</strong> ${factura.cliente.direccion}</p>
                    </div>
                    <div class="col-6 text-end">
                        <h6 class="text-muted fw-bold">ATENDIDO POR</h6>
                        <p class="mb-0"><strong>Usuario:</strong> ${factura.usuario.nombre}</p>
                        <p class="mb-0"><strong>Rol:</strong> ${factura.usuario.rol}</p>
                    </div>
                </div>
                <table class="table table-bordered mt-4">
                    <thead class="table-dark">
                        <tr>
                            <th>Producto</th>
                            <th class="text-center">Cant.</th>
                            <th class="text-end">Precio Unit.</th>
                            <th class="text-end">Subtotal</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${filasProductos}
                    </tbody>
                </table>
                <div class="row mt-4">
                    <div class="col-8"></div>
                    <div class="col-4 text-end">
                        <h3 class="text-success fw-bold">TOTAL: $${factura.total.toLocaleString()}</h3>
                    </div>
                </div>
                <div class="text-center mt-5 text-muted">
                    <small>¡Gracias por su compra!</small>
                </div>
            `;

            document.getElementById('factura-html').innerHTML = htmlFactura;
            const modal = new bootstrap.Modal(document.getElementById('facturaModal'));
            modal.show();
            
            productosFactura = [];
            total = 0;
            actualizarTabla();
            cargarHistorial(); 
            
        } else {
            alert("Error al generar factura");
        }
    } catch (error) {
        alert("Error al conectar con Billing Service.");
    }
}

// ==========================================
// 5. CARGAR HISTORIAL DE FACTURAS
// ==========================================
async function cargarHistorial() {
    try {
        const response = await fetch('http://localhost:8084/facturas');
        const facturas = await response.json();
        
        const tbody = document.querySelector('#tabla-historial tbody');
        tbody.innerHTML = '';
        
        if(facturas.length === 0) {
            tbody.innerHTML = '<tr><td colspan="4" class="text-center text-muted">No hay facturas generadas aún.</td></tr>';
            return;
        }

        facturas.reverse().forEach(f => {
            tbody.innerHTML += `
                <tr>
                    <td><strong>#${f.factura_id}</strong></td>
                    <td>${f.fecha}</td>
                    <td>${f.cliente.nombre}</td>
                    <td class="text-success fw-bold">$${f.total.toLocaleString()}</td>
                </tr>
            `;
        });
    } catch (error) {
        console.error("Error al cargar historial:", error);
    }
}