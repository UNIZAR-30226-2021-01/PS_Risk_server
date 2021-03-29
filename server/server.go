package server

import (
	"PS_Risk_server/baseDatos"
	"PS_Risk_server/mensajes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Servidor struct {
	Puerto     string
	UsuarioDAO baseDatos.UsuarioDAO
	AmigosDAO  baseDatos.AmigosDAO
	TiendaDAO  baseDatos.TiendaDAO
	Tienda     baseDatos.Tienda
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
		Puerto:     p,
		UsuarioDAO: baseDatos.NuevoUsuarioDAO(b),
		AmigosDAO:  baseDatos.NuevoAmigosDAO(b),
		TiendaDAO:  td,
		Tienda:     tienda}, nil
}

func (s *Servidor) Iniciar() error {
	http.HandleFunc("/registrar", s.registroUsuario)
	http.HandleFunc("/iniciarSesion", s.inicioSesion)
	http.HandleFunc("/recargarUsuario", s.obtenerUsuarioHandler)
	http.HandleFunc("/personalizarUsuario", s.personalizarUsuarioHandler)
	http.HandleFunc("/gestionAmistad", s.gestionAmistadHandler)
	http.HandleFunc("/notificaciones", s.notificacionesHandler)
	http.HandleFunc("/amigos", s.amigosHandler)
	http.HandleFunc("/enviarSolicitudAmistad", s.solicitudAmistadHandler)
	err := http.ListenAndServe(":"+s.Puerto, nil)
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

func (s *Servidor) registroUsuario(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	err := r.ParseForm()
	if err != nil {
		devolverError(1, err.Error(), w)
		return
	}
	nombre := r.FormValue("nombre")
	correo := r.FormValue("correo")
	clave := r.FormValue("clave")
	recibeCorreos, err := strconv.ParseBool(r.FormValue("recibeCorreos"))
	if err != nil || nombre == "" || correo == "" || clave == "" {
		devolverError(1, "Campos formulario incorrectos", w)
		return
	}
	user, err := s.UsuarioDAO.CrearCuenta(nombre, correo, clave, recibeCorreos)
	if err != nil {
		devolverError(1, err.Error(), w)
		return
	}
	respuesta, _ := json.MarshalIndent(s.crearMensajeUsuario(user), "", " ")
	fmt.Fprint(w, string(respuesta))
}

func (s *Servidor) inicioSesion(w http.ResponseWriter, r *http.Request) {
	var user baseDatos.Usuario
	w.Header().Set("Access-Control-Allow-Origin", "*")
	err := r.ParseForm()
	if err != nil {
		devolverError(1, err.Error(), w)
		return
	}
	usuario := r.FormValue("usuario")
	clave := r.FormValue("clave")
	if err != nil || usuario == "" || clave == "" {
		devolverError(1, "Campos formulario incorrectos", w)
		return
	}
	if strings.Contains(usuario, "@") {
		user, err = s.UsuarioDAO.IniciarSesionCorreo(usuario, clave)
	} else {
		user, err = s.UsuarioDAO.IniciarSesionNombre(usuario, clave)
	}
	if err != nil {
		devolverError(1, "No se ha podido iniciar sesion", w)
		return
	}
	respuesta, _ := json.MarshalIndent(s.crearMensajeUsuario(user), "", " ")
	fmt.Fprint(w, string(respuesta))
}

func (s *Servidor) personalizarUsuarioHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	err := r.ParseForm()
	if err != nil {
		devolverError(1, err.Error(), w)
		return
	}
	nuevoDato := r.FormValue("nuevoDato")
	clave := r.FormValue("clave")
	campo := r.FormValue("tipo")
	id, err := strconv.Atoi(r.FormValue("idUsuario"))
	if nuevoDato == "" || clave == "" || campo == "" || err != nil {
		devolverError(1, "Campos formulario incorrectos", w)
		return
	}
	if !(strings.EqualFold(campo, "Aspecto") || strings.EqualFold(campo, "Icono") ||
		strings.EqualFold(campo, "Correo") || strings.EqualFold(campo, "Clave") ||
		strings.EqualFold(campo, "Nombre") || strings.EqualFold(campo, "RecibeCorreos")) {
		devolverError(1, "El campo indicado a modificar no existe", w)
		return
	}
	user, err := s.UsuarioDAO.ObtenerUsuario(id, clave)
	if err != nil {
		devolverError(1, "No se ha podido obtener el usuario", w)
		return
	}
	err = user.Modificar(campo, nuevoDato)
	if err != nil {
		devolverError(1, err.Error(), w)
		return
	}
	respuesta, _ := json.MarshalIndent(s.UsuarioDAO.ActualizarUsuario(user), "", " ")
	fmt.Fprint(w, string(respuesta))
}

func (s *Servidor) gestionAmistadHandler(w http.ResponseWriter, r *http.Request) {
	var resultado mensajes.JsonData
	w.Header().Set("Access-Control-Allow-Origin", "*")
	err := r.ParseForm()
	if err != nil {
		devolverError(1, err.Error(), w)
		return
	}
	clave := r.FormValue("clave")
	decision := r.FormValue("decision")
	idUsuario, err := strconv.Atoi(r.FormValue("idUsuario"))
	if err != nil || clave == "" {
		devolverError(1, "Campos formulario incorrectos", w)
		return
	}
	idAmigo, err := strconv.Atoi(r.FormValue("idAmigo"))
	if err != nil {
		devolverError(1, "Campos formulario incorrectos", w)
		return
	}
	user, err := s.UsuarioDAO.ObtenerUsuario(idUsuario, clave)
	if err != nil {
		devolverError(1, "No se ha podido obtener el usuario", w)
		return
	}
	switch strings.ToLower(decision) {
	case "borrar":
		resultado = s.AmigosDAO.EliminarAmigo(user, idAmigo)
	case "aceptar":
		resultado = s.AmigosDAO.AceptarSolicitudAmistad(user, idAmigo)
	case "rechazar":
		resultado = s.AmigosDAO.RechazarSolicitudAmistad(user, idAmigo)
	default:
		resultado = mensajes.ErrorJson("Campos formulario incorrectos",
			baseDatos.ErrorTipoIncorrecto)
	}
	respuesta, _ := json.MarshalIndent(resultado, "", " ")
	fmt.Fprint(w, string(respuesta))
}

func (s *Servidor) solicitudAmistadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	err := r.ParseForm()
	if err != nil {
		devolverError(1, err.Error(), w)
		return
	}
	nombre := r.FormValue("nombreAmigo")
	clave := r.FormValue("clave")
	idUsuario, err := strconv.Atoi(r.FormValue("idUsuario"))
	if err != nil || nombre == "" || clave == "" {
		devolverError(1, "Campos fromulario incorrectos", w)
		return
	}
	user, err := s.UsuarioDAO.ObtenerUsuario(idUsuario, clave)
	if err != nil {
		devolverError(1, "No se ha podido obtener el usuario", w)
		return
	}
	respuesta, _ := json.MarshalIndent(s.AmigosDAO.EnviarSolicitudAmistad(user, nombre), "", " ")
	fmt.Fprint(w, string(respuesta))
}

func (s *Servidor) obtenerUsuarioHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	err := r.ParseForm()
	if err != nil {
		devolverError(1, err.Error(), w)
		return
	}
	clave := r.FormValue("clave")
	idUsuario, err := strconv.Atoi(r.FormValue("idUsuario"))
	if err != nil || clave == "" {
		devolverError(1, "Campos fromulario incorrectos", w)
		return
	}
	user, err := s.UsuarioDAO.ObtenerUsuario(idUsuario, clave)
	if err != nil {
		devolverError(1, "No se ha podido obtener el usuario", w)
		return
	}
	respuesta, _ := json.MarshalIndent(s.crearMensajeUsuario(user), "", " ")
	fmt.Fprint(w, string(respuesta))
}

func (s *Servidor) notificacionesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	err := r.ParseForm()
	if err != nil {
		devolverError(1, err.Error(), w)
		return
	}
	clave := r.FormValue("clave")
	idUsuario, err := strconv.Atoi(r.FormValue("idUsuario"))
	if err != nil || clave == "" {
		devolverError(1, "Campos fromulario incorrectos", w)
		return
	}
	user, err := s.UsuarioDAO.ObtenerUsuario(idUsuario, clave)
	if err != nil {
		devolverError(1, "No se ha podido obtener el usuario", w)
		return
	}
	respuesta, _ := json.MarshalIndent(s.UsuarioDAO.ObtenerNotificaciones(user), "", " ")
	fmt.Fprint(w, string(respuesta))
}

func (s *Servidor) amigosHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	err := r.ParseForm()
	if err != nil {
		devolverError(1, err.Error(), w)
		return
	}
	clave := r.FormValue("clave")
	idUsuario, err := strconv.Atoi(r.FormValue("idUsuario"))
	if err != nil || clave == "" {
		devolverError(1, "Campos fromulario incorrectos", w)
		return
	}
	user, err := s.UsuarioDAO.ObtenerUsuario(idUsuario, clave)
	if err != nil {
		devolverError(1, "No se ha podido obtener el usuario", w)
		return
	}
	respuesta, _ := json.MarshalIndent(s.AmigosDAO.ObtenerAmigos(user), "", " ")
	fmt.Fprint(w, string(respuesta))
}
