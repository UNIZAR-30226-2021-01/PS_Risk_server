package server

import (
	"PS_Risk_server/baseDatos"
	"PS_Risk_server/mensajes"
	"PS_Risk_server/mensajesInternos"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
)

func devolverErrorWebsocket(code int, err string, ws *websocket.Conn) {
	resultado := mensajes.ErrorJsonPartida(err, code)
	ws.WriteJSON(resultado)
}

type mensajeCreacion struct {
	IdUsuario   int    `json:"idUsuario"`
	Clave       string `json:"clave"`
	TiempoTurno int    `json:"tiempoTurno"`
	NombreSala  string `json:"nombreSala"`
}

func (s *Servidor) crearPartidaHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Upgrader err: %v\n", err)
		return
	}
	defer ws.Close()
	var mensajeInicial mensajeCreacion
	err = ws.ReadJSON(&mensajeInicial)
	if err != nil {
		devolverErrorWebsocket(1, err.Error(), ws)
		return
	}
	u, err := s.UsuarioDAO.ObtenerUsuario(mensajeInicial.IdUsuario, mensajeInicial.Clave)
	if err != nil {
		devolverErrorWebsocket(1, err.Error(), ws)
		return
	}
	p, err := s.PartidasDAO.CrearPartida(u, mensajeInicial.TiempoTurno,
		mensajeInicial.NombreSala, ws)
	if err != nil {
		devolverErrorWebsocket(1, err.Error(), ws)
		return
	}
	s.Partidas.Store(p.IdPartida, p)

	aux := mensajes.JsonData{}
	mapstructure.Decode(p, &aux)
	aux["_tipoMensaje"] = "d"
	enviarPorWebsocket(p, aux, u.Id)

	go s.atenderSala(p)

	s.recibirMensajesUsuarioEnSala(u, ws, p)
}

type mensajeEntrarSala struct {
	IdUsuario int    `json:"idUsuario"`
	Clave     string `json:"clave"`
	IdSala    int    `json:"idSala"`
}

func (s *Servidor) aceptarSalaHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Upgrader err: %v\n", err)
		return
	}
	defer ws.Close()
	var mensajeInicial mensajeEntrarSala
	err = ws.ReadJSON(&mensajeInicial)
	if err != nil {
		devolverErrorWebsocket(1, err.Error(), ws)
		return
	}
	u, err := s.UsuarioDAO.ObtenerUsuario(mensajeInicial.IdUsuario,
		mensajeInicial.Clave)
	if err != nil {
		devolverErrorWebsocket(1, err.Error(), ws)
		return
	}
	idPartida := mensajeInicial.IdSala
	p, ok := s.Partidas.Load(idPartida)
	if !ok {
		devolverErrorWebsocket(1, "Partida no encontrada", ws)
		return
	}
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

func (s *Servidor) recibirMensajesUsuarioEnSala(u baseDatos.Usuario,
	ws *websocket.Conn, p *baseDatos.Partida) {
	var mensajeRecibido mensajeSala
	for {
		err := ws.ReadJSON(&mensajeRecibido)
		if err != nil {
			if e, ok := err.(*websocket.CloseError); (ok &&
				(e.Code == websocket.CloseNormalClosure ||
					e.Code == websocket.CloseAbnormalClosure ||
					e.Code == websocket.CloseNoStatusReceived)) || err == io.ErrUnexpectedEOF {
				p.Mensajes <- mensajesInternos.SalidaUsuario{
					IdUsuario: u.Id,
				}
			} else {
				p.Mensajes <- mensajesInternos.MensajeInvalido{
					IdUsuario: u.Id,
					Err:       err.Error(),
				}
			}
			return
		}
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

func enviarPorWebsocket(p *baseDatos.Partida, mensaje mensajes.JsonData, idUsuario int) {
	wsInterface, _ := p.Conexiones.Load(idUsuario)
	if wsInterface != nil {
		ws, ok := wsInterface.(*websocket.Conn)
		if ok {
			ws.WriteJSON(mensaje)
		}
	}
}

func enviarATodos(p *baseDatos.Partida, mensaje mensajes.JsonData) {
	for _, jugador := range p.Jugadores {
		enviarPorWebsocket(p, mensaje, jugador.Id)
	}
}

func devolverErrorUsuario(p *baseDatos.Partida, code, idUsuario int, err string) {
	wsInterface, _ := p.Conexiones.Load(idUsuario)
	if wsInterface != nil {
		ws, ok := wsInterface.(*websocket.Conn)
		if ok {
			devolverErrorWebsocket(code, err, ws)
		}
	}
}

func (s *Servidor) atenderSala(p *baseDatos.Partida) {
	for {
		mensajeRecibido := <-p.Mensajes
		switch mt := mensajeRecibido.(type) {
		case mensajesInternos.LlegadaUsuario:
			u, err := s.UsuarioDAO.ObtenerUsuarioId(mt.IdUsuario)
			if err != nil {
				devolverErrorWebsocket(1, err.Error(), mt.Ws)
			} else {
				mensajeEnviar := s.PartidasDAO.EntrarPartida(p, u, mt.Ws)
				_, hayError := mensajeEnviar["err"]
				if hayError {
					mt.Ws.WriteJSON(mensajeEnviar)
				} else {
					enviarATodos(p, mensajeEnviar)
				}
			}
		case mensajesInternos.InvitacionPartida:
			err := s.PartidasDAO.InvitarPartida(p, mt.IdCreador, mt.IdInvitado)
			if err != nil {
				devolverErrorUsuario(p, 1, mt.IdCreador, err.Error())
			} else {
				devolverErrorUsuario(p, baseDatos.NoError, mt.IdCreador, "")
			}
		case mensajesInternos.InicioPartida:
			mensajeEnviar := s.PartidasDAO.IniciarPartida(p, mt.IdUsuario)
			_, hayError := mensajeEnviar["err"]
			if hayError {
				enviarPorWebsocket(p, mensajeEnviar, mt.IdUsuario)
			} else {
				enviarATodos(p, mensajeEnviar)
				return
			}
		case mensajesInternos.SalidaUsuario:
			if mt.IdUsuario == p.IdCreador {
				enviarATodos(p, mensajes.ErrorJsonPartida("El creador de "+
					"la sala se ha desconectado", 1))
				s.Partidas.Delete(p.IdPartida)
				s.PartidasDAO.BorrarPartida(p)
			} else {
				err := s.PartidasDAO.AbandonarPartida(p, mt.IdUsuario)
				if err != nil {
					fmt.Println("Error al quitar un usuario de una partida:",
						err.Error())
				}
			}
		case mensajesInternos.MensajeInvalido:
			devolverErrorUsuario(p, 1, mt.IdUsuario, mt.Err)
		case mensajesInternos.FinPartida:
			enviarATodos(p, mensajes.JsonData{
				"_tipoMensaje": "f",
				"ganador":      "",
				"riskos":       0,
			})
			s.Partidas.Delete(p.IdPartida)
			s.PartidasDAO.BorrarPartida(p)
			return
		}
	}
}
