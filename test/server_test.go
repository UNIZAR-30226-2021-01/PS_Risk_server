package test

import (
	"PS_Risk_server/baseDatos"
	"PS_Risk_server/mensajes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"
)

const (
	// Pon aqui la url de la base de datos de Heroku
	baseDeDatos   = "..."
	nombre        = "usuario0"
	correo        = "correo@mail.com"
	clave         = "6d5074b4bf2b913866157d7674f1eda042c5c614876de876f7512702d2572a06"
	recibeCorreos = true
)

// NewSHA256 ...
func NewSHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func TestCrearCuenta(t *testing.T) {
	bd := baseDatos.NuevaBDConexionLocal(baseDeDatos, true)
	idEsperado := bd.LeerMaxIdUsuario() + 1
	fmt.Println("id esperado:", idEsperado)
	obtenido := bd.CrearCuenta(nombre, correo, clave, recibeCorreos)
	coincide, esperado := coincideUsuario(nombre, correo, clave,
		recibeCorreos, idEsperado, obtenido)
	if !coincide {
		t.Errorf("CrearCuenta() = %q, se esperaba %q", obtenido, esperado)
	}
}

func TestActualizarUsuario(t *testing.T) {
	bd := baseDatos.NuevaBDConexionLocal(baseDeDatos, false)
	obtenido := bd.ObtenerUsuario(1, clave)
	coincide, esperado := coincideUsuario(nombre, correo, clave,
		recibeCorreos, 1, obtenido)
	if !coincide {
		t.Errorf("CrearCuenta() = %q, se esperaba %q", obtenido, esperado)
	}
}

func coincideUsuario(nombre, correo, clave string,
	recibeCorreos bool, idEsperado int,
	respuestaObtenida mensajes.JsonData) (bool, map[string]interface{}) {

	usuarioEsperado := mensajes.UsuarioJson(idEsperado, 1, 1, 0, nombre, correo, clave, recibeCorreos)
	cosmeticosEsperado := []mensajes.JsonData{mensajes.CosmeticoJson(1, 0)}
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
