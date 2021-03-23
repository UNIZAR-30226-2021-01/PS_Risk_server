package test

import (
	"PS_Risk_server/baseDatos"
	"PS_Risk_server/mensajes"
	"crypto/sha256"
	"encoding/json"
	"testing"
)

const (
	// Pon aqui la url de la base de datos de Heroku
	baseDeDatos       = "..."
	nombreTest        = "usuario0"
	correoTest        = "correo@mail.com"
	claveTest         = "6d5074b4bf2b913866157d7674f1eda042c5c614876de876f7512702d2572a06"
	recibeCorreosTest = true
)

// NewSHA256 ...
func NewSHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func TestCrearCuenta(t *testing.T) {
	bd := baseDatos.NuevaBDConexionLocal(baseDeDatos, true)
	obtenido := bd.CrearCuenta(nombreTest, correoTest, claveTest, recibeCorreosTest)
	coincide, esperado := coincideUsuario(nombreTest, correoTest, claveTest, recibeCorreosTest, 1, obtenido)
	if !coincide {
		t.Errorf("CrearCuenta() = %q, se esperaba %q", obtenido, esperado)
	}
	bd.Cerrar()
}

func TestActualizarUsuario(t *testing.T) {
	bd := baseDatos.NuevaBDConexionLocal(baseDeDatos, false)
	obtenido := bd.ObtenerUsuario(1, claveTest)
	coincide, esperado := coincideUsuario(nombreTest, correoTest, claveTest, recibeCorreosTest, 1, obtenido)
	if !coincide {
		t.Errorf("CrearCuenta() = %q, se esperaba %q", obtenido, esperado)
	}
	bd.Cerrar()
}

func TestModificarUsuario(t *testing.T) {
	nuevoNombre := "nuevo Nombre"
	nuevoCorreo := "nuevoCorreo@mail.com"
	nuevaClave := "82ef67bb06675af2d43639806236ad1189253ce86e6210c072a3a265987df429"

	bd := baseDatos.NuevaBDConexionLocal(baseDeDatos, false)

	obtenido := bd.ModificarUsuario(1, claveTest, "clave", nuevaClave)
	if obtenido["code"].(int) != 0 {
		t.Errorf("Error obtenido modificando usuario, %q", obtenido["err"])
	}
	obtenido = bd.ModificarUsuario(1, nuevaClave, "nombre", nuevoNombre)
	if obtenido["code"].(int) != 0 {
		t.Errorf("Error obtenido modificando usuario, %q", obtenido["err"])
	}
	obtenido = bd.ModificarUsuario(1, nuevaClave, "correo", nuevoCorreo)
	if obtenido["code"].(int) != 0 {
		t.Errorf("Error obtenido modificando usuario, %q", obtenido["err"])
	}

	obtenido = bd.ObtenerUsuario(1, nuevaClave)
	coincide, esperado := coincideUsuario(nuevoNombre, nuevoCorreo, nuevaClave, recibeCorreosTest, 1, obtenido)
	if !coincide {
		t.Errorf("CrearCuenta() = %q, se esperaba %q", obtenido, esperado)
	}
	bd.Cerrar()
}

func TestIniciarSesionNombre(t *testing.T) {
	bd := baseDatos.NuevaBDConexionLocal(baseDeDatos, false)
	obtenido := bd.IniciarSesionNombre(nombreTest, claveTest)
	coincide, esperado := coincideUsuario(nombreTest, correoTest, claveTest, recibeCorreosTest, 1, obtenido)
	if !coincide {
		t.Errorf("CrearCuenta() = %q, se esperaba %q", obtenido, esperado)
	}
	bd.Cerrar()
}

func TestIniciarSesionCorreo(t *testing.T) {
	bd := baseDatos.NuevaBDConexionLocal(baseDeDatos, false)
	obtenido := bd.IniciarSesionCorreo(correoTest, claveTest)
	coincide, esperado := coincideUsuario(nombreTest, correoTest, claveTest, recibeCorreosTest, 1, obtenido)
	if !coincide {
		t.Errorf("CrearCuenta() = %q, se esperaba %q", obtenido, esperado)
	}
	bd.Cerrar()
}

func TestAceptarSolicitudAmistad(t *testing.T) {
	//nuevoNombre := "nuevo Nombre"
	//nuevoCorreo := "nuevoCorreo@mail.com"
	/*bd := baseDatos.NuevaBDConexionLocal(baseDeDatos, true)
	// Necesario usuario 1 con clave claveTest
	bd.CrearCuenta(nombreTest, correoTest, claveTest, recibeCorreosTest)*/
	bd := baseDatos.NuevaBDConexionLocal(baseDeDatos, false)
	// Necesario usuario 2 con clave claveTest
	//bd.CrearCuenta(nuevoNombre, nuevoCorreo, claveTest, recibeCorreosTest)
	// Necesaria notificación de solicitud de amistad enviada por 2 a 1
	//bd.EnviarSolicitudAmistad(1, 2, claveTest)
	obtenido := bd.AceptarSolicitudAmistad(1, 2, claveTest)
	coincide, esperado := coincideErrorNulo(obtenido)
	if !coincide {
		t.Errorf("AceptarSolicitudAmistad() = %q, se esperaba %q", obtenido, esperado)
	}
}

func TestRechazarSolicitudAmistad(t *testing.T) {
	// Necesario usuario 1 con clave claveTest
	// Necesario usuario 2
	// Necesaria notificación de solicitud de amistad enviada por 2 a 1
	bd := baseDatos.NuevaBDConexionLocal(baseDeDatos, false)
	obtenido := bd.RechazarSolicitudAmistad(1, 2, claveTest)
	coincide, esperado := coincideErrorNulo(obtenido)
	if !coincide {
		t.Errorf("RechazarSolicitudAmistad() = %q, se esperaba %q", obtenido, esperado)
	}
}

func TestEliminarAmigo(t *testing.T) {
	// Necesario usuario 1
	// Necesario usuario 2 con clave claveTest
	// Necesario que sean amigos
	bd := baseDatos.NuevaBDConexionLocal(baseDeDatos, false)
	obtenido := bd.EliminarAmigo(2, 1, claveTest)
	coincide, esperado := coincideErrorNulo(obtenido)
	if !coincide {
		t.Errorf("EliminarAmigo() = %q, se esperaba %q", obtenido, esperado)
	}
}

func coincideUsuario(nombre, correo, clave string,
	recibeCorreos bool, idEsperado int,
	respuestaObtenida mensajes.JsonData) (bool, mensajes.JsonData) {

	usuarioEsperado := mensajes.UsuarioJson(idEsperado, 0, 0, 0, nombre, correo, clave, recibeCorreos)
	cosmeticosEsperado := []mensajes.JsonData{mensajes.CosmeticoJson(0, 0)}
	esperado := mensajes.JsonData{
		"usuario":        usuarioEsperado,
		"iconos":         cosmeticosEsperado,
		"aspectos":       cosmeticosEsperado,
		"tiendaIconos":   cosmeticosEsperado,
		"tiendaAspectos": cosmeticosEsperado,
	}

	err := respuestaObtenida["err"]
	if err != nil {
		return false, esperado
	}

	usuarioObtenido, _ := json.Marshal(respuestaObtenida)
	jsonEsperado, _ := json.Marshal(esperado)

	if string(usuarioObtenido) != string(jsonEsperado) {
		return false, esperado
	}

	return true, esperado
}

func coincideErrorNulo(respuestaObtenida mensajes.JsonData) (bool, mensajes.JsonData) {
	esperado := mensajes.ErrorJson("", baseDatos.NoError)
	err := respuestaObtenida["err"]
	if err == nil || err != "" {
		return false, esperado
	}
	codigoError := respuestaObtenida["code"]
	if codigoError == nil || codigoError != baseDatos.NoError {
		return false, esperado
	}
	return true, esperado
}
