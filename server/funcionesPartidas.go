package server

import (
	"PS_Risk_server/baseDatos"
	"PS_Risk_server/mensajes"
	"PS_Risk_server/mensajesInternos"
	"PS_Risk_server/partidas"
	"PS_Risk_server/usuarios"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

func devolverErrorWebsocket(code int, err string, ws *websocket.Conn) {
	resultado := mensajes.ErrorJsonPartida(err, code)
	log.Println("devolverErrorWebsocket:", resultado)
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
	} else {
		devolverErrorWebsocket(baseDatos.NoError, "", ws)
	}
	s.Partidas[p.IdPartida] = p

	go s.atenderSala(p)

	s.recibirMensajesUsuarioEnSala(p.IdPartida, u, ws)
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
	s.Partidas[idPartida].Mensajes <- mensajesInternos.LlegadaUsuario{
		U: u,
		//IdUsuario: u.Id,
		Ws: ws,
	}
	s.recibirMensajesUsuarioEnSala(idPartida, u, ws)
}

type mensajeSala struct {
	Tipo       string `json:"tipo"`
	IdInvitado int    `json:"idInvitado,omitempty"`
}

func (s *Servidor) recibirMensajesUsuarioEnSala(idPartida int, u usuarios.Usuario,
	ws *websocket.Conn) {
	var mensajeRecibido mensajeSala
	for {
		err := ws.ReadJSON(&mensajeRecibido)
		if err != nil {
			if err == io.ErrUnexpectedEOF {
				s.Partidas[idPartida].Mensajes <- mensajesInternos.SalidaUsuario{
					U: u,
					//IdUsuario: u.Id,
				}
			} else {
				s.Partidas[idPartida].Mensajes <- mensajesInternos.MensajeInvalido{
					IdUsuario: u.Id,
					Err:       err.Error(),
				}
			}
			return
		}
		if strings.EqualFold(mensajeRecibido.Tipo, "Iniciar") {
			s.Partidas[idPartida].Mensajes <- mensajesInternos.InicioPartida{
				U: u,
				//IdUsuario: u.Id,
			}
		} else if strings.EqualFold(mensajeRecibido.Tipo, "Invitar") {

			s.Partidas[idPartida].Mensajes <- mensajesInternos.InvitacionPartida{
				U: u,
				//IdCreador:  u.Id,
				IdInvitado: mensajeRecibido.IdInvitado,
			}
		} else {
			s.Partidas[idPartida].Mensajes <- mensajesInternos.MensajeInvalido{
				IdUsuario: u.Id,
				Err:       "El tipo de acción debe ser Invitar o Iniciar",
			}
		}
	}
}

func enviarPorWebsocket(p *partidas.Partida, mensaje mensajes.JsonData, idUsuario int) {
	wsInterface, _ := p.Conexiones.Load(idUsuario)
	if wsInterface != nil {
		ws, ok := wsInterface.(*websocket.Conn)
		if ok {
			ws.WriteJSON(mensaje)
		}
	}
}

func enviarATodos(p *partidas.Partida, mensaje mensajes.JsonData) {
	for _, id := range p.Jugadores {
		enviarPorWebsocket(p, mensaje, id)
	}
}

func devolverErrorUsuario(p *partidas.Partida, code, idUsuario int, err string) {
	log.Println("devolverErrorUsuario:", err)
	wsInterface, _ := p.Conexiones.Load(idUsuario)
	if wsInterface != nil {
		ws, ok := wsInterface.(*websocket.Conn)
		if ok {
			devolverErrorWebsocket(code, err, ws)
		}
	}
}

func (s *Servidor) atenderSala(p *partidas.Partida) {
	for {
		mensajeRecibido := <-p.Mensajes
		switch mt := mensajeRecibido.(type) {
		case mensajesInternos.LlegadaUsuario:
			u := mt.U
			/*u, err := s.UsuarioDAO.ObtenerUsuarioId(mt.IdUsuario)
			if err != nil {
				devolverErrorWebsocket(1, err.Error(), mt.Ws)
			} else {*/
			mensajeEnviar := s.PartidasDAO.EntrarPartida(p, u, mt.Ws)
			_, hayError := mensajeEnviar["err"]
			if hayError {
				mt.Ws.WriteJSON(mensajeEnviar)
			} else {
				enviarATodos(p, mensajeEnviar)
			}
			//}
		case mensajesInternos.InvitacionPartida:
			u := mt.U
			err := s.PartidasDAO.InvitarPartida(p, u, mt.IdInvitado)
			if err != nil {
				log.Print(err)
				devolverErrorUsuario(p, 1, u.Id, err.Error())
			} else {
				devolverErrorUsuario(p, baseDatos.NoError, u.Id, "")
			}
			/*u, err := s.UsuarioDAO.ObtenerUsuarioId(mt.IdCreador)
			if err != nil {
				devolverErrorUsuario(p, 1, mt.IdCreador, err.Error())
			} else {
				err = s.PartidasDAO.InvitarPartida(p, u, mt.IdInvitado)
				if err != nil {
					devolverErrorUsuario(p, 1, mt.IdCreador, err.Error())
				} else {
					devolverErrorUsuario(p, baseDatos.NoError, mt.IdCreador, "")
				}
			}*/
		case mensajesInternos.InicioPartida:
			u := mt.U
			/*u, err := s.UsuarioDAO.ObtenerUsuarioId(mt.IdUsuario)
			if err != nil {
				devolverErrorUsuario(p, 1, mt.IdUsuario, err.Error())
			} else {*/
			mensajeEnviar := s.PartidasDAO.IniciarPartida(p, u)
			_, hayError := mensajeEnviar["err"]
			if hayError {
				enviarPorWebsocket(p, mensajeEnviar, u.Id)
				//enviarPorWebsocket(p, mensajeEnviar, mt.IdUsuario)
			} else {
				enviarATodos(p, mensajeEnviar)
				return
			}
			//}
		case mensajesInternos.SalidaUsuario:
			u := mt.U
			/*u, err := s.UsuarioDAO.ObtenerUsuarioId(mt.IdUsuario)
			if err != nil {
				fmt.Println("Error al leer un usuario:", err.Error())
			} else {*/
			if u.Id == p.IdCreador {
				enviarATodos(p, mensajes.ErrorJsonPartida("El creador de "+
					"la sala se ha desconectado", 1))
				s.PartidasDAO.BorrarPartida(p)
			} else {
				err := s.PartidasDAO.AbandonarPartida(p, u)
				if err != nil {
					fmt.Println("Error al quitar un usuario de una partida:",
						err.Error())
				}
			}
			//}
		case mensajesInternos.MensajeInvalido:
			devolverErrorUsuario(p, 1, mt.IdUsuario, mt.Err)
		case mensajesInternos.FinPartida:
			enviarATodos(p, mensajes.JsonData{
				"_tipoMensaje": "f",
				"ganador":      "",
				"riskos":       0,
			})
			s.PartidasDAO.BorrarPartida(p)
			return
		}
	}
}
