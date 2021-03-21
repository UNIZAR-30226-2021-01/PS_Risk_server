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
	baseDeDatos       = "postgres://ehoyoqkpcgqyrc:c0f640db53a47820e660f19d987b013d7b01d9ac217b669f971070926969dfe0@ec2-54-72-155-238.eu-west-1.compute.amazonaws.com:5432/d3cnglhgpo3rgk"
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
