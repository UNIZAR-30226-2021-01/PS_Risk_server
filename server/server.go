package server

import (
	"PS_Risk_server/baseDatos"
	"PS_Risk_server/mensajes"
	"sync"

	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/form/v4"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

/*
	Servidor almacena los DAO, la tienda, las partidas y las salas activas.
*/
type Servidor struct {
	Puerto      string
	UsuarioDAO  baseDatos.UsuarioDAO
	AmigosDAO   baseDatos.AmigosDAO
	TiendaDAO   baseDatos.TiendaDAO
	PartidasDAO baseDatos.PartidaDAO
	Tienda      baseDatos.Tienda
	upgrader    websocket.Upgrader
	Partidas    sync.Map
}

/*
	NuevoServidor crea un servidor, establece una conexión con la base de datos
	postgreSQL, crea los DAO que va a utilizar y carga la tienda de la base de
	datos.
*/
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
		Partidas:    sync.Map{}}, nil
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
	handler := cors.Default().Handler(mux)
	s.upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	err := http.ListenAndServe(":"+s.Puerto, handler)
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
		return mensajes.ErrorJson(err.Error(), 1)
	}

	// Obtener aspectos de usuario
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
	registroUsuario maneja las peticiones para registrar usuarios.
*/
func (s *Servidor) registroUsuarioHandler(w http.ResponseWriter, r *http.Request) {
	var f formularioRegistro

	// Comprobar que los datos recibidos son correctos
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

	// Crear la cuenta en la base de datos
	user, err := s.UsuarioDAO.CrearCuenta(f.Nombre, strings.ToLower(f.Correo),
		f.Clave, f.RecibeCorreos)
	if err != nil {
		devolverError(1, err.Error(), w)
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
		devolverError(1, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(1, "Campos formulario incorrectos", w)
		return
	}

	// Solicitar los datos de inicio de sesión de un usuario a la base de datos
	if strings.Contains(f.Usuario, "@") {
		user, err = s.UsuarioDAO.IniciarSesionCorreo(f.Usuario, f.Clave)
	} else {
		user, err = s.UsuarioDAO.IniciarSesionNombre(f.Usuario, f.Clave)
	}
	if err != nil {
		devolverError(1, "No se ha podido iniciar sesion", w)
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
		devolverError(1, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(1, "Campos formulario incorrectos", w)
		return
	}

	// Obtener datos actuales del usuario de la base de datos
	user, err := s.UsuarioDAO.ObtenerUsuario(f.ID, f.Clave)
	if err != nil {
		devolverError(1, "No se ha podido obtener el usuario", w)
		return
	}

	// Modificar los datos y devolver el resultado de la modificación
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
		devolverError(1, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(1, "Campos formulario incorrectos", w)
		return
	}

	// Obtener los datos del usuario que solicita una acción
	user, err := s.UsuarioDAO.ObtenerUsuario(f.ID, f.Clave)
	if err != nil {
		devolverError(1, "No se ha podido obtener el usuario", w)
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
		resultado = mensajes.ErrorJson("La decision no es valida",
			baseDatos.ErrorTipoIncorrecto)
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
		devolverError(1, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(1, "Campos formulario incorrectos", w)
		return
	}

	// Obtener los datos del usuario que va a enviar la solicitud
	user, err := s.UsuarioDAO.ObtenerUsuario(f.ID, f.Clave)
	if err != nil {
		devolverError(1, "No se ha podido obtener el usuario", w)
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
	solicitada en la función método.
*/
func (s *Servidor) obtener(w http.ResponseWriter, r *http.Request,
	metodo func(baseDatos.Usuario) mensajes.JsonData) {

	var f formularioObtener

	// Comprobar que los datos recibidos son correctos
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

	// Obtener los datos básicos del usuario
	user, err := s.UsuarioDAO.ObtenerUsuario(f.ID, f.Clave)
	if err != nil {
		devolverError(1, "No se ha podido obtener el usuario", w)
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
	obtenerPartidasHandler maneja las peticiones para obtener las partidas de un usuario.
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
		devolverError(1, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(1, "Campos formulario incorrectos", w)
		return
	}

	// Obtener los datos básicos del usuario
	user, err := s.UsuarioDAO.ObtenerUsuario(f.ID, f.Clave)
	if err != nil {
		devolverError(1, "No se ha podido obtener el usuario", w)
		return
	}

	// Obtener los identificadores de partidas
	ids, err := s.PartidasDAO.ObtenerPartidas(user)
	if err != nil {
		devolverError(1, "No se han podido obtener las partidas", w)
		return
	}

	// Obtener los datos de las partidas
	for _, id := range ids {
		partida, ok := s.Partidas.Load(id)
		if ok {
			p := partida.(*baseDatos.Partida)
			turno := p.TurnoActual
			jsonArray = append(jsonArray, mensajes.JsonData{
				"id":          p.IdPartida,
				"nombre":      p.Nombre,
				"nombreTurno": p.Jugadores[turno].Nombre,
				"turnoActual": turno,
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
		devolverError(1, err.Error(), w)
		return
	}
	err = decoder.Decode(&f, r.PostForm)
	if err != nil {
		devolverError(1, "Campos formulario incorrectos", w)
		return
	}

	// Obtener los datos del usuario que va a comprar
	user, err := s.UsuarioDAO.ObtenerUsuario(f.ID, f.Clave)
	if err != nil {
		devolverError(1, "No se ha podido obtener el usuario", w)
		return
	}

	// Comprar icono u aspecto
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

	// Devolver resultado de la compra
	respuesta, _ := json.MarshalIndent(resultado, "", " ")
	fmt.Fprint(w, string(respuesta))
}
