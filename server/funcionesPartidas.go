package server

import (
	"PS_Risk_server/baseDatos"
	"PS_Risk_server/mensajes"
	"PS_Risk_server/mensajesInternos"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
)

type mensajeCreacion struct {
	IdUsuario   int    `json:"idUsuario"`
	Clave       string `json:"clave"`
	TiempoTurno int    `json:"tiempoTurno"`
	NombreSala  string `json:"nombreSala"`
}

/*
	crearPartidaHandler maneja las conexiones de tipo websocket para crear partidas.
*/
func (s *Servidor) crearPartidaHandler(w http.ResponseWriter, r *http.Request) {
	var mensajeInicial mensajeCreacion

	// Crear conexión websocket
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Upgrader err: %v\n", err)
		return
	}
	defer ws.Close()

	// Leer los datos de partida y usuario enviados
	err = ws.ReadJSON(&mensajeInicial)
	if err != nil {
		devolverErrorWebsocket(mensajes.ErrorPeticion, err.Error(), ws)
		return
	}

	// Comprobar si el usuario está registrado
	u, err := s.UsuarioDAO.ObtenerUsuario(mensajeInicial.IdUsuario, mensajeInicial.Clave)
	if err != nil {
		devolverErrorWebsocket(mensajes.ErrorUsuario, err.Error(), ws)
		return
	}

	// Crea la partida en la base de datos
	p, err := s.PartidasDAO.CrearPartida(u, mensajeInicial.TiempoTurno,
		mensajeInicial.NombreSala, ws)
	if err != nil {
		devolverErrorWebsocket(mensajes.ErrorPeticion, err.Error(), ws)
		return
	}

	// Guarda la referencia a la partida en el servidor
	s.Partidas.Store(p.IdPartida, p)

	// Envía al usuario los datos de la partida en formato json
	aux := mensajes.JsonData{}
	mapstructure.Decode(p, &aux)
	aux["_tipoMensaje"] = "d"
	p.Enviar(u.Id, aux)

	// Lanza una rutina para atender a la sala
	go s.atenderSala(p)

	s.recibirMensajesUsuario(u, ws, p)
}

type mensajeEntrarSala struct {
	IdUsuario int    `json:"idUsuario"`
	Clave     string `json:"clave"`
	IdSala    int    `json:"idSala"`
}

/*
	aceptarSalaHandler maneja las conexiones de tipo websocket para entrar a salas.
*/
func (s *Servidor) aceptarSalaHandler(w http.ResponseWriter, r *http.Request) {
	var mensajeInicial mensajeEntrarSala

	// Crear conexión websocket
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Upgrader err: %v\n", err)
		return
	}
	defer ws.Close()

	// Leer los datos de partida y usuario enviados
	err = ws.ReadJSON(&mensajeInicial)
	if err != nil {
		devolverErrorWebsocket(mensajes.ErrorPeticion, err.Error(), ws)
		return
	}

	// Comprobar si el usuario está registrado
	u, err := s.UsuarioDAO.ObtenerUsuario(mensajeInicial.IdUsuario,
		mensajeInicial.Clave)
	if err != nil {
		devolverErrorWebsocket(mensajes.ErrorUsuario, err.Error(), ws)
		return
	}

	// Carga la referencia a la partida guardada en el servidor
	p, ok := s.Partidas.Load(mensajeInicial.IdSala)
	if !ok {
		devolverErrorWebsocket(mensajes.ErrorPeticion, "Partida no encontrada", ws)
		return
	}

	// Notifica a la partida de que un jugador nuevo ha entrado
	p.(*baseDatos.Partida).Mensajes <- mensajesInternos.LlegadaUsuario{
		IdUsuario: u.Id,
		Ws:        ws,
	}
	s.recibirMensajesUsuario(u, ws, p.(*baseDatos.Partida))
}

type mensajeSala struct {
	Tipo                string `json:"tipo"`
	IdInvitado          int    `json:"idInvitado,omitempty"`
	IdTerritorio        int    `json:"id,omitempty"`
	IdTerritorioOrigen  int    `json:"origen,omitempty"`
	IdTerritorioDestino int    `json:"destino,omitempty"`
	Tropas              int    `json:"tropas,omitempty"`
}

/*
	recibirMensajesUsuario recibe los mensajes que un usuario envía a través de su
	websocket. Comprueba si los datos enviados son correctos y notifica de eventos
	a la partida.
*/
func (s *Servidor) recibirMensajesUsuario(u baseDatos.Usuario,
	ws *websocket.Conn, p *baseDatos.Partida) {

	var mensajeRecibido mensajeSala

	// Bucle infinito para leer mensajes
	for {
		// Leer mensaje
		err := ws.ReadJSON(&mensajeRecibido)

		// Si da error leer el mensaje se desconecta al usuario
		if err != nil {
			// Notificar a la sala de la desconexión del usuario
			p.Mensajes <- mensajesInternos.SalidaUsuario{
				IdUsuario: u.Id,
			}
			return
		}

		// Enviar el mensaje a la función que gestiona la partida
		if strings.EqualFold(mensajeRecibido.Tipo, "Iniciar") {
			p.Mensajes <- mensajesInternos.InicioPartida{
				IdUsuario: u.Id,
			}
		} else if strings.EqualFold(mensajeRecibido.Tipo, "Invitar") {
			p.Mensajes <- mensajesInternos.InvitacionPartida{
				IdCreador:  u.Id,
				IdInvitado: mensajeRecibido.IdInvitado,
			}
		} else if strings.EqualFold(mensajeRecibido.Tipo, "Fase") {
			p.Mensajes <- mensajesInternos.MensajeFase{
				IdUsuario: u.Id,
			}
		} else if strings.EqualFold(mensajeRecibido.Tipo, "Refuerzos") {
			p.Mensajes <- mensajesInternos.MensajeRefuerzos{
				IdUsuario:    u.Id,
				IdTerritorio: mensajeRecibido.IdTerritorio,
				Tropas:       mensajeRecibido.Tropas,
			}
		} else if strings.EqualFold(mensajeRecibido.Tipo, "Ataque") {
			p.Mensajes <- mensajesInternos.MensajeAtaque{
				IdUsuario:           u.Id,
				IdTerritorioOrigen:  mensajeRecibido.IdTerritorioOrigen,
				IdTerritorioDestino: mensajeRecibido.IdTerritorioDestino,
				Tropas:              mensajeRecibido.Tropas,
			}
		} else if strings.EqualFold(mensajeRecibido.Tipo, "Movimiento") {
			p.Mensajes <- mensajesInternos.MensajeMover{
				IdUsuario:           u.Id,
				IdTerritorioOrigen:  mensajeRecibido.IdTerritorioOrigen,
				IdTerritorioDestino: mensajeRecibido.IdTerritorioDestino,
				Tropas:              mensajeRecibido.Tropas,
			}
		} else {
			p.Mensajes <- mensajesInternos.MensajeInvalido{
				IdUsuario: u.Id,
				Err:       "Mensaje erróneo",
			}
		}
	}
}

/*
	atenderSala recibe las diferentes notificaciones de una sala y ejecuta las
	acciones que se requieran para cada notificación.
*/
func (s *Servidor) atenderSala(p *baseDatos.Partida) {
	// Bucle infinito para leer notificaciones
	for {
		mensajeRecibido := <-p.Mensajes

		switch mt := mensajeRecibido.(type) {
		case mensajesInternos.LlegadaUsuario:
			// Cargar los datos del usuario de la base de datos
			u, err := s.UsuarioDAO.ObtenerUsuarioId(mt.IdUsuario)
			if err != nil {
				devolverErrorWebsocket(mensajes.ErrorUsuario, err.Error(), mt.Ws)
			} else {
				// Añadir el usuario a la partida
				mensajeEnviar := s.PartidasDAO.EntrarPartida(p, u, mt.Ws)
				if _, hayError := mensajeEnviar["err"]; hayError {
					mt.Ws.WriteJSON(mensajeEnviar)
					mt.Ws.Close()
				} else {
					// Si no hay error enviar a todos los jugadores de la sala
					// el nuevo usuario
					p.EnviarATodos(mensajeEnviar)
				}
			}
		case mensajesInternos.InvitacionPartida:
			// Invitar usuario a la partida
			err := s.PartidasDAO.InvitarPartida(p, mt.IdCreador, mt.IdInvitado)
			if err != nil {
				p.EnviarError(mt.IdCreador, mensajes.ErrorPeticion, err.Error())
			}
		case mensajesInternos.InicioPartida:
			// Iniciar la partida
			mensajeEnviar := s.PartidasDAO.IniciarPartida(p, mt.IdUsuario)
			if _, hayError := mensajeEnviar["err"]; hayError {
				p.Enviar(mt.IdUsuario, mensajeEnviar)
			} else {
				// Enviar a todos los jugadores de la sala los datos de inicio
				p.EnviarATodos(mensajeEnviar)
				// Función que gestiona una partida empezada
				s.atenderPartida(p)
				// Borrar la partida de la estructura del servidor
				s.Partidas.Delete(p.IdPartida)
				// Borrar la partida de la base de datos
				s.PartidasDAO.BorrarPartida(p)
				return
			}
		case mensajesInternos.SalidaUsuario:
			if mt.IdUsuario == p.IdCreador {
				// Enviar a todos que el creador ha abandonado
				p.EnviarATodos(mensajes.ErrorJsonPartida("El creador de "+
					"la sala se ha desconectado", mensajes.CierreSala))
				// Borrar la partida de la estructura del servidor
				s.Partidas.Delete(p.IdPartida)
				// Borrar la partida de la base de datos
				s.PartidasDAO.BorrarPartida(p)
			} else {
				mensajeEnviar := s.PartidasDAO.AbandonarPartida(p, mt.IdUsuario)
				if _, hayError := mensajeEnviar["err"]; !hayError {
					// Enviar al resto de jugadores de la sala los nuevos datos
					p.EnviarATodos(mensajeEnviar)
				}
			}
		case mensajesInternos.MensajeInvalido:
			p.EnviarError(mt.IdUsuario, mensajes.ErrorPeticion, mt.Err)
		}
	}
}

/*
	atenderPartida recibe las diferentes notificaciones de una partida y ejecuta
	las acciones que se requieran para cada notificación.
*/
func (s *Servidor) atenderPartida(p *baseDatos.Partida) {
	// Bucle infinito para leer notificaciones
	for {
		mensajeRecibido := <-p.Mensajes

		switch mt := mensajeRecibido.(type) {
		case mensajesInternos.MensajeFase:
			// Mensaje para avanzar de fase
			pos := p.ObtenerPosicionJugador(mt.IdUsuario)
			msg := p.AvanzarFase(pos)
			if _, hayError := msg["err"]; hayError {
				p.Enviar(mt.IdUsuario, msg)
			} else {
				p.EnviarATodos(msg)
			}
			// Actualizar información de usuarios
			// TODO
			// Guardar información en la base de datos
			// TODO
			// Temporización de los turnos
			// TODO
		case mensajesInternos.MensajeRefuerzos:
			// Mensaje para realizar un refuerzo
			pos := p.ObtenerPosicionJugador(mt.IdUsuario)
			msg := p.Refuerzo(mt.IdTerritorio, pos, mt.Tropas)
			if _, hayError := msg["err"]; hayError {
				p.Enviar(mt.IdUsuario, msg)
			} else {
				p.EnviarATodos(msg)
			}
		case mensajesInternos.MensajeAtaque:
			// Mensaje para realizar un ataque
			pos := p.ObtenerPosicionJugador(mt.IdUsuario)
			msg := p.Ataque(mt.IdTerritorioOrigen, mt.IdTerritorioDestino, pos, mt.Tropas)
			if _, hayError := msg["err"]; hayError {
				p.Enviar(mt.IdUsuario, msg)
			} else {
				p.EnviarATodos(msg)
			}
		case mensajesInternos.MensajeMover:
			// Mensaje para realizar un movimiento
		case mensajesInternos.LlegadaUsuario:
			if pos := p.ObtenerPosicionJugador(mt.IdUsuario); pos == -1 {
				mt.Ws.WriteJSON(mensajes.ErrorJsonPartida("No estás en esta partida", 1))
				mt.Ws.Close()
			} else {
				p.Conexiones.Store(mt.IdUsuario, mt.Ws)
				msg := mensajes.JsonData{}
				mapstructure.Decode(p, &msg)
				msg["_tipoMensaje"] = "p"
				p.Enviar(mt.IdUsuario, msg)
			}
		case mensajesInternos.SalidaUsuario:
			// Desconexión de un usuario
			p.Conexiones.Delete(mt.IdUsuario)
		case mensajesInternos.MensajeInvalido:
			p.EnviarError(mensajes.ErrorPeticion, mt.IdUsuario, mt.Err)
		}
	}
}

/*
	devolverErrorWebsocket envía un error a través de un websocket
*/
func devolverErrorWebsocket(code int, err string, ws *websocket.Conn) {
	resultado := mensajes.ErrorJsonPartida(err, code)
	ws.WriteJSON(resultado)
}
