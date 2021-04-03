package server

import (
	"PS_Risk_server/baseDatos"
	"PS_Risk_server/mensajes"
	"PS_Risk_server/mensajesInternos"
	"PS_Risk_server/partidas"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-playground/form/v4"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

type Servidor struct {
	Puerto      string
	UsuarioDAO  baseDatos.UsuarioDAO
	AmigosDAO   baseDatos.AmigosDAO
	TiendaDAO   baseDatos.TiendaDAO
	Tienda      baseDatos.Tienda
	upgrader    websocket.Upgrader
	PartidasDAO baseDatos.PartidaDAO
	Partidas    map[int]partidas.Partida
}

func NuevoServidor(p, bbdd string) (*Servidor, error) {
	b, err := baseDatos.CrearBD(bbdd)
	if err != nil {
		return nil, err
	}
	td := baseDatos.NuevaTiendaDAO(b)
	tienda, err := td.ObtenerTienda()
	if err != nil {
		return nil, err
	}
	return &Servidor{
		Puerto:      p,
		UsuarioDAO:  baseDatos.NuevoUsuarioDAO(b),
		AmigosDAO:   baseDatos.NuevoAmigosDAO(b),
		TiendaDAO:   td,
		Tienda:      tienda,
		upgrader:    websocket.Upgrader{},
		PartidasDAO: baseDatos.NuevaPartidaDAO(b),
		Partidas:    make(map[int]partidas.Partida)}, nil
}

func (s *Servidor) Iniciar() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/registrar", s.registroUsuario)
	mux.HandleFunc("/iniciarSesion", s.inicioSesion)
	mux.HandleFunc("/recargarUsuario", s.obtenerUsuarioHandler)
	mux.HandleFunc("/personalizarUsuario", s.personalizarUsuarioHandler)
	mux.HandleFunc("/gestionAmistad", s.gestionAmistadHandler)
	mux.HandleFunc("/notificaciones", s.notificacionesHandler)
	mux.HandleFunc("/amigos", s.amigosHandler)
	mux.HandleFunc("/enviarSolicitudAmistad", s.solicitudAmistadHandler)
	mux.HandleFunc("/comprar", s.comprarHandler)
	mux.HandleFunc("/crearSala", s.crearPartidaHandler)
	handler := cors.Default().Handler(mux)
	err := http.ListenAndServe(":"+s.Puerto, handler)
	return err
}

func (s *Servidor) crearMensajeUsuario(u baseDatos.Usuario) mensajes.JsonData {
	iconos, err := s.TiendaDAO.ObtenerIconos(u)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), 1)
	}
	aspectos, err := s.TiendaDAO.ObtenerAspectos(u)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), 1)
	}
	return mensajes.JsonData{
		"usuario":        u.ToJSON(),
		"iconos":         iconos,
		"aspectos":       aspectos,
		"tiendaIconos":   s.Tienda.Iconos,
		"tiendaAspectos": s.Tienda.Aspectos,
	}
}

func devolverError(code int, err string, w http.ResponseWriter) {
	resultado := mensajes.ErrorJson(err, code)
	respuesta, _ := json.MarshalIndent(resultado, "", " ")
	fmt.Fprint(w, string(respuesta))
}

type formularioRegistro struct {
	Nombre        string `form:"nombre"`
	Correo        string `form:"correo"`
	Clave         string `form:"clave"`
	RecibeCorreos bool   `form:"recibeCorreos"`
}

func (s *Servidor) registroUsuario(w http.ResponseWriter, r *http.Request) {
	var f formularioRegistro
	decoder := form.NewDecoder()
	err := r.ParseForm()
	if err != nil {
		devolverError(1, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(1, "Campos formulario incorrectos", w)
		return
	}
	user, err := s.UsuarioDAO.CrearCuenta(f.Nombre, f.Correo, f.Clave, f.RecibeCorreos)
	if err != nil {
		devolverError(1, err.Error(), w)
		return
	}
	respuesta, _ := json.MarshalIndent(s.crearMensajeUsuario(user), "", " ")
	fmt.Fprint(w, string(respuesta))
}

type formularioInicioSesion struct {
	Usuario string `form:"usuario"`
	Clave   string `form:"clave"`
}

func (s *Servidor) inicioSesion(w http.ResponseWriter, r *http.Request) {
	var (
		user baseDatos.Usuario
		f    formularioInicioSesion
	)
	decoder := form.NewDecoder()
	err := r.ParseForm()
	if err != nil {
		devolverError(1, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(1, "Campos formulario incorrectos", w)
		return
	}
	if strings.Contains(f.Usuario, "@") {
		user, err = s.UsuarioDAO.IniciarSesionCorreo(f.Usuario, f.Clave)
	} else {
		user, err = s.UsuarioDAO.IniciarSesionNombre(f.Usuario, f.Clave)
	}
	if err != nil {
		devolverError(1, "No se ha podido iniciar sesion", w)
		return
	}
	respuesta, _ := json.MarshalIndent(s.crearMensajeUsuario(user), "", " ")
	fmt.Fprint(w, string(respuesta))
}

type formularioPersonalizarUsuario struct {
	ID    int    `form:"idUsuario"`
	Clave string `form:"clave"`
	Tipo  string `form:"tipo"`
	Dato  string `form:"nuevoDato"`
}

func (s *Servidor) personalizarUsuarioHandler(w http.ResponseWriter, r *http.Request) {
	var f formularioPersonalizarUsuario
	decoder := form.NewDecoder()
	err := r.ParseForm()
	if err != nil {
		devolverError(1, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(1, "Campos formulario incorrectos", w)
		return
	}
	user, err := s.UsuarioDAO.ObtenerUsuario(f.ID, f.Clave)
	if err != nil {
		devolverError(1, "No se ha podido obtener el usuario", w)
		return
	}
	err = user.Modificar(f.Tipo, f.Dato)
	if err != nil {
		devolverError(1, err.Error(), w)
		return
	}
	respuesta, _ := json.MarshalIndent(s.UsuarioDAO.ActualizarUsuario(user), "", " ")
	fmt.Fprint(w, string(respuesta))
}

type formularioGestionAmistad struct {
	ID       int    `form:"idUsuario"`
	IDamigo  int    `form:"idAmigo"`
	Clave    string `form:"clave"`
	Decision string `form:"decision"`
}

func (s *Servidor) gestionAmistadHandler(w http.ResponseWriter, r *http.Request) {
	var (
		resultado mensajes.JsonData
		f         formularioGestionAmistad
	)
	decoder := form.NewDecoder()
	err := r.ParseForm()
	if err != nil {
		devolverError(1, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(1, "Campos formulario incorrectos", w)
		return
	}
	user, err := s.UsuarioDAO.ObtenerUsuario(f.ID, f.Clave)
	if err != nil {
		devolverError(1, "No se ha podido obtener el usuario", w)
		return
	}
	switch strings.ToLower(f.Decision) {
	case "borrar":
		resultado = s.AmigosDAO.EliminarAmigo(user, f.IDamigo)
	case "aceptar":
		resultado = s.AmigosDAO.AceptarSolicitudAmistad(user, f.IDamigo)
	case "rechazar":
		resultado = s.AmigosDAO.RechazarSolicitudAmistad(user, f.IDamigo)
	default:
		resultado = mensajes.ErrorJson("La decision no es valida",
			baseDatos.ErrorTipoIncorrecto)
	}
	respuesta, _ := json.MarshalIndent(resultado, "", " ")
	fmt.Fprint(w, string(respuesta))
}

type formularioSolicitudAmistad struct {
	ID     int    `form:"idUsuario"`
	Nombre string `form:"nombreAmigo"`
	Clave  string `form:"clave"`
}

func (s *Servidor) solicitudAmistadHandler(w http.ResponseWriter, r *http.Request) {
	var f formularioSolicitudAmistad
	decoder := form.NewDecoder()
	err := r.ParseForm()
	if err != nil {
		devolverError(1, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(1, "Campos formulario incorrectos", w)
		return
	}
	user, err := s.UsuarioDAO.ObtenerUsuario(f.ID, f.Clave)
	if err != nil {
		devolverError(1, "No se ha podido obtener el usuario", w)
		return
	}
	respuesta, _ := json.MarshalIndent(s.AmigosDAO.EnviarSolicitudAmistad(user, f.Nombre), "", " ")
	fmt.Fprint(w, string(respuesta))
}

type formularioObtener struct {
	ID    int    `form:"idUsuario"`
	Clave string `form:"clave"`
}

func (s *Servidor) obtener(w http.ResponseWriter, r *http.Request,
	metodo func(baseDatos.Usuario) mensajes.JsonData) {
	var f formularioObtener
	decoder := form.NewDecoder()
	err := r.ParseForm()
	if err != nil {
		devolverError(1, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(1, "Campos formulario incorrectos", w)
		return
	}
	user, err := s.UsuarioDAO.ObtenerUsuario(f.ID, f.Clave)
	if err != nil {
		devolverError(1, "No se ha podido obtener el usuario", w)
		return
	}
	respuesta, _ := json.MarshalIndent(metodo(user), "", " ")
	fmt.Fprint(w, string(respuesta))
}

func (s *Servidor) obtenerUsuarioHandler(w http.ResponseWriter, r *http.Request) {
	s.obtener(w, r, s.crearMensajeUsuario)
}

func (s *Servidor) notificacionesHandler(w http.ResponseWriter, r *http.Request) {
	s.obtener(w, r, s.UsuarioDAO.ObtenerNotificaciones)
}

func (s *Servidor) amigosHandler(w http.ResponseWriter, r *http.Request) {
	s.obtener(w, r, s.AmigosDAO.ObtenerAmigos)
}

type formularioComprar struct {
	ID        int    `form:"idUsuario"`
	Cosmetico int    `form:"cosmetico"`
	Clave     string `form:"clave"`
	Tipo      string `form:"tipo"`
}

func (s *Servidor) comprarHandler(w http.ResponseWriter, r *http.Request) {
	var (
		resultado mensajes.JsonData
		f         formularioComprar
	)
	decoder := form.NewDecoder()
	err := r.ParseForm()
	if err != nil {
		devolverError(1, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(1, "Campos formulario incorrectos", w)
		return
	}
	user, err := s.UsuarioDAO.ObtenerUsuario(f.ID, f.Clave)
	if err != nil {
		devolverError(1, "No se ha podido obtener el usuario", w)
		return
	}
	switch f.Tipo {
	case "Aspecto":
		p, enc := s.Tienda.ObtenerPrecioAspecto(f.Cosmetico)
		if !enc {
			devolverError(1, "Aspecto no encontrado", w)
			return
		}
		resultado = s.TiendaDAO.ComprarAspecto(&user, f.Cosmetico, p)
	case "Icono":
		p, enc := s.Tienda.ObtenerPrecioIcono(f.Cosmetico)
		if !enc {
			devolverError(1, "Icono no encontrado", w)
			return
		}
		resultado = s.TiendaDAO.ComprarIcono(&user, f.Cosmetico, p)
	default:
		devolverError(1, "El tipo no existe", w)
		return
	}
	respuesta, _ := json.MarshalIndent(resultado, "", " ")
	fmt.Fprint(w, string(respuesta))
}

func devolverErrorWebsocket(code int, err string, ws *websocket.Conn) {
	resultado := mensajes.ErrorJson(err, code)
	ws.WriteJSON(resultado)
}

// si se ha cerrado la conexión, ReadJSON devuelve io.ErrUnexpectedEOF

func (s *Servidor) crearPartidaHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Upgrader err: %v\n", err)
		return
	}
	defer ws.Close()
	mensajeRecibido := mensajes.JsonData{}
	err = ws.ReadJSON(&mensajeRecibido)
	if err != nil {
		devolverErrorWebsocket(1, err.Error(), ws)
		return
	}
	idUsuario, ok := mensajeRecibido["idUsuario"].(int)
	if !ok {
		devolverErrorWebsocket(1, "El id del usuario debe ser un entero", ws)
		return
	}
	u, err := s.UsuarioDAO.ObtenerUsuario(idUsuario, mensajeRecibido["clave"].(string))
	if err != nil {
		devolverErrorWebsocket(1, err.Error(), ws)
		return
	}
	tiempoTurno, ok := mensajeRecibido["tiempoTurno"].(int)
	if !ok {
		devolverErrorWebsocket(1, "El tiempo de turno debe ser un entero (minutos)", ws)
		return
	}
	p, err := s.PartidasDAO.CrearPartida(u, tiempoTurno,
		mensajeRecibido["nombreSala"].(string), ws)
	if err != nil {
		devolverErrorWebsocket(1, err.Error(), ws)
		return
	}
	s.Partidas[p.IdPartida] = p

	go s.atenderSala(&p)

	for {
		err = ws.ReadJSON(&mensajeRecibido)
		if err != nil {
			if err == io.ErrUnexpectedEOF {
				p.Mensajes <- mensajesInternos.FinPartida{}
			} else {
				devolverErrorWebsocket(1, err.Error(), ws)
			}
			return
		}
		tipoAccion, ok := mensajeRecibido["tipo"].(string)
		if !ok {
			devolverErrorWebsocket(1, "El tipo de acción debe ser un string", ws)
		} else if strings.ToLower(tipoAccion) == "invitar" {
			p.Mensajes <- mensajesInternos.InvitacionPartida{
				IdCreador:  u.Id,
				IdInvitado: mensajeRecibido["idInvitado"].(int),
			}
		} else if strings.ToLower(tipoAccion) == "iniciar" {
			p.Mensajes <- mensajesInternos.InicioPartida{
				IdUsuario: u.Id,
			}
		} else {
			devolverErrorWebsocket(1, "El tipo de acción debe ser Invitar o Iniciar", ws)
		}
	}
}

func enviarATodos(p partidas.Partida, mensaje mensajes.JsonData) {
	for _, ws := range p.Conexiones {
		if ws != nil {
			ws.WriteJSON(mensaje)
		}
	}
}

func (s *Servidor) atenderSala(p *partidas.Partida) {
	for {
		mensajeRecibido := <-p.Mensajes
		switch mt := mensajeRecibido.(type) {
		case mensajesInternos.LlegadaUsuario:
			u, err := s.UsuarioDAO.ObtenerUsuarioId(mt.IdUsuario)
			if err != nil {
				devolverErrorWebsocket(1, err.Error(), mt.Ws)
			} else {
				mensajeEnviar := s.PartidasDAO.UnirsePartida(p, u, mt.Ws)
				_, hayError := mensajeEnviar["err"]
				if hayError {
					mt.Ws.WriteJSON(mensajeEnviar)
				} else {
					enviarATodos(*p, mensajeEnviar)
				}
			}
		case mensajesInternos.InvitacionPartida:
			u, err := s.UsuarioDAO.ObtenerUsuarioId(mt.IdCreador)
			if err != nil {
				devolverErrorWebsocket(1, err.Error(), p.Conexiones[mt.IdCreador])
			} else {
				err := s.PartidasDAO.InvitarPartida(p, u, mt.IdInvitado)
				if err != nil {
					devolverErrorWebsocket(1, err.Error(), p.Conexiones[mt.IdCreador])
				}
			}
		case mensajesInternos.InicioPartida:
			u, err := s.UsuarioDAO.ObtenerUsuarioId(mt.IdUsuario)
			if err != nil {
				devolverErrorWebsocket(1, err.Error(), p.Conexiones[mt.IdUsuario])
			} else {
				mensajeEnviar := s.PartidasDAO.IniciarPartida(p, u)
				_, hayError := mensajeEnviar["err"]
				if hayError {
					p.Conexiones[mt.IdUsuario].WriteJSON(mensajeEnviar)
				} else {
					enviarATodos(*p, mensajeEnviar)
				}
			}
			return
		case mensajesInternos.SalidaUsuario:
			u, err := s.UsuarioDAO.ObtenerUsuarioId(mt.IdUsuario)
			fmt.Println("Error al leer un usuario:", err.Error())
			err = s.PartidasDAO.QuitarJugadorPartida(p, u)
			fmt.Println("Error al quitar un usuario de una partida:", err.Error())
		case mensajesInternos.FinPartida:
			s.PartidasDAO.BorrarPartida(p)
			return
		}
	}
}
