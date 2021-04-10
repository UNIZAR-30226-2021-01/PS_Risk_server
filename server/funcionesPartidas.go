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
		devolverErrorWebsocket(1, err.Error(), ws)
		return
	}

	// Comprobar si el usuario está registrado
	u, err := s.UsuarioDAO.ObtenerUsuario(mensajeInicial.IdUsuario, mensajeInicial.Clave)
	if err != nil {
		devolverErrorWebsocket(1, err.Error(), ws)
		return
	}

	// Crea la partida en la base de datos
	p, err := s.PartidasDAO.CrearPartida(u, mensajeInicial.TiempoTurno,
		mensajeInicial.NombreSala, ws)
	if err != nil {
		devolverErrorWebsocket(1, err.Error(), ws)
		return
	}

	// Guarda la referencia a la partida en el servidor
	s.Partidas.Store(p.IdPartida, p)

	// Envía al usuario los datos de la partida en formato json
	aux := mensajes.JsonData{}
	mapstructure.Decode(p, &aux)
	aux["_tipoMensaje"] = "d"
	enviarPorWebsocket(p, aux, u.Id)

	// Lanza una rutina para atender a la sala
	go s.atenderSala(p)

	s.recibirMensajesUsuarioEnSala(u, ws, p)
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
		devolverErrorWebsocket(1, err.Error(), ws)
		return
	}

	// Comprobar si el usuario está registrado
	u, err := s.UsuarioDAO.ObtenerUsuario(mensajeInicial.IdUsuario,
		mensajeInicial.Clave)
	if err != nil {
		devolverErrorWebsocket(1, err.Error(), ws)
		return
	}

	// Carga la referencia a la partida guardada en el servidor
	p, ok := s.Partidas.Load(mensajeInicial.IdSala)
	if !ok {
		devolverErrorWebsocket(1, "Partida no encontrada", ws)
		return
	}

	// Notifica a la partida de que un jugador nuevo ha entrado
	p.(*baseDatos.Partida).Mensajes <- mensajesInternos.LlegadaUsuario{
		IdUsuario: u.Id,
		Ws:        ws,
	}
	s.recibirMensajesUsuarioEnSala(u, ws, p.(*baseDatos.Partida))
}

type mensajeSala struct {
	Tipo       string `json:"tipo"`
	IdInvitado int    `json:"idInvitado,omitempty"`
}

/*
	recibirMensajesUsuarioEnSala recibe los mensajes que un usuario que está en
	una sala envía a través de su websocket. Comprueba si los datos enviados son
	correctos y notifica de eventos a la partida.
*/
func (s *Servidor) recibirMensajesUsuarioEnSala(u baseDatos.Usuario,
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

		// Comprobar si es una invitación a otro usuario o iniciar la partida y notificarla
		if strings.EqualFold(mensajeRecibido.Tipo, "Iniciar") {
			p.Mensajes <- mensajesInternos.InicioPartida{
				IdUsuario: u.Id,
			}
		} else if strings.EqualFold(mensajeRecibido.Tipo, "Invitar") {
			p.Mensajes <- mensajesInternos.InvitacionPartida{
				IdCreador:  u.Id,
				IdInvitado: mensajeRecibido.IdInvitado,
			}
		} else {
			p.Mensajes <- mensajesInternos.MensajeInvalido{
				IdUsuario: u.Id,
				Err:       "El tipo de acción debe ser Invitar o Iniciar",
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
				devolverErrorWebsocket(1, err.Error(), mt.Ws)
			} else {
				// Añadir el usuario a la partida
				mensajeEnviar := s.PartidasDAO.EntrarPartida(p, u, mt.Ws)
				_, hayError := mensajeEnviar["err"]
				if hayError {
					mt.Ws.WriteJSON(mensajeEnviar)
				} else {
					// Si no hay error enviar a todos los jugadores de la sala
					// el nuevo usuario
					enviarATodos(p, mensajeEnviar)
				}
			}
		case mensajesInternos.InvitacionPartida:
			// Invitar usuario a la partida
			err := s.PartidasDAO.InvitarPartida(p, mt.IdCreador, mt.IdInvitado)
			if err != nil {
				devolverErrorUsuario(p, 1, mt.IdCreador, err.Error())
			} else {
				devolverErrorUsuario(p, baseDatos.NoError, mt.IdCreador, "")
			}
		case mensajesInternos.InicioPartida:
			// Iniciar la partida
			mensajeEnviar := s.PartidasDAO.IniciarPartida(p, mt.IdUsuario)
			_, hayError := mensajeEnviar["err"]
			if hayError {
				enviarPorWebsocket(p, mensajeEnviar, mt.IdUsuario)
			} else {
				// Enviar a todos los jugadores de la sala los datos de inicio
				enviarATodos(p, mensajeEnviar)
				return // Cambiar función de gestión de sala por gestión de partida
			}
		case mensajesInternos.SalidaUsuario:
			if mt.IdUsuario == p.IdCreador {
				// Enviar a todos que el creador ha abandonado
				enviarATodos(p, mensajes.ErrorJsonPartida("El creador de "+
					"la sala se ha desconectado", 1))
				// Borrar la partida de la estructura del servidor
				s.Partidas.Delete(p.IdPartida)
				// Borrar la partida de la base de datos
				s.PartidasDAO.BorrarPartida(p)
			} else {
				mensajeEnviar := s.PartidasDAO.AbandonarPartida(p, mt.IdUsuario)
				_, hayError := mensajeEnviar["err"]
				if !hayError {
					// Enviar al resto de jugadores de la sala los nuevos datos
					enviarATodos(p, mensajeEnviar)
				}
			}
		case mensajesInternos.MensajeInvalido:
			devolverErrorUsuario(p, 1, mt.IdUsuario, mt.Err)
		}
	}
}

// REQUIEREN  REVISION

/*
	enviarPorWebsocket envía a un usuario dentro de una sala un mensaje a través
	de su websocket
*/
func enviarPorWebsocket(p *baseDatos.Partida, mensaje mensajes.JsonData, idUsuario int) {
	wsInterface, _ := p.Conexiones.Load(idUsuario)
	if wsInterface != nil {
		ws, ok := wsInterface.(*websocket.Conn)
		if ok {
			ws.WriteJSON(mensaje)
		}
	}
}

/*
	enviarATodos envía a todos los usuarios de una sala un mensaje a través de sus
	websockets
*/
func enviarATodos(p *baseDatos.Partida, mensaje mensajes.JsonData) {
	for _, jugador := range p.Jugadores {
		enviarPorWebsocket(p, mensaje, jugador.Id)
	}
}

/*
	devolverErrorUsuario envía un error a un usuario de una sala a través de su websocket
*/
func devolverErrorUsuario(p *baseDatos.Partida, code, idUsuario int, err string) {
	wsInterface, _ := p.Conexiones.Load(idUsuario)
	if wsInterface != nil {
		ws, ok := wsInterface.(*websocket.Conn)
		if ok {
			devolverErrorWebsocket(code, err, ws)
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
