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
		fmt.Println("Usuario recibido:", usuario)
		fmt.Println("Clave recibida:", clave)
		if strings.Contains(usuario, "@") {
			resultado = s.bd.IniciarSesionCorreo(usuario, clave)
		} else {
			resultado = s.bd.IniciarSesionNombre(usuario, clave)
		}
	}
	respuesta, _ := json.MarshalIndent(resultado, "", " ")
	fmt.Fprintf(w, string(respuesta))
}
