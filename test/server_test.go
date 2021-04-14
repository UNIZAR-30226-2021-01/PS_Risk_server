package test

import (
	"PS_Risk_server/mensajes"
	"crypto/sha256"
	"encoding/json"
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
	esperado := mensajes.ErrorJson("", mensajes.NoError)
	return comprobarJson(respuestaObtenida, esperado), esperado
}

func comprobarJson(a, b mensajes.JsonData) bool {
	aByte, _ := json.Marshal(a)
	bByte, _ := json.Marshal(b)
	return string(aByte) == string(bByte)
}
