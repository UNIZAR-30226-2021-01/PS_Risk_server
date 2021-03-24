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
	puerto string
	bd     *baseDatos.BD
}

func NuevoServidor(p, bbdd string) (*Servidor, error) {
	b, err := baseDatos.NuevaBD(bbdd)
	if err != nil {
		return nil, err
	}
	return &Servidor{puerto: p, bd: b}, nil
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
	err := http.ListenAndServe(":"+s.puerto, nil)
	return err
}

func (s *Servidor) registroUsuario(w http.ResponseWriter, r *http.Request) {
	var resultado mensajes.JsonData
	err := r.ParseForm()
	if err != nil {
		resultado = mensajes.ErrorJson(err.Error(), baseDatos.ErrorTipoIncorrecto)
	} else {
		nombre := r.FormValue("nombre")
		correo := r.FormValue("correo")
		clave := r.FormValue("clave")
		recibeCorreos, err := strconv.ParseBool(r.FormValue("recibeCorreos"))
		if err != nil || nombre == "" || correo == "" || clave == "" {
			resultado = mensajes.ErrorJson("Campos formulario incorrectos",
				baseDatos.ErrorTipoIncorrecto)
		} else {
			resultado = s.bd.CrearCuenta(nombre, correo, clave, recibeCorreos)
		}
	}
	respuesta, _ := json.MarshalIndent(resultado, "", " ")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, string(respuesta))
}

func (s *Servidor) inicioSesion(w http.ResponseWriter, r *http.Request) {
	var resultado mensajes.JsonData
	err := r.ParseForm()
	if err != nil {
		resultado = mensajes.ErrorJson(err.Error(), baseDatos.ErrorTipoIncorrecto)
	} else {
		usuario := r.FormValue("usuario")
		clave := r.FormValue("clave")
		if err != nil || usuario == "" || clave == "" {
			resultado = mensajes.ErrorJson("Campos formulario incorrectos",
				baseDatos.ErrorTipoIncorrecto)
		} else {
			if strings.Contains(usuario, "@") {
				resultado = s.bd.IniciarSesionCorreo(usuario, clave)
			} else {
				resultado = s.bd.IniciarSesionNombre(usuario, clave)
			}
		}
	}
	respuesta, _ := json.MarshalIndent(resultado, "", " ")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, string(respuesta))
}

func (s *Servidor) personalizarUsuarioHandler(w http.ResponseWriter, r *http.Request) {
	var resultado mensajes.JsonData
	err := r.ParseForm()
	if err != nil {
		resultado = mensajes.ErrorJson(err.Error(), baseDatos.ErrorTipoIncorrecto)
	} else {
		id, err := strconv.Atoi(r.FormValue("idUsuario"))
		if err != nil {
			resultado = mensajes.ErrorJson(err.Error(), baseDatos.ErrorTipoIncorrecto)
		} else {
			nuevoDato := r.FormValue("nuevoDato")
			clave := r.FormValue("clave")
			campo := r.FormValue("tipo")
			if nuevoDato == "" || clave == "" || campo == "" {
				resultado = mensajes.ErrorJson("Campos formulario incorrectos",
					baseDatos.ErrorTipoIncorrecto)
			} else {
				if !(strings.EqualFold(campo, "Aspecto") || strings.EqualFold(campo, "Icono") ||
					strings.EqualFold(campo, "Correo") || strings.EqualFold(campo, "Clave") ||
					strings.EqualFold(campo, "Nombre") || strings.EqualFold(campo, "RecibeCorreos")) {
					resultado = mensajes.ErrorJson("El campo indicado a modificar no existe",
						baseDatos.ErrorCampoIncorrecto)
				} else {
					resultado = s.bd.ModificarUsuario(id, clave, campo, nuevoDato)
				}
			}
		}
	}
	respuesta, _ := json.MarshalIndent(resultado, "", " ")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, string(respuesta))
}

func (s *Servidor) gestionAmistadHandler(w http.ResponseWriter, r *http.Request) {
	var resultado mensajes.JsonData
	err := r.ParseForm()
	if err != nil {
		resultado = mensajes.ErrorJson(err.Error(), baseDatos.ErrorTipoIncorrecto)
	} else {
		clave := r.FormValue("clave")
		decision := r.FormValue("decision")
		idUsuario, err := strconv.Atoi(r.FormValue("idUsuario"))
		if err != nil || clave == "" {
			resultado = mensajes.ErrorJson("Campos formulario incorrectos",
				baseDatos.ErrorTipoIncorrecto)
		} else {
			idAmigo, err := strconv.Atoi(r.FormValue("idAmigo"))
			if err != nil {
				resultado = mensajes.ErrorJson("Campos formulario incorrectos",
					baseDatos.ErrorTipoIncorrecto)
			} else {
				switch strings.ToLower(decision) {
				case "borrar":
					resultado = s.bd.EliminarAmigo(idUsuario, idAmigo, clave)
				case "aceptar":
					resultado = s.bd.AceptarSolicitudAmistad(idUsuario, idAmigo, clave)
				case "rechazar":
					resultado = s.bd.RechazarSolicitudAmistad(idUsuario, idAmigo, clave)
				default:
					resultado = mensajes.ErrorJson("Campos formulario incorrectos",
						baseDatos.ErrorTipoIncorrecto)
				}
			}
		}
	}
	respuesta, _ := json.MarshalIndent(resultado, "", " ")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, string(respuesta))
}

func (s *Servidor) solicitudAmistadHandler(w http.ResponseWriter, r *http.Request) {
	var resultado mensajes.JsonData
	err := r.ParseForm()
	if err != nil {
		resultado = mensajes.ErrorJson(err.Error(), baseDatos.ErrorTipoIncorrecto)
	} else {
		nombre := r.FormValue("nombreAmigo")
		clave := r.FormValue("clave")
		idUsuario, err := strconv.Atoi(r.FormValue("idUsuario"))
		if err != nil || nombre == "" || clave == "" {
			resultado = mensajes.ErrorJson("Campos fromulario incorrectos",
				baseDatos.ErrorTipoIncorrecto)
		} else {
			resultado = s.bd.EnviarSolicitudAmistad(idUsuario, nombre, clave)
		}
	}
	respuesta, _ := json.MarshalIndent(resultado, "", " ")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, string(respuesta))
}

func (s *Servidor) obtenerUsuarioHandler(w http.ResponseWriter, r *http.Request) {
	obtenerHandler(w, r, s.bd.ObtenerUsuario)
}

func (s *Servidor) notificacionesHandler(w http.ResponseWriter, r *http.Request) {
	obtenerHandler(w, r, s.bd.ObtenerNotificaciones)
}

func (s *Servidor) amigosHandler(w http.ResponseWriter, r *http.Request) {
	obtenerHandler(w, r, s.bd.ObtenerAmigos)
}

func obtenerHandler(w http.ResponseWriter, r *http.Request,
	metodo func(int, string) mensajes.JsonData) {

	var resultado mensajes.JsonData
	err := r.ParseForm()
	if err != nil {
		resultado = mensajes.ErrorJson(err.Error(), baseDatos.ErrorTipoIncorrecto)
	} else {
		clave := r.FormValue("clave")
		idUsuario, err := strconv.Atoi(r.FormValue("idUsuario"))
		if err != nil || clave == "" {
			resultado = mensajes.ErrorJson("Campos fromulario incorrectos",
				baseDatos.ErrorTipoIncorrecto)
		} else {
			resultado = metodo(idUsuario, clave)
		}
	}
	respuesta, _ := json.MarshalIndent(resultado, "", " ")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, string(respuesta))
}
