package test

import (
	"PS_Risk_server/baseDatos"
	"PS_Risk_server/mensajes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

const (
	baseDeDatos = "..."
)

// NewSHA256 ...
func NewSHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func TestCrearCuenta(t *testing.T) {
	nombre := "usuario0"
	correo := "correo@mail.com"
	h := []byte("clave")
	result := NewSHA256(h)
	clave := "6d5074b4bf2b913866157d7674f1eda042c5c614876de876f7512702d2572a06"
	if hex.EncodeToString(result) != clave {
		t.Errorf("want %v; got %v", clave, hex.EncodeToString(result))
	}
	recibeCorreos := true
	bd := baseDatos.NuevaBDConexionLocal(baseDeDatos, true)
	idEsperado := bd.LeerMaxIdUsuario() + 1
	fmt.Println("id esperado:", idEsperado)
	obtenido := bd.CrearCuenta(nombre, correo, clave, recibeCorreos)
	fmt.Println(obtenido)
	coincide, esperado := coincideCrearUsuario(nombre, correo, clave,
		recibeCorreos, idEsperado, obtenido["usuario"].(mensajes.JsonData))
	if !coincide {
		t.Errorf("CrearCuenta() = %q, se esperaba %q", obtenido, esperado)
	}
}

func coincideCrearUsuario(nombre, correo, clave string,
	recibeCorreos bool, idEsperado int,
	respuestaObtenida mensajes.JsonData) (bool, map[string]interface{}) {

	esperado := mensajes.UsuarioJson(idEsperado, 1, 1, 0, nombre, correo, clave, recibeCorreos)

	err := respuestaObtenida["err"]
	if err != nil {
		return false, esperado
	}

	if respuestaObtenida["id"].(int) != esperado["id"].(int) ||
		respuestaObtenida["nombre"].(string) != esperado["nombre"].(string) ||
		respuestaObtenida["correo"].(string) != esperado["correo"].(string) ||
		respuestaObtenida["clave"].(string) != esperado["clave"].(string) ||
		respuestaObtenida["recibeCorreos"].(bool) != esperado["recibeCorreos"].(bool) ||
		respuestaObtenida["icono"].(int) != esperado["icono"].(int) ||
		respuestaObtenida["aspecto"].(int) != esperado["aspecto"].(int) ||
		respuestaObtenida["riskos"].(int) != esperado["riskos"].(int) {
		return false, esperado
	}

	return true, esperado
}
