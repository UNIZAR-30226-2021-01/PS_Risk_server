package test

import (
	"PS_Risk_server/baseDatos"
	"PS_Risk_server/server"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

// NewSHA256 ...
func NewSHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func TestHello(t *testing.T) {
	esperado := "Hello, world."
	if obtenido := server.Hello(); obtenido != esperado {
		t.Errorf("Hello() = %q\nEsperado %q", obtenido, esperado)
	}
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
	bd := baseDatos.NuevaBDConexionLocal("postgres", true)
	idEsperado := bd.LeerMaxIdUsuario() + 1
	fmt.Println("id esperado:", idEsperado)
	obtenido := bd.CrearCuenta(nombre, correo, clave, recibeCorreos)
	coincide, esperado := coincideCrearUsuario(nombre, correo, clave,
		recibeCorreos, idEsperado, obtenido)
	if !coincide {
		t.Errorf("CrearCuenta() = %q, se esperaba %q", obtenido, esperado)
	}
}

func coincideCrearUsuario(nombre, correo, clave string,
	recibeCorreos bool, idEsperado int,
	respuestaObtenida map[string]interface{}) (bool, map[string]interface{}) {
	incorrecto := false
	campoError := "error"
	campoCodigo := "code"
	campoUsuario := "usuario"
	campoNombre := "nombre"
	campoCorreo := "correo"
	campoClave := "clave"
	campoRecibe := "recibeCorreos"
	campoId := "id"

	e := respuestaObtenida[campoError]
	if e == nil {
		incorrecto = true
	} else {
		switch errorObtenido := e.(type) {
		case map[string]interface{}:
			if errorObtenido[campoCodigo] == nil ||
				errorObtenido[campoCodigo] != baseDatos.NoError {
				incorrecto = true
				fmt.Println("Error en el código de error")
			}
		default:
			incorrecto = true
			fmt.Println("Error en el campo error")
		}
	}

	if !incorrecto {
		u := respuestaObtenida["usuario"]
		if u == nil {
			incorrecto = true
		} else {
			switch usuarioObtenido := u.(type) {
			case map[string]interface{}:
				if usuarioObtenido[campoId] == nil ||
					usuarioObtenido[campoId] != idEsperado {
					incorrecto = true
					fmt.Println("Error en el id_usuario")
				} else if usuarioObtenido[campoNombre] == nil ||
					usuarioObtenido[campoNombre] != nombre {
					incorrecto = true
					fmt.Println("Error en el nombre de usuario")
				} else if usuarioObtenido[campoCorreo] == nil ||
					usuarioObtenido[campoCorreo] != correo {
					incorrecto = true
					fmt.Println("Error en el correo")
				} else if usuarioObtenido[campoClave] == nil ||
					usuarioObtenido[campoClave] != clave {
					incorrecto = true
					fmt.Println("Error en la clave")
				} else if usuarioObtenido[campoRecibe] == nil ||
					usuarioObtenido[campoRecibe] != recibeCorreos {
					incorrecto = true
					fmt.Println("Error en el campo recibeCorreos")
				}
			default:
				incorrecto = true
				fmt.Println("Error en el campo usuario")
			}
		}
	}

	if !incorrecto {
		iconos := respuestaObtenida["iconos"]
		aspectos := respuestaObtenida["aspectos"]
		tiendaIconos := respuestaObtenida["tiendaIconos"]
		tiendaAspectos := respuestaObtenida["tiendaAspectos"]
		if iconos == nil || aspectos == nil || tiendaIconos == nil ||
			tiendaAspectos == nil {
			incorrecto = true
			fmt.Println("Error con las listas de iconos o aspectos")
		}
	}

	var esperado map[string]interface{}
	esperado = nil

	if incorrecto {
		esperado = make(map[string]interface{})
		errorEsperado := make(map[string]interface{})
		errorEsperado[campoCodigo] = baseDatos.NoError
		errorEsperado["err"] = ""
		esperado[campoError] = errorEsperado
		usuarioEsperado := make(map[string]interface{})
		usuarioEsperado[campoId] = idEsperado
		usuarioEsperado[campoNombre] = nombre
		usuarioEsperado[campoCorreo] = correo
		usuarioEsperado[campoClave] = clave
		usuarioEsperado[campoRecibe] = recibeCorreos
		esperado[campoUsuario] = usuarioEsperado
	}

	return !incorrecto, esperado
}
