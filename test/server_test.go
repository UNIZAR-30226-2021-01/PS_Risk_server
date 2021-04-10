package test

import (
	"PS_Risk_server/baseDatos"
	"PS_Risk_server/mensajes"
	"crypto/sha256"
	"encoding/json"
	"log"
)

const (
	baseDeDatos       = "..."
	nombreTest        = "usuario0"
	correoTest        = "correo0@mail.com"
	nombreAmigo1      = "nombre1"
	correoAmigo1      = "correo1@mail.com"
	nombreAmigo2      = "nombre2"
	correoAmigo2      = "correo2@mail.com"
	claveTest         = "6d5074b4bf2b913866157d7674f1eda042c5c614876de876f7512702d2572a06"
	recibeCorreosTest = true
	nombreSalaTest    = "partida0"
	tiempoTurnoTest   = 30
)

// NewSHA256 ...
func NewSHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func coincideErrorNulo(respuestaObtenida mensajes.JsonData) (bool, mensajes.JsonData) {
	esperado := mensajes.ErrorJson("", baseDatos.NoError)
	return comprobarJson(respuestaObtenida, esperado), esperado
}

func comprobarDatosPartida(a, b mensajes.JsonData, importaOrdenJugadores bool) bool {
	if importaOrdenJugadores {
		return comprobarJson(a, b)
	}
	ja, ok := a["jugadores"]
	if !ok {
		log.Println("no hay jugadores en a")
		return false
	}
	ta, ok := a["listaTerritorios"]
	if !ok {
		log.Println("no hay territorios en a")
		return false
	}
	jb, ok := b["jugadores"]
	if !ok {
		log.Println("no hay jugadores en b")
		return false
	}
	tb, ok := b["listaTerritorios"]
	if !ok {
		log.Println("no hay territorios en b")
		return false
	}
	if len(ja.([]mensajes.JsonData)) != len(jb.([]mensajes.JsonData)) ||
		len(ta.([]mensajes.JsonData)) != len(tb.([]mensajes.JsonData)) {
		log.Println("la longitud de los vectores no coincide:")
		log.Println("\tja:", len(ja.([]mensajes.JsonData)))
		log.Println("\tjb:", len(jb.([]mensajes.JsonData)))
		log.Println("\tta:", len(ta.([]mensajes.JsonData)))
		log.Println("\ttb:", len(tb.([]mensajes.JsonData)))
		return false
	}
	c := a
	delete(c, "jugadores")
	delete(c, "listaTerritorios")
	d := b
	delete(d, "jugadores")
	delete(d, "listaTerritorios")
	return comprobarJson(c, d)
}

func comprobarJson(a, b mensajes.JsonData) bool {
	aByte, _ := json.Marshal(a)
	bByte, _ := json.Marshal(b)
	return string(aByte) == string(bByte)
}
