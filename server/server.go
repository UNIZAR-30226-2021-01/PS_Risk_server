package server

import (
	"PS_Risk_server/baseDatos"
	"PS_Risk_server/mensajes"
	"PS_Risk_server/mensajesInternos"

	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"github.com/badoux/checkmail"
	"github.com/go-playground/form/v4"
	"github.com/gorilla/websocket"
	"github.com/jordan-wright/email"
	"github.com/rs/cors"
)

type tokenRecuperacion struct {
	ttl int
	id  int
}

type TTLmap struct {
	m   map[string]tokenRecuperacion
	l   sync.Mutex
	ttl int
}

func CrearTTLmap(ttl int) *TTLmap {
	m := TTLmap{
		m:   make(map[string]tokenRecuperacion),
		l:   sync.Mutex{},
		ttl: ttl,
	}

	go func() {
		for range time.Tick(time.Duration(1) * time.Minute) {
			m.l.Lock()
			for k, v := range m.m {
				v.ttl--
				if v.ttl == 0 {
					delete(m.m, k)
				}
			}
			m.l.Unlock()
		}
	}()

	return &m
}

func (m *TTLmap) NuevoToken(id int) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", errors.New("no se ha podido generar el token")
	}
	token := hex.EncodeToString(b)

	m.l.Lock()
	for _, v := range m.m {
		if v.id == id {
			m.l.Unlock()
			return "", errors.New("ya se ha solicitado restablecer clave")
		}
	}
	m.m[token] = tokenRecuperacion{ttl: m.ttl, id: id}
	m.l.Unlock()

	return token, nil
}

func (m *TTLmap) ConsumirToken(t string) (int, error) {
	m.l.Lock()
	if v, ok := m.m[t]; ok {
		delete(m.m, t)
		m.l.Unlock()
		return v.id, nil
	}
	m.l.Unlock()
	return -1, errors.New("token inválido")
}

/*
	Servidor almacena los DAO, la tienda, las partidas y las salas activas.
*/
type Servidor struct {
	Puerto      string
	SMTPserver  string
	SMTPport    string
	Correo      string
	ClaveCorreo string
	UsuarioDAO  baseDatos.UsuarioDAO
	AmigosDAO   baseDatos.AmigosDAO
	TiendaDAO   baseDatos.TiendaDAO
	PartidasDAO baseDatos.PartidaDAO
	Tienda      baseDatos.Tienda
	upgrader    websocket.Upgrader
	Partidas    sync.Map
	Restablecer *TTLmap
}

/*
	NuevoServidor crea un servidor, establece una conexión con la base de datos
	postgreSQL, crea los DAO que va a utilizar y carga la tienda de la base de
	datos.
*/
func NuevoServidor(p, bbdd, smtpServer, smtpPort, mail, mailPass string) (*Servidor, error) {
	b, err := sql.Open("postgres", bbdd)
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
		SMTPserver:  smtpServer,
		SMTPport:    smtpPort,
		Correo:      mail,
		ClaveCorreo: mailPass,
		UsuarioDAO:  baseDatos.NuevoUsuarioDAO(b),
		AmigosDAO:   baseDatos.NuevoAmigosDAO(b),
		TiendaDAO:   td,
		Tienda:      tienda,
		upgrader:    websocket.Upgrader{},
		PartidasDAO: baseDatos.NuevaPartidaDAO(b),
		Partidas:    sync.Map{},
		Restablecer: CrearTTLmap(10)}, nil
}

/*
	Iniciar hace que el servidor empiece a atender las diferentes peticiones
	de los clientes.
*/
func (s *Servidor) Iniciar() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/registrar", s.registroUsuarioHandler)
	mux.HandleFunc("/iniciarSesion", s.inicioSesionHandler)
	mux.HandleFunc("/recargarUsuario", s.obtenerUsuarioHandler)
	mux.HandleFunc("/personalizarUsuario", s.personalizarUsuarioHandler)
	mux.HandleFunc("/gestionAmistad", s.gestionAmistadHandler)
	mux.HandleFunc("/notificaciones", s.notificacionesHandler)
	mux.HandleFunc("/amigos", s.amigosHandler)
	mux.HandleFunc("/enviarSolicitudAmistad", s.solicitudAmistadHandler)
	mux.HandleFunc("/comprar", s.comprarHandler)
	mux.HandleFunc("/crearSala", s.crearPartidaHandler)
	mux.HandleFunc("/aceptarSala", s.aceptarSalaHandler)
	mux.HandleFunc("/partidas", s.partidasHandler)
	mux.HandleFunc("/rechazarPartida", s.rechazarPartidaHandler)
	mux.HandleFunc("/entrarPartida", s.aceptarSalaHandler)
	mux.HandleFunc("/borrarNotificacionTurno", s.borrarNotificacionTurnoHandler)
	mux.HandleFunc("/borrarCuenta", s.borrarCuentaHandler)
	mux.HandleFunc("/olvidoClave", s.olvidoClaveHandler)
	mux.HandleFunc("/restablecerClave", s.restablecerClaveHandler)

	// Eliminar todas las salas que no se han iniciado
	err := s.PartidasDAO.EliminarSalas()
	if err != nil {
		return err
	}
	// Restaurar todas las partidas de la base de datos
	err = s.RestaurarPartidas()
	if err != nil {
		return err
	}
	handler := cors.Default().Handler(mux)
	s.upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	err = http.ListenAndServe(":"+s.Puerto, handler)
	return err
}

/*
	crearMensajeUsuario crea un mensaje en formato json con los datos de inicio
	de sesión de un usuario.
	Crea un error en caso de no poder obtener los datos.
*/
func (s *Servidor) crearMensajeUsuario(u baseDatos.Usuario) mensajes.JsonData {
	// Obtener iconos de usuario
	iconos, err := s.TiendaDAO.ObtenerIconos(u)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
	}

	// Obtener aspectos de usuario
	aspectos, err := s.TiendaDAO.ObtenerAspectos(u)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
	}

	return mensajes.JsonData{
		"usuario":        u.ToJSON(),
		"iconos":         iconos,
		"aspectos":       aspectos,
		"tiendaIconos":   s.Tienda.Iconos,
		"tiendaAspectos": s.Tienda.Aspectos,
	}
}

/*
	devolverError envía un error en formato json.
*/
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

/*
	registroUsuarioHandler maneja las peticiones para registrar usuarios.
*/
func (s *Servidor) registroUsuarioHandler(w http.ResponseWriter, r *http.Request) {
	var f formularioRegistro

	// Comprobar que los datos recibidos son correctos
	decoder := form.NewDecoder()
	err := r.ParseForm()
	if err != nil {
		devolverError(mensajes.ErrorPeticion, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(mensajes.ErrorPeticion, "Campos formulario incorrectos", w)
		return
	}

	// Comprobar si el formato del correo es correcto
	correo := strings.ToLower(f.Correo)
	if correo != "" {
		if err := checkmail.ValidateFormat(correo); err != nil {
			devolverError(mensajes.ErrorPeticion, "El correo no esta en formato correcto", w)
			return
		}
		if err := checkmail.ValidateHostAndUser(s.SMTPserver, s.Correo, correo); err != nil {
			devolverError(mensajes.ErrorPeticion, "No se ha podido validar el correo", w)
			return
		}
	}

	// Crear la cuenta en la base de datos
	user, err := s.UsuarioDAO.CrearCuenta(f.Nombre, correo, f.Clave, f.RecibeCorreos)
	if err != nil {
		devolverError(mensajes.ErrorPeticion, err.Error(), w)
		return
	}

	// Devolver resultado
	respuesta, _ := json.MarshalIndent(s.crearMensajeUsuario(user), "", " ")
	fmt.Fprint(w, string(respuesta))
}

type formularioInicioSesion struct {
	Usuario string `form:"usuario"`
	Clave   string `form:"clave"`
}

/*
	inicioSesion maneja las peticiones para iniciar sesión.
*/
func (s *Servidor) inicioSesionHandler(w http.ResponseWriter, r *http.Request) {
	var (
		user baseDatos.Usuario
		f    formularioInicioSesion
	)

	// Comprobar que los datos recibidos son correctos
	decoder := form.NewDecoder()
	err := r.ParseForm()
	if err != nil {
		devolverError(mensajes.ErrorPeticion, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(mensajes.ErrorPeticion, "Campos formulario incorrectos", w)
		return
	}

	// Solicitar los datos de inicio de sesión de un usuario a la base de datos
	if strings.Contains(f.Usuario, "@") {
		user, err = s.UsuarioDAO.IniciarSesionCorreo(f.Usuario, f.Clave)
	} else {
		user, err = s.UsuarioDAO.IniciarSesionNombre(f.Usuario, f.Clave)
	}
	if err != nil {
		devolverError(mensajes.ErrorPeticion, "No se ha podido iniciar sesión", w)
		return
	}

	// Devolver resultado
	respuesta, _ := json.MarshalIndent(s.crearMensajeUsuario(user), "", " ")
	fmt.Fprint(w, string(respuesta))
}

type formularioPersonalizarUsuario struct {
	ID    int    `form:"idUsuario"`
	Clave string `form:"clave"`
	Tipo  string `form:"tipo"`
	Dato  string `form:"nuevoDato"`
}

/*
	personalizarUsuarioHandler maneja las peticiones para cambiar los datos del usuario.
*/
func (s *Servidor) personalizarUsuarioHandler(w http.ResponseWriter, r *http.Request) {
	var f formularioPersonalizarUsuario

	// Comprobar que los datos recibidos son correctos
	decoder := form.NewDecoder()
	err := r.ParseForm()
	if err != nil {
		devolverError(mensajes.ErrorPeticion, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(mensajes.ErrorPeticion, "Campos formulario incorrectos", w)
		return
	}

	// Obtener datos actuales del usuario de la base de datos
	user, err := s.UsuarioDAO.ObtenerUsuario(f.ID, f.Clave)
	if err != nil {
		devolverError(mensajes.ErrorUsuario, "No se ha podido obtener el usuario", w)
		return
	}

	if f.Tipo == "Correo" {
		if err := checkmail.ValidateFormat(f.Dato); err != nil {
			devolverError(mensajes.ErrorPeticion, "El correo no esta en formato correcto", w)
			return
		}
		if err := checkmail.ValidateHostAndUser(s.SMTPserver, s.Correo, f.Dato); err != nil {
			devolverError(mensajes.ErrorPeticion, "No se ha podido validar el correo", w)
			return
		}
	}

	// Modificar los datos y devolver el resultado de la modificación
	err = user.Modificar(f.Tipo, f.Dato)
	if err != nil {
		devolverError(mensajes.ErrorPeticion, err.Error(), w)
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

/*
	gestionAmistadHandler maneja las peticiones de borrar a un amigo, aceptar una
	solicitud y rechazarla.
*/
func (s *Servidor) gestionAmistadHandler(w http.ResponseWriter, r *http.Request) {
	var (
		resultado mensajes.JsonData
		f         formularioGestionAmistad
	)

	// Comprobar que los datos recibidos son correctos
	decoder := form.NewDecoder()
	err := r.ParseForm()
	if err != nil {
		devolverError(mensajes.ErrorPeticion, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(mensajes.ErrorPeticion, "Campos formulario incorrectos", w)
		return
	}

	// Obtener los datos del usuario que solicita una acción
	user, err := s.UsuarioDAO.ObtenerUsuario(f.ID, f.Clave)
	if err != nil {
		devolverError(mensajes.ErrorUsuario, "No se ha podido obtener el usuario", w)
		return
	}

	// Realizar la acción en la base de datos
	switch strings.ToLower(f.Decision) {
	case "borrar":
		resultado = s.AmigosDAO.EliminarAmigo(user, f.IDamigo)
	case "aceptar":
		resultado = s.AmigosDAO.AceptarSolicitudAmistad(user, f.IDamigo)
	case "rechazar":
		resultado = s.AmigosDAO.RechazarSolicitudAmistad(user, f.IDamigo)
	default:
		resultado = mensajes.ErrorJson("La decision no es válida",
			mensajes.ErrorPeticion)
	}

	// Devolver resultado
	respuesta, _ := json.MarshalIndent(resultado, "", " ")
	fmt.Fprint(w, string(respuesta))
}

type formularioSolicitudAmistad struct {
	ID     int    `form:"idUsuario"`
	Nombre string `form:"nombreAmigo"`
	Clave  string `form:"clave"`
}

/*
	solicitudAmistadHandler maneja las peticiones de enviar una solicitud de amistad.
*/
func (s *Servidor) solicitudAmistadHandler(w http.ResponseWriter, r *http.Request) {
	var f formularioSolicitudAmistad

	// Comprobar que los datos recibidos son correctos
	decoder := form.NewDecoder()
	err := r.ParseForm()
	if err != nil {
		devolverError(mensajes.ErrorPeticion, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(mensajes.ErrorPeticion, "Campos formulario incorrectos", w)
		return
	}

	// Obtener los datos del usuario que va a enviar la solicitud
	user, err := s.UsuarioDAO.ObtenerUsuario(f.ID, f.Clave)
	if err != nil {
		devolverError(mensajes.ErrorUsuario, "No se ha podido obtener el usuario", w)
		return
	}

	// Enviar la solicitud y devolver el resultado
	respuesta, _ := json.MarshalIndent(s.AmigosDAO.EnviarSolicitudAmistad(user, f.Nombre), "", " ")
	fmt.Fprint(w, string(respuesta))
}

type formularioObtener struct {
	ID    int    `form:"idUsuario"`
	Clave string `form:"clave"`
}

/*
	obtener comprueba que los datos recibidos son correctos y devuelve la información
	solicitada en la función metodo.
*/
func (s *Servidor) obtener(w http.ResponseWriter, r *http.Request,
	metodo func(baseDatos.Usuario) mensajes.JsonData) {

	var f formularioObtener

	// Comprobar que los datos recibidos son correctos
	decoder := form.NewDecoder()
	err := r.ParseForm()
	if err != nil {
		devolverError(mensajes.ErrorPeticion, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(mensajes.ErrorPeticion, "Campos formulario incorrectos", w)
		return
	}

	// Obtener los datos básicos del usuario
	user, err := s.UsuarioDAO.ObtenerUsuario(f.ID, f.Clave)
	if err != nil {
		devolverError(mensajes.ErrorUsuario, "No se ha podido obtener el usuario", w)
		return
	}

	// Devolver los datos solicitados en metodo del usuario
	respuesta, _ := json.MarshalIndent(metodo(user), "", " ")
	fmt.Fprint(w, string(respuesta))
}

/*
	obtenerUsuarioHandler maneja las peticiones para obtener los datos de inicio
	de sesión de un usuario utilizando su id y clave.
*/
func (s *Servidor) obtenerUsuarioHandler(w http.ResponseWriter, r *http.Request) {
	s.obtener(w, r, s.crearMensajeUsuario)
}

/*
	notificacionesHandler maneja las peticiones para obtener las notificaciones
	de un usuario.
*/
func (s *Servidor) notificacionesHandler(w http.ResponseWriter, r *http.Request) {
	s.obtener(w, r, s.UsuarioDAO.ObtenerNotificaciones)
}

/*
	amigosHandler maneja las peticiones para obtener los amigos de un usuario.
*/
func (s *Servidor) amigosHandler(w http.ResponseWriter, r *http.Request) {
	s.obtener(w, r, s.AmigosDAO.ObtenerAmigos)
}

/*
	obtenerPartidasHandler maneja las peticiones para obtener las partidas en
	las que juega un usuario.
*/
func (s *Servidor) partidasHandler(w http.ResponseWriter, r *http.Request) {
	var (
		f         formularioObtener
		jsonArray []mensajes.JsonData
	)

	// Comprobar que los datos recibidos son correctos
	decoder := form.NewDecoder()
	err := r.ParseForm()
	if err != nil {
		devolverError(mensajes.ErrorPeticion, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(mensajes.ErrorPeticion, "Campos formulario incorrectos", w)
		return
	}

	// Obtener los datos básicos del usuario
	user, err := s.UsuarioDAO.ObtenerUsuario(f.ID, f.Clave)
	if err != nil {
		devolverError(mensajes.ErrorUsuario, "No se ha podido obtener el usuario", w)
		return
	}

	// Obtener los identificadores de partidas
	ids, err := s.PartidasDAO.ObtenerPartidas(user)
	if err != nil {
		devolverError(mensajes.ErrorPeticion, "No se han podido obtener las partidas", w)
		return
	}

	// Obtener los datos de las partidas
	for _, id := range ids {
		partida, ok := s.Partidas.Load(id)
		if ok {
			p := partida.(*baseDatos.Partida)
			jsonArray = append(jsonArray, mensajes.JsonData{
				"id":          p.IdPartida,
				"nombre":      p.Nombre,
				"nombreTurno": p.Jugadores[p.TurnoJugador].Nombre,
				"turnoActual": p.TurnoActual,
				"tiempoTurno": p.TiempoTurno,
				"ultimoTurno": p.UltimoTurno,
			})
		}
	}

	// Devolver las partidas
	respuesta, _ := json.MarshalIndent(mensajes.JsonData{"partidas": jsonArray}, "", " ")
	fmt.Fprint(w, string(respuesta))
}

type formularioComprar struct {
	ID        int    `form:"idUsuario"`
	Cosmetico int    `form:"cosmetico"`
	Clave     string `form:"clave"`
	Tipo      string `form:"tipo"`
}

/*
	comprarHandler maneja las peticiones para comprar iconos y aspectos.
*/
func (s *Servidor) comprarHandler(w http.ResponseWriter, r *http.Request) {
	var (
		resultado mensajes.JsonData
		f         formularioComprar
	)

	// Comprobar que los datos recibidos son correctos
	decoder := form.NewDecoder()
	err := r.ParseForm()
	if err != nil {
		devolverError(mensajes.ErrorPeticion, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(mensajes.ErrorPeticion, "Campos formulario incorrectos", w)
		return
	}

	// Obtener los datos del usuario que va a comprar
	user, err := s.UsuarioDAO.ObtenerUsuario(f.ID, f.Clave)
	if err != nil {
		devolverError(mensajes.ErrorUsuario, "No se ha podido obtener el usuario", w)
		return
	}

	// Comprar icono u aspecto
	switch f.Tipo {
	case "Aspecto":
		p, enc := s.Tienda.ObtenerPrecioAspecto(f.Cosmetico)
		if !enc {
			devolverError(mensajes.ErrorPeticion, "Aspecto no encontrado", w)
			return
		}
		resultado = s.TiendaDAO.ComprarAspecto(&user, f.Cosmetico, p)
	case "Icono":
		p, enc := s.Tienda.ObtenerPrecioIcono(f.Cosmetico)
		if !enc {
			devolverError(mensajes.ErrorPeticion, "Icono no encontrado", w)
			return
		}
		resultado = s.TiendaDAO.ComprarIcono(&user, f.Cosmetico, p)
	default:
		devolverError(mensajes.ErrorPeticion, "El tipo no existe", w)
		return
	}

	// Devolver resultado de la compra
	respuesta, _ := json.MarshalIndent(resultado, "", " ")
	fmt.Fprint(w, string(respuesta))
}

type formularioRechazarPartida struct {
	IdUsuario int    `form:"idUsuario"`
	IdSala    int    `form:"idSala"`
	Clave     string `form:"clave"`
}

func (s *Servidor) rechazarPartidaHandler(w http.ResponseWriter, r *http.Request) {
	var f formularioRechazarPartida

	// Comprobar que los datos recibidos son correctos
	decoder := form.NewDecoder()
	err := r.ParseForm()
	if err != nil {
		devolverError(mensajes.ErrorPeticion, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(mensajes.ErrorPeticion, "Campos formulario incorrectos", w)
		return
	}

	// Obtener los datos del usuario que hace la petición
	u, err := s.UsuarioDAO.ObtenerUsuario(f.IdUsuario, f.Clave)
	if err != nil {
		devolverError(mensajes.ErrorUsuario, "No se ha podido obtener el usuario", w)
		return
	}

	// Rechazar la invitación
	err = s.PartidasDAO.RechazarPartida(f.IdSala, u)
	if err != nil {
		devolverError(mensajes.ErrorPeticion, "No se ha podido rechazar la "+
			"invitación", w)
		return
	}

	// Devolver resultado
	devolverError(mensajes.NoError, "", w)
}

func (s *Servidor) borrarNotificacionTurnoHandler(w http.ResponseWriter, r *http.Request) {
	var f formularioRechazarPartida

	// Comprobar que los datos recibidos son correctos
	decoder := form.NewDecoder()
	err := r.ParseForm()
	if err != nil {
		devolverError(mensajes.ErrorPeticion, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(mensajes.ErrorPeticion, "Campos formulario incorrectos", w)
		return
	}

	// Obtener los datos del usuario que hace la petición
	u, err := s.UsuarioDAO.ObtenerUsuario(f.IdUsuario, f.Clave)
	if err != nil {
		devolverError(mensajes.ErrorUsuario, "No se ha podido obtener el usuario", w)
		return
	}

	// Borrar la notificación
	err = s.PartidasDAO.BorrarNotificacionTurno(f.IdSala, u.Id)
	if err != nil {
		devolverError(mensajes.ErrorPeticion, "No se ha podido eliminar la "+
			"notificación", w)
		return
	}

	// Devolver resultado
	devolverError(mensajes.NoError, "", w)
}

func (s *Servidor) borrarCuentaHandler(w http.ResponseWriter, r *http.Request) {
	var f formularioObtener

	// Comprobar que los datos recibidos son correctos
	decoder := form.NewDecoder()
	err := r.ParseForm()
	if err != nil {
		devolverError(mensajes.ErrorPeticion, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(mensajes.ErrorPeticion, "Campos formulario incorrectos", w)
		return
	}

	// Obtener los datos básicos del usuario
	user, err := s.UsuarioDAO.ObtenerUsuario(f.ID, f.Clave)
	if err != nil {
		devolverError(mensajes.ErrorUsuario, "No se ha podido obtener el usuario", w)
		return
	}

	// Obtener partidas empezadas en las que participa
	partidas, err := s.PartidasDAO.ObtenerPartidas(user)
	if err != nil {
		devolverError(mensajes.ErrorUsuario, "No se ha podido eliminar la cuenta: "+err.Error(), w)
		return
	}

	// Eliminar cuenta
	err = s.UsuarioDAO.BorrarUsuario(user)
	if err != nil {
		devolverError(mensajes.ErrorUsuario, "No se ha podido eliminar la cuenta: "+err.Error(), w)
		return
	}

	// Obtener partidas sin empezar en las que participa y avisar de que el
	// usuario ya no existe
	salas, err := s.PartidasDAO.ObtenerSalas()
	if err != nil {
		devolverError(mensajes.ErrorUsuario, "No se ha podido eliminar la cuenta: "+err.Error(), w)
		return
	}
	for _, idSala := range salas {
		pInterface, _ := s.Partidas.Load(idSala)
		p := pInterface.(*baseDatos.Partida)
		if p.EstaEnPartida(f.ID) {
			p.Mensajes <- mensajesInternos.CuentaEliminada{
				IdUsuario: f.ID,
			}
		}
	}

	// Avisar a todas las partidas en las que está de que el usuario ya no existe
	for _, idPartida := range partidas {
		pInterface, _ := s.Partidas.Load(idPartida)
		p := pInterface.(*baseDatos.Partida)
		p.Mensajes <- mensajesInternos.CuentaEliminada{
			IdUsuario: f.ID,
		}
	}

	// Informar de que se ha eliminado correctamente
	devolverError(mensajes.NoError, "", w)
}

type formulariOlvidoClave struct {
	Correo string `form:"correo"`
}

func (s *Servidor) olvidoClaveHandler(w http.ResponseWriter, r *http.Request) {
	var f formulariOlvidoClave

	// Comprobar que los datos recibidos son correctos
	decoder := form.NewDecoder()
	err := r.ParseForm()
	if err != nil {
		devolverError(mensajes.ErrorPeticion, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(mensajes.ErrorPeticion, "Campos formulario incorrectos", w)
		return
	}

	id, err := s.UsuarioDAO.ObtenerId(f.Correo)
	if err != nil {
		devolverError(mensajes.ErrorPeticion, err.Error(), w)
		return
	}

	t, err := s.Restablecer.NuevoToken(id)
	if err != nil {
		devolverError(mensajes.ErrorPeticion, err.Error(), w)
		return
	}

	e := email.NewEmail()
	e.From = "PixelRisk <" + s.Correo + ">"
	e.To = []string{f.Correo}
	e.Subject = "Recuperación de clave"
	e.Text = []byte("https://risk-webapp.herokuapp.com/restablecerClave/" + t)
	err = e.Send(s.SMTPserver+":"+s.SMTPport, smtp.PlainAuth("", s.Correo, s.ClaveCorreo, s.SMTPserver))
	if err != nil {
		devolverError(mensajes.ErrorPeticion, "No se ha podido enviar el correo para restablecer la clave", w)
		return
	}

	devolverError(mensajes.NoError, "", w)
}

type formularioRestablecerClave struct {
	Token string `form:"token"`
	Clave string `form:"clave"`
}

func (s *Servidor) restablecerClaveHandler(w http.ResponseWriter, r *http.Request) {
	var f formularioRestablecerClave

	// Comprobar que los datos recibidos son correctos
	decoder := form.NewDecoder()
	err := r.ParseForm()
	if err != nil {
		devolverError(mensajes.ErrorPeticion, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(mensajes.ErrorPeticion, "Campos formulario incorrectos", w)
		return
	}

	id, err := s.Restablecer.ConsumirToken(f.Token)
	if err != nil {
		devolverError(mensajes.ErrorPeticion, "Token inválido", w)
		return
	}

	u, err := s.UsuarioDAO.ObtenerUsuarioId(id)
	if err != nil {
		devolverError(mensajes.ErrorPeticion, err.Error(), w)
		return
	}

	u.Modificar("Clave", f.Clave)
	respuesta, _ := json.MarshalIndent(s.UsuarioDAO.ActualizarUsuario(u), "", " ")
	fmt.Fprint(w, string(respuesta))
}
