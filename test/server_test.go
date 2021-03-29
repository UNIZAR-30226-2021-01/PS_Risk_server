package test

import (
	"PS_Risk_server/baseDatos"
	"PS_Risk_server/mensajes"
	"crypto/sha256"
	"encoding/json"
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

func comprobarJson(a, b mensajes.JsonData) bool {
	aByte, _ := json.Marshal(a)
	bByte, _ := json.Marshal(b)
	return string(aByte) == string(bByte)
}
