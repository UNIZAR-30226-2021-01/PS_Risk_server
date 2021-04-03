package mensajesInternos

import (
	"github.com/gorilla/websocket"
)

type MensajePartida interface{}

type LlegadaUsuario struct {
	IdUsuario int
	//U  baseDatos.Usuario
	Ws *websocket.Conn
}

type InicioPartida struct {
	IdUsuario int
}

type InvitacionPartida struct {
	IdCreador, IdInvitado int
}

type AccionPartida struct {
	IdUsuario, Origen, Objetivo, NumTropasOrigen, NumTropasObjetivo int
	Tipo                                                            string
}

type FinPartida struct{}

/*
func AccionDesdeJson(accion JsonData, idUsuario int) (AccionPartida, error) {
	var respuesta AccionPartida
	origen, ok := accion["origen"].(int)
	if !ok {
		return respuesta, errors.New("el territorio de origen debe ser un entero")
	}
	objetivo, ok := accion["objetivo"].(int)
	if !ok {
		return respuesta, errors.New("el territorio objetivo debe ser un entero")
	}
}*/
