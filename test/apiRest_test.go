package test

import (
	"PS_Risk_server/mensajes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"testing"
)

const apiUrl = "http://localhost:8080"

func realizarPeticionAPI(funcion string, datos url.Values, t *testing.T) mensajes.JsonData {

	res, err := http.PostForm(apiUrl+"/"+funcion, datos)

	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	datosRecibidos := mensajes.JsonData{}
	err = json.Unmarshal(body, &datosRecibidos)
	if err != nil {
		t.Fatal(err)
	}
	return datosRecibidos
}

func Test_CrearEliminarCuenta(t *testing.T) {
	res := realizarPeticionAPI("registrar",
		url.Values{
			"nombre":        {"NombreTest"},
			"correo":        {""},
			"clave":         {"claveTest"},
			"recibeCorreos": {"false"},
		}, t)

	id := int(res["usuario"].(map[string]interface{})["id"].(float64))
	usuarioTest := map[string]interface{}{
		"id":            id,
		"nombre":        "NombreTest",
		"icono":         0,
		"aspecto":       0,
		"correo":        "",
		"riskos":        1000,
		"recibeCorreos": false,
	}

	if !comprobarJson(res["usuario"].(map[string]interface{}), usuarioTest) {
		t.Log(usuarioTest)
		t.Log(res["usuario"].(map[string]interface{}))
		t.Fatal("No coinciden los usuarios")
	}

	res = realizarPeticionAPI("borrarCuenta",
		url.Values{
			"idUsuario": {strconv.Itoa(id)},
			"clave":     {"claveTest"},
		}, t)
	if res["code"].(float64) != 0 {
		t.Fatal(res)
	}
}

func crearCuenta(nombre, correo, clave string, recibeCorreos bool, t *testing.T) int {
	res := realizarPeticionAPI("registrar",
		url.Values{
			"nombre":        {nombre},
			"correo":        {correo},
			"clave":         {clave},
			"recibeCorreos": {strconv.FormatBool(recibeCorreos)},
		}, t)
	if _, ok := res["code"].(float64); ok {
		t.Fatal(res)
	}
	return int(res["usuario"].(map[string]interface{})["id"].(float64))
}

func borrarCuenta(id int, clave string, t *testing.T) {
	res := realizarPeticionAPI("borrarCuenta",
		url.Values{
			"idUsuario": {strconv.Itoa(id)},
			"clave":     {clave},
		}, t)

	if res["code"].(float64) != 0 {
		t.Fatal(res)
	}
}

func Test_RecargarUsuario(t *testing.T) {
	id := crearCuenta("NombreTest", "", "claveTest", false, t)

	res := realizarPeticionAPI("recargarUsuario",
		url.Values{
			"idUsuario": {strconv.Itoa(id)},
			"clave":     {"claveTest"},
		}, t)

	usuarioTest := map[string]interface{}{
		"id":            id,
		"nombre":        "NombreTest",
		"icono":         0,
		"aspecto":       0,
		"correo":        "",
		"riskos":        1000,
		"recibeCorreos": false,
	}
	if !comprobarJson(res["usuario"].(map[string]interface{}), usuarioTest) {
		t.Log(usuarioTest)
		t.Log(res["usuario"].(map[string]interface{}))
		t.Fatal("No coinciden los usuarios")
	}

	borrarCuenta(id, "claveTest", t)
}

func Test_IniciarSesionNombre(t *testing.T) {
	id := crearCuenta("NombreTest", "", "claveTest", false, t)

	res := realizarPeticionAPI("iniciarSesion",
		url.Values{
			"usuario": {"NombreTest"},
			"clave":   {"claveTest"},
		}, t)

	usuarioTest := map[string]interface{}{
		"id":            id,
		"nombre":        "NombreTest",
		"icono":         0,
		"aspecto":       0,
		"correo":        "",
		"riskos":        1000,
		"recibeCorreos": false,
	}
	if !comprobarJson(res["usuario"].(map[string]interface{}), usuarioTest) {
		t.Log(usuarioTest)
		t.Log(res["usuario"].(map[string]interface{}))
		t.Fatal("No coinciden los usuarios")
	}

	borrarCuenta(id, "claveTest", t)
}

func Test_EnviarSolicitudDeAmistad(t *testing.T) {

	idAmigoTest1 := crearCuenta("NombreTestAmigo1", "", "claveTest", false, t)
	idAmigoTest2 := crearCuenta("NombreTestAmigo2", "", "claveTest", false, t)

	res := realizarPeticionAPI("enviarSolicitudAmistad",
		url.Values{
			"idUsuario":   {strconv.Itoa(idAmigoTest1)},
			"nombreAmigo": {"NombreTestAmigo2"},
			"clave":       {"claveTest"},
		}, t)
	if res["code"].(float64) != 0 {
		t.Fatal(res)
	}

	res = realizarPeticionAPI("notificaciones",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest2)},
			"clave":     {"claveTest"},
		}, t)

	idEnvio := int(res["notificaciones"].([]interface{})[0].(map[string]interface{})["idEnvio"].(float64))
	if idEnvio != idAmigoTest1 {
		t.Log(res)
		t.Fatal("No se ha recibido la notificacion correcta")
	}

	borrarCuenta(idAmigoTest1, "claveTest", t)
	borrarCuenta(idAmigoTest2, "claveTest", t)

}

func enviarSolicitud(id int, amigo, clave string, t *testing.T) {
	res := realizarPeticionAPI("enviarSolicitudAmistad",
		url.Values{
			"idUsuario":   {strconv.Itoa(id)},
			"nombreAmigo": {amigo},
			"clave":       {clave},
		}, t)
	if res["code"].(float64) != 0 {
		t.Fatal(res)
	}
}

func Test_AceptarSolicitudAmistad(t *testing.T) {
	idAmigoTest1 := crearCuenta("NombreTestAmigo1", "", "claveTest", false, t)
	idAmigoTest2 := crearCuenta("NombreTestAmigo2", "", "claveTest", false, t)

	enviarSolicitud(idAmigoTest1, "NombreTestAmigo2", "claveTest", t)

	res := realizarPeticionAPI("gestionAmistad",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest2)},
			"idAmigo":   {strconv.Itoa(idAmigoTest1)},
			"clave":     {"claveTest"},
			"decision":  {"Aceptar"},
		}, t)
	if res["code"].(float64) != 0 {
		t.Fatal(res)
	}

	res = realizarPeticionAPI("amigos",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest1)},
			"clave":     {"claveTest"},
		}, t)
	if res["amigos"].([]interface{})[0].(map[string]interface{})["nombre"].(string) != "NombreTestAmigo2" {
		t.Fatal(res)
	}

	res = realizarPeticionAPI("amigos",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest2)},
			"clave":     {"claveTest"},
		}, t)
	if res["amigos"].([]interface{})[0].(map[string]interface{})["nombre"].(string) != "NombreTestAmigo1" {
		t.Fatal(res)
	}

	res = realizarPeticionAPI("notificaciones",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest2)},
			"clave":     {"claveTest"},
		}, t)
	if res["notificaciones"] != nil {
		t.Error(res)
	}

	borrarCuenta(idAmigoTest1, "claveTest", t)
	borrarCuenta(idAmigoTest2, "claveTest", t)
}

func Test_RechazarSolicitudAmistad(t *testing.T) {
	idAmigoTest1 := crearCuenta("NombreTestAmigo1", "", "claveTest", false, t)
	idAmigoTest2 := crearCuenta("NombreTestAmigo2", "", "claveTest", false, t)

	enviarSolicitud(idAmigoTest1, "NombreTestAmigo2", "claveTest", t)

	res := realizarPeticionAPI("gestionAmistad",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest2)},
			"idAmigo":   {strconv.Itoa(idAmigoTest1)},
			"clave":     {"claveTest"},
			"decision":  {"Rechazar"},
		}, t)
	if res["code"].(float64) != 0 {
		t.Fatal(res)
	}

	res = realizarPeticionAPI("amigos",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest1)},
			"clave":     {"claveTest"},
		}, t)
	if res["amigos"] != nil {
		t.Fatal(res)
	}

	res = realizarPeticionAPI("amigos",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest2)},
			"clave":     {"claveTest"},
		}, t)
	if res["amigos"] != nil {
		t.Fatal(res)
	}

	res = realizarPeticionAPI("notificaciones",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest2)},
			"clave":     {"claveTest"},
		}, t)
	if res["notificaciones"] != nil {
		t.Error(res)
	}

	borrarCuenta(idAmigoTest1, "claveTest", t)
	borrarCuenta(idAmigoTest2, "claveTest", t)
}

func GestionAmistad(id1, id2 int, clave, decision string, t *testing.T) {
	res := realizarPeticionAPI("gestionAmistad",
		url.Values{
			"idUsuario": {strconv.Itoa(id1)},
			"idAmigo":   {strconv.Itoa(id2)},
			"clave":     {clave},
			"decision":  {decision},
		}, t)
	if res["code"].(float64) != 0 {
		t.Fatal(res)
	}
}

func Test_aux(t *testing.T) {
	res := realizarPeticionAPI("borrarCuenta",
		url.Values{
			"idUsuario": {strconv.Itoa(45)},
			"clave":     {"claveTest"},
		}, t)

	if res["code"].(float64) != 0 {
		t.Fatal(res)
	}
}
