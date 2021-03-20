package server

import (
	"PS_Risk_server/baseDatos"
	"PS_Risk_server/mensajes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

func Hello() string {
	return "Hello, world."
}

type servidor struct {
	puerto string
	bd     *baseDatos.BD
}

func NuevoServidor(p, bbdd string) (*servidor, error) {
	b, err := baseDatos.NuevaBD(bbdd)
	if err != nil {
		return nil, err
	}
	return &servidor{puerto: p, bd: b}, nil
}

func (s *servidor) registroUsuario(w http.ResponseWriter, r *http.Request) {
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

func main() {
	fmt.Println("Hola mundo")
	serv, err := NuevoServidor(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Println(err)
		return
	}
	http.HandleFunc("/registrar", serv.registroUsuario)
	err = http.ListenAndServe(":"+os.Args[1], nil)
	if err != nil {
		fmt.Println(err)
	}
}
