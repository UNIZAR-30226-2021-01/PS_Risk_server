package test

import (
	"PS_Risk_server/baseDatos"
	"PS_Risk_server/mensajes"
	"crypto/sha256"
	"encoding/json"
	"testing"
)

// NewSHA256 ...
func NewSHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
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

func TestEnviarSolicitudAmistad(t *testing.T) {
	bd := baseDatos.NuevaBDConexionLocal(baseDeDatos, false)
	defer bd.Cerrar()
	obtenido := bd.EnviarSolicitudAmistad(2, nombreAmigo2, claveTest)
	if c, _ := coincideErrorNulo(obtenido); !c {
		t.Errorf("Error %q, al enviar solicitud de amistad", obtenido)
	}
	obtenido = bd.ObtenerNotificaciones(3, claveTest)
	esperado := mensajes.JsonData{
		"notificaciones": []mensajes.JsonData{
			mensajes.NotificacionJson(2, "Peticion de amistad", nombreAmigo1),
		},
	}
	if !comprobarJson(obtenido, esperado) {
		t.Errorf("ObtenerNotificaciones() = %q, se esperaba %q", obtenido, esperado)
	}
}

func TestAceptarSolicitudAmistad(t *testing.T) {
	//nuevoNombre := "nuevo Nombre"
	//nuevoCorreo := "nuevoCorreo@mail.com"
	/*bd := baseDatos.NuevaBDConexionLocal(baseDeDatos, true)
	// Necesario usuario 1 con clave claveTest
	bd.CrearCuenta(nombreTest, correoTest, claveTest, recibeCorreosTest)*/
	bd := baseDatos.NuevaBDConexionLocal(baseDeDatos, false)
	defer bd.Cerrar()
	// Necesario usuario 2 con clave claveTest
	//bd.CrearCuenta(nuevoNombre, nuevoCorreo, claveTest, recibeCorreosTest)
	// Necesaria notificación de solicitud de amistad enviada por 2 a 1
	//bd.EnviarSolicitudAmistad(1, 2, claveTest)
	obtenido := bd.AceptarSolicitudAmistad(3, 2, claveTest)
	if c, _ := coincideErrorNulo(obtenido); !c {
		t.Errorf("Error %q, al aceptar solicitud de amistad", obtenido)
		return
	}
	esperado := mensajes.JsonData{
		"amigos": []mensajes.JsonData{
			mensajes.AmigoJson(3, 0, 0, nombreAmigo2),
		},
	}
	obtenido = bd.ObtenerAmigos(2, claveTest)
	if !comprobarJson(obtenido, esperado) {
		t.Errorf("ObtenerAmigos() = %q, se esperaba %q", obtenido, esperado)
		return
	}
	esperado = mensajes.JsonData{
		"amigos": []mensajes.JsonData{
			mensajes.AmigoJson(2, 0, 0, nombreAmigo1),
		},
	}
	obtenido = bd.ObtenerAmigos(3, claveTest)
	if !comprobarJson(obtenido, esperado) {
		t.Errorf("ObtenerAmigos() = %q, se esperaba %q", obtenido, esperado)
		return
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
	obtenido := bd.EliminarAmigo(3, 2, claveTest)
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

	return comprobarJson(respuestaObtenida, esperado), esperado
}

func coincideErrorNulo(respuestaObtenida mensajes.JsonData) (bool, mensajes.JsonData) {
	esperado := mensajes.ErrorJson("", baseDatos.NoError)
	return comprobarJson(respuestaObtenida, esperado), esperado
}

func comprobarJson(a, b mensajes.JsonData) bool {
	aByte, _ := json.Marshal(a)
	bByte, _ := json.Marshal(b)
	return string(aByte) == string(bByte)
}
