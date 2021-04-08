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

type Servidor struct {
	Puerto      string
	UsuarioDAO  baseDatos.UsuarioDAO
	AmigosDAO   baseDatos.AmigosDAO
	TiendaDAO   baseDatos.TiendaDAO
	Tienda      baseDatos.Tienda
	upgrader    websocket.Upgrader
	PartidasDAO baseDatos.PartidaDAO
	Partidas    sync.Map
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
		Partidas:    sync.Map{}}, nil
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
	mux.HandleFunc("/aceptarSala", s.aceptarSalaHandler)
	handler := cors.Default().Handler(mux)
	s.upgrader.CheckOrigin = func(r *http.Request) bool { return true }
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
	user, err := s.UsuarioDAO.CrearCuenta(f.Nombre, strings.ToLower(f.Correo), f.Clave, f.RecibeCorreos)
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
