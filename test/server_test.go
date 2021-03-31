package test

import (
	"PS_Risk_server/baseDatos"
	"PS_Risk_server/mensajes"
	"crypto/sha256"
	"encoding/json"
)

const (
	baseDeDatos       = "postgres://ehoyoqkpcgqyrc:c0f640db53a47820e660f19d987b013d7b01d9ac217b669f971070926969dfe0@ec2-54-72-155-238.eu-west-1.compute.amazonaws.com:5432/d3cnglhgpo3rgk"
	nombreTest        = "usuario0"
	correoTest        = "correo0@mail.com"
	nombreAmigo1      = "nombre1"
	correoAmigo1      = "correo1@mail.com"
	nombreAmigo2      = "nombre2"
	correoAmigo2      = "correo2@mail.com"
	claveTest         = "6d5074b4bf2b913866157d7674f1eda042c5c614876de876f7512702d2572a06"
	recibeCorreosTest = true
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
