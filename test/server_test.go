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
