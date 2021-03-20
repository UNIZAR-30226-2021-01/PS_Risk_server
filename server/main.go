package server

import (
	"PS_Risk_server/baseDatos"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

var bd baseDatos.BD

func Hello() string {
	return "Hello, world."
}

func usuarioInvalido() map[string]interface{} {
	usuario := make(map[string]interface{})
	usuario["id"] = 0
	usuario["nombre"] = ""
	usuario["icono"] = 0
	usuario["aspecto"] = 0
	usuario["correo"] = ""
	usuario["riskos"] = 0
	usuario["recibeCorreos"] = false
	return usuario
}

func registroUsuario(w http.ResponseWriter, r *http.Request) {
	var resultado map[string]interface{}
	nombre := r.FormValue("nombre")
	correo := r.FormValue("correo")
	clave := r.FormValue("clave")
	recibeCorreos, err := strconv.ParseBool(r.FormValue("recibeCorreos"))
	if err != nil {
		resultado = make(map[string]interface{})
		errorDado := make(map[string]interface{})
		errorDado["code"] = baseDatos.ErrorTipoIncorrecto
		errorDado["err"] = err.Error()
		resultado["error"] = errorDado
		resultado["usuario"] = usuarioInvalido()
		resultado["iconos"] = make(map[string]interface{})
		resultado["aspectos"] = make(map[string]interface{})
		resultado["tiendaIconos"] = make(map[string]interface{})
		resultado["tiendaAspectos"] = make(map[string]interface{})
	} else {
		resultado = bd.CrearCuenta(nombre, correo, clave, recibeCorreos)
	}
	respuesta, _ := json.MarshalIndent(resultado, "", " ")
	fmt.Fprintf(w, string(respuesta))
}

func main() {
	fmt.Println("Hola mundo")
	port := os.Getenv("PORT")
	fmt.Println("Iniciando servidor en puerto", port)
	http.HandleFunc("/registrar", registroUsuario)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
