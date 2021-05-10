/*
	El paquete mensajesInternos proporciona tipos de datos para enviar la
	información que se recibe de los usuarios conectados a una sala o partida
	desde la goroutine que recibe los mensajes a la que gestiona la sala o partida.
*/
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
	IdUsuario       int
	Ws              *websocket.Conn
	RecibirMensajes chan bool
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
	MensajeFase indica que un usuario quiere avanzar la fase de turno en la que
	se encuentra.
*/
type MensajeFase struct {
	IdUsuario int
}

/*
	MensajeRefuerzos indica cuántas tropas de refuerzo quiere colocar un jugador
	y en qué territorio.
*/
type MensajeRefuerzos struct {
	IdUsuario    int
	IdTerritorio int
	Tropas       int
}

/*
	MensajeAtaque indica desde qué territorio ataca un jugador, con cuántas
	tropas y a qué territorio.
*/
type MensajeAtaque struct {
	IdUsuario           int
	IdTerritorioOrigen  int
	IdTerritorioDestino int
	Tropas              int
}

/*
	MensajeMover indica cuántas tropas mueve un jugador, desde qué territorio y
	hasta cuál.
*/
type MensajeMover struct {
	IdUsuario           int
	IdTerritorioOrigen  int
	IdTerritorioDestino int
	Tropas              int
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

/*
	CuentaEliminada indica que se ha borrado la cuenta de usuario de un jugador.
*/
type CuentaEliminada struct {
	IdUsuario int
}
