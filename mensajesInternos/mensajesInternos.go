package mensajesInternos

import (
	"github.com/gorilla/websocket"
)

/*
	MensajePartida es la interfaz que deben implementar todos los mensajes que
	recibe una partida.
*/
type MensajePartida interface{}

/*
	LlegadaUsuario indica que un usuario quiere unirse a la partida, cuál es su
	id y cuál es la conexión con él.
*/
type LlegadaUsuario struct {
	IdUsuario int
	Ws        *websocket.Conn
}

/*
	SalidaUsuario indica que un usuario se ha desconectado de la partida y cuál
	es su id.
*/
type SalidaUsuario struct {
	IdUsuario int
}

/*
	InicioPartida indica que un usuario ha solicitado que comience la partida y
	cuál es su id.
*/
type InicioPartida struct {
	IdUsuario int
}

/*
	InvitacionPartida indica que un usuario quiere invitar a otro a la partida,
	y los ids de ambos.
*/
type InvitacionPartida struct {
	IdCreador, IdInvitado int
}

/*
	AccionPartida indica que un usuario quiere realizar una acción en la partida,
	cuál es su id y los necesarios para llevarla a cabo.
*/
type AccionPartida struct {
	IdUsuario                                            int
	Origen, Objetivo, NumTropasOrigen, NumTropasObjetivo int
	Tipo                                                 string
}

/*
	FinPartida indica que la partida ha terminado.
*/
type FinPartida struct{}

/*
	MensajeInvalido indica que se ha recibido un mensaje incorrecto, el id del
	emisor y una explicación del error.
*/
type MensajeInvalido struct {
	IdUsuario int
	Err       string
}
