package server

import (
	"PS_Risk_server/baseDatos"
	"PS_Risk_server/mensajes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	http.HandleFunc("/recargarUsuario", s.recargarUsuarioHandler)
	http.HandleFunc("/personalizarUsuario", s.personalizarUsuarioHandler)
	err := http.ListenAndServe(":"+os.Args[1], nil)
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
		if err != nil {
			resultado = mensajes.ErrorJson(err.Error(), baseDatos.ErrorTipoIncorrecto)
		} else {
			resultado = s.bd.CrearCuenta(nombre, correo, clave, recibeCorreos)
		}
	}
	respuesta, _ := json.MarshalIndent(resultado, "", " ")
	fmt.Fprintf(w, string(respuesta))
}

func (s *Servidor) inicioSesion(w http.ResponseWriter, r *http.Request) {
	var resultado mensajes.JsonData
	err := r.ParseForm()
	if err != nil {
		resultado = mensajes.ErrorJson(err.Error(), baseDatos.ErrorTipoIncorrecto)
	} else {
		usuario := r.FormValue("usuario")
		clave := r.FormValue("clave")
		if strings.Contains(usuario, "@") {
			resultado = s.bd.IniciarSesionCorreo(usuario, clave)
		} else {
			resultado = s.bd.IniciarSesionNombre(usuario, clave)
		}
	}
	respuesta, _ := json.MarshalIndent(resultado, "", " ")
	fmt.Fprintf(w, string(respuesta))
}

func (s *Servidor) recargarUsuarioHandler(w http.ResponseWriter, r *http.Request) {
	var resultado mensajes.JsonData
	err := r.ParseForm()
	if err != nil {
		resultado = mensajes.ErrorJson(err.Error(), baseDatos.ErrorTipoIncorrecto)
	} else {
		id, err := strconv.Atoi(r.FormValue("idUsuario"))
		if err != nil {
			resultado = mensajes.ErrorJson(err.Error(), baseDatos.ErrorTipoIncorrecto)
		} else {
			clave := r.FormValue("clave")
			resultado = s.bd.ObtenerUsuario(id, clave)
		}
	}
	respuesta, _ := json.MarshalIndent(resultado, "", " ")
	fmt.Fprintf(w, string(respuesta))
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
			if !(strings.EqualFold(campo, "Aspecto") || strings.EqualFold(campo, "Icono") ||
				strings.EqualFold(campo, "Correo") || strings.EqualFold(campo, "Clave") ||
				strings.EqualFold(campo, "Nombre")) {
				resultado = mensajes.ErrorJson("El campo indicado a modificar no existe", baseDatos.ErrorCampoIncorrecto)
			} else {
				resultado = s.bd.ModificarUsuario(id, clave, campo, nuevoDato)
			}
		}
	}
	respuesta, _ := json.MarshalIndent(resultado, "", " ")
	fmt.Fprintf(w, string(respuesta))
}
