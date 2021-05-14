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

const (
	nombre1 = "NombreTest"
	nombre2 = "NombreTest2"
	nombre3 = "NombreTest3"
	correo1 = "780448@unizar.es"
	correo2 = "780378@unizar.es"
	correo3 = "779333@unizar.es"
	clave1  = "claveTest"
	clave2  = "claveTest2"
)

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
	// Usuario con correo que no recibe
	res := realizarPeticionAPI("registrar",
		url.Values{
			"nombre":        {nombre1},
			"correo":        {""},
			"clave":         {clave1},
			"recibeCorreos": {"false"},
		}, t)

	id := int(res["usuario"].(map[string]interface{})["id"].(float64))
	usuarioTest := map[string]interface{}{
		"id":            id,
		"nombre":        nombre1,
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
	comprobarCamposExtraUsuarioDefecto(res, t)

	borrarCuenta(id, clave1, t)
}

func Test_CrearCuentaRestoCasos(t *testing.T) {
	// Usuario con correo que no recibe
	res := realizarPeticionAPI("registrar",
		url.Values{
			"nombre":        {nombre1},
			"correo":        {correo1},
			"clave":         {clave1},
			"recibeCorreos": {"false"},
		}, t)

	id1 := int(res["usuario"].(map[string]interface{})["id"].(float64))
	usuarioTest := map[string]interface{}{
		"id":            id1,
		"nombre":        "NombreTest",
		"icono":         0,
		"aspecto":       0,
		"correo":        correo1,
		"riskos":        1000,
		"recibeCorreos": false,
	}

	if !comprobarJson(res["usuario"].(map[string]interface{}), usuarioTest) {
		t.Log(usuarioTest)
		t.Log(res["usuario"].(map[string]interface{}))
		t.Fatal("No coinciden los usuarios")
	}

	// Usuario con correo que recibe
	res = realizarPeticionAPI("registrar",
		url.Values{
			"nombre":        {nombre2},
			"correo":        {correo2},
			"clave":         {clave1},
			"recibeCorreos": {"true"},
		}, t)

	id2 := int(res["usuario"].(map[string]interface{})["id"].(float64))
	usuarioTest["id"] = id2
	usuarioTest["nombre"] = nombre2
	usuarioTest["correo"] = correo2
	usuarioTest["recibeCorreos"] = true
	if !comprobarJson(res["usuario"].(map[string]interface{}), usuarioTest) {
		t.Log(usuarioTest)
		t.Log(res["usuario"].(map[string]interface{}))
		t.Fatal("No coinciden los usuarios")
	}

	// Casos que deben dar error
	crearCuentasIncorrectas(t)

	borrarCuenta(id1, clave1, t)
	borrarCuenta(id2, clave1, t)
}

func crearCuentasIncorrectas(t *testing.T) {
	claveDemasiadoLarga := "a"
	for len(claveDemasiadoLarga) <= 64 {
		claveDemasiadoLarga = claveDemasiadoLarga + claveDemasiadoLarga
	}
	crearCuentaError("", correo3, clave1, true, t)
	crearCuentaError("12345678901234567890b", correo3, clave1, true, t)
	crearCuentaError("nombre@invalido", correo3, clave1, true, t)
	crearCuentaError(nombre1, correo3, clave1, true, t)
	crearCuentaError(nombre3, "correo@invalido", clave1, true, t)
	crearCuentaError(nombre3, "", clave1, true, t)
	crearCuentaError(nombre3, correo1, clave1, true, t)
	crearCuentaError(nombre3, correo3, "", true, t)
	crearCuentaError(nombre3, correo3, claveDemasiadoLarga, true, t)

	// Valor no parseable a bool en recibeCorreos
	res := realizarPeticionAPI("registrar",
		url.Values{
			"nombre":        {nombre3},
			"correo":        {correo3},
			"clave":         {clave1},
			"recibeCorreos": {"k"},
		}, t)

	if _, ok := res["code"].(float64); !ok {
		borrarCuenta(int(res["usuario"].(map[string]interface{})["id"].(float64)),
			clave1, t)
		t.Log(res)
		t.Fatal("No ha habido error al crear la cuenta")
	}
}

func crearCuentaError(nombre, correo, clave string, recibeCorreos bool,
	t *testing.T) /*map[string]interface{}*/ {
	res := realizarPeticionAPI("registrar",
		url.Values{
			"nombre":        {nombre},
			"correo":        {correo},
			"clave":         {clave},
			"recibeCorreos": {strconv.FormatBool(recibeCorreos)},
		}, t)

	if _, ok := res["code"].(float64); !ok {
		borrarCuenta(int(res["usuario"].(map[string]interface{})["id"].(float64)),
			clave1, t)
		t.Log(res)
		t.Fatal("No ha habido error al crear la cuenta")
	}

	//return res
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
		t.Log(res)
		t.Fatal("Error al borrar la cuenta")
	}
}

func Test_RecargarUsuario(t *testing.T) {
	id := crearCuenta(nombre1, "", clave1, false, t)

	res := realizarPeticionAPI("recargarUsuario",
		url.Values{
			"idUsuario": {strconv.Itoa(id)},
			"clave":     {clave1},
		}, t)

	usuarioTest := map[string]interface{}{
		"id":            id,
		"nombre":        nombre1,
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

	// Peticiones inválidas
	res = realizarPeticionAPI("recargarUsuario",
		url.Values{
			"idUsuario": {strconv.Itoa(-2)},
			"clave":     {clave1},
		}, t)
	if _, ok := res["code"]; !ok {
		t.Log(res)
		t.Fatal("Se ha recargado un usuario que no debería existir")
	}

	res = realizarPeticionAPI("recargarUsuario",
		url.Values{
			"idUsuario": {"a"},
			"clave":     {clave1},
		}, t)
	if _, ok := res["code"]; !ok {
		t.Log(res)
		t.Fatal("Se ha recargado un usuario con id no numérico")
	}

	res = realizarPeticionAPI("recargarUsuario",
		url.Values{
			"idUsuario": {strconv.Itoa(id)},
			"clave":     {""},
		}, t)
	if _, ok := res["code"]; !ok {
		t.Log(res)
		t.Fatal("Se ha recargado un usuario con una clave que no corresponde")
	}

	borrarCuenta(id, clave1, t)
}

func Test_IniciarSesion(t *testing.T) {
	id := crearCuenta(nombre1, correo1, clave1, false, t)

	res := realizarPeticionAPI("iniciarSesion",
		url.Values{
			"usuario": {nombre1},
			"clave":   {clave1},
		}, t)

	usuarioTest := map[string]interface{}{
		"id":            id,
		"nombre":        nombre1,
		"icono":         0,
		"aspecto":       0,
		"correo":        correo1,
		"riskos":        1000,
		"recibeCorreos": false,
	}
	if !comprobarJson(res["usuario"].(map[string]interface{}), usuarioTest) {
		t.Log(usuarioTest)
		t.Log(res["usuario"].(map[string]interface{}))
		t.Fatal("No coinciden los usuarios")
	}
	comprobarCamposExtraUsuarioDefecto(res, t)

	res = realizarPeticionAPI("iniciarSesion",
		url.Values{
			"usuario": {correo1},
			"clave":   {clave1},
		}, t)
	if !comprobarJson(res["usuario"].(map[string]interface{}), usuarioTest) {
		t.Log(usuarioTest)
		t.Log(res["usuario"].(map[string]interface{}))
		t.Fatal("No coinciden los usuarios")
	}
	comprobarCamposExtraUsuarioDefecto(res, t)

	iniciarSesionError(nombre2, clave1, t)
	iniciarSesionError(correo2, clave1, t)
	iniciarSesionError(nombre1, "claveIncorrecta", t)

	borrarCuenta(id, clave1, t)
}

func iniciarSesionError(usuario, clave string, t *testing.T) {
	res := realizarPeticionAPI("iniciarSesion",
		url.Values{
			"usuario": {usuario},
			"clave":   {clave},
		}, t)

	if _, ok := res["code"]; !ok {
		t.Log(res)
		t.Fatal("No ha habido error al iniciar sesión")
	}
}

func Test_EnviarSolicitudDeAmistad(t *testing.T) {

	idAmigoTest1 := crearCuenta("NombreTestAmigo1", "", clave1, false, t)
	idAmigoTest2 := crearCuenta("NombreTestAmigo2", "", clave1, false, t)

	res := realizarPeticionAPI("enviarSolicitudAmistad",
		url.Values{
			"idUsuario":   {strconv.Itoa(idAmigoTest1)},
			"nombreAmigo": {"NombreTestAmigo2"},
			"clave":       {clave1},
		}, t)
	if res["code"].(float64) != 0 {
		t.Fatal(res)
	}

	res = realizarPeticionAPI("notificaciones",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest2)},
			"clave":     {clave1},
		}, t)

	idEnvio := int(res["notificaciones"].([]interface{})[0].(map[string]interface{})["idEnvio"].(float64))
	if idEnvio != idAmigoTest1 {
		t.Log(res)
		t.Fatal("No se ha recibido la notificacion correcta")
	}

	borrarCuenta(idAmigoTest1, clave1, t)
	borrarCuenta(idAmigoTest2, clave1, t)

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
	idAmigoTest1 := crearCuenta("NombreTestAmigo1", "", clave1, false, t)
	idAmigoTest2 := crearCuenta("NombreTestAmigo2", "", clave1, false, t)

	enviarSolicitud(idAmigoTest1, "NombreTestAmigo2", clave1, t)

	res := realizarPeticionAPI("gestionAmistad",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest2)},
			"idAmigo":   {strconv.Itoa(idAmigoTest1)},
			"clave":     {clave1},
			"decision":  {"Aceptar"},
		}, t)
	if res["code"].(float64) != 0 {
		t.Fatal(res)
	}

	res = realizarPeticionAPI("amigos",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest1)},
			"clave":     {clave1},
		}, t)
	if res["amigos"].([]interface{})[0].(map[string]interface{})["nombre"].(string) != "NombreTestAmigo2" {
		t.Fatal(res)
	}

	res = realizarPeticionAPI("amigos",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest2)},
			"clave":     {clave1},
		}, t)
	if res["amigos"].([]interface{})[0].(map[string]interface{})["nombre"].(string) != "NombreTestAmigo1" {
		t.Fatal(res)
	}

	res = realizarPeticionAPI("notificaciones",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest2)},
			"clave":     {clave1},
		}, t)
	if res["notificaciones"] != nil {
		t.Error(res)
	}

	borrarCuenta(idAmigoTest1, clave1, t)
	borrarCuenta(idAmigoTest2, clave1, t)
}

func Test_DobleSolicitudAmistad(t *testing.T) {
	idAmigoTest1 := crearCuenta("NombreTestAmigo1", "", clave1, false, t)
	idAmigoTest2 := crearCuenta("NombreTestAmigo2", "", clave1, false, t)
	enviarSolicitud(idAmigoTest1, "NombreTestAmigo2", clave1, t)
	enviarSolicitud(idAmigoTest2, "NombreTestAmigo1", clave1, t)
	res := realizarPeticionAPI("amigos",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest1)},
			"clave":     {clave1},
		}, t)
	if res["amigos"].([]interface{})[0].(map[string]interface{})["nombre"].(string) != "NombreTestAmigo2" {
		t.Fatal(res)
	}

	res = realizarPeticionAPI("amigos",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest2)},
			"clave":     {clave1},
		}, t)
	if res["amigos"].([]interface{})[0].(map[string]interface{})["nombre"].(string) != "NombreTestAmigo1" {
		t.Fatal(res)
	}
	borrarCuenta(idAmigoTest1, clave1, t)
	borrarCuenta(idAmigoTest2, clave1, t)
}

func Test_RechazarSolicitudAmistad(t *testing.T) {
	idAmigoTest1 := crearCuenta("NombreTestAmigo1", "", clave1, false, t)
	idAmigoTest2 := crearCuenta("NombreTestAmigo2", "", clave1, false, t)

	enviarSolicitud(idAmigoTest1, "NombreTestAmigo2", clave1, t)

	res := realizarPeticionAPI("gestionAmistad",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest2)},
			"idAmigo":   {strconv.Itoa(idAmigoTest1)},
			"clave":     {clave1},
			"decision":  {"Rechazar"},
		}, t)
	if res["code"].(float64) != 0 {
		t.Fatal(res)
	}

	res = realizarPeticionAPI("amigos",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest1)},
			"clave":     {clave1},
		}, t)
	if res["amigos"] != nil {
		t.Fatal(res)
	}

	res = realizarPeticionAPI("amigos",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest2)},
			"clave":     {clave1},
		}, t)
	if res["amigos"] != nil {
		t.Fatal(res)
	}

	res = realizarPeticionAPI("notificaciones",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest2)},
			"clave":     {clave1},
		}, t)
	if res["notificaciones"] != nil {
		t.Error(res)
	}

	borrarCuenta(idAmigoTest1, clave1, t)
	borrarCuenta(idAmigoTest2, clave1, t)
}

func Test_EliminarAmigo(t *testing.T) {
	idAmigoTest1 := crearCuenta("NombreTestAmigo1", "", clave1, false, t)
	idAmigoTest2 := crearCuenta("NombreTestAmigo2", "", clave1, false, t)
	enviarSolicitud(idAmigoTest1, "NombreTestAmigo2", clave1, t)
	enviarSolicitud(idAmigoTest2, "NombreTestAmigo1", clave1, t)

	res := realizarPeticionAPI("gestionAmistad",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest2)},
			"idAmigo":   {strconv.Itoa(idAmigoTest1)},
			"clave":     {clave1},
			"decision":  {"Borrar"},
		}, t)
	if res["code"].(float64) != 0 {
		t.Fatal(res)
	}

	res = realizarPeticionAPI("amigos",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest1)},
			"clave":     {clave1},
		}, t)
	if res["amigos"] != nil {
		t.Fatal(res)
	}

	res = realizarPeticionAPI("amigos",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest2)},
			"clave":     {clave1},
		}, t)
	if res["amigos"] != nil {
		t.Fatal(res)
	}

	borrarCuenta(idAmigoTest1, clave1, t)
	borrarCuenta(idAmigoTest2, clave1, t)
}

func comprobarCamposExtraUsuarioDefecto(res map[string]interface{}, t *testing.T) {
	if res["iconos"] == nil || len(res["iconos"].([]interface{})) != 1 {
		t.Log(res)
		t.Fatal("Lista de iconos comprados incorrecta")
	}
	if res["aspectos"] == nil || len(res["aspectos"].([]interface{})) != 1 {
		t.Log(res)
		t.Fatal("Lista de aspectos comprados incorrecta")
	}
	if res["tiendaIconos"] == nil || len(res["tiendaIconos"].([]interface{})) != 13 {
		t.Log(res)
		t.Fatal("Lista de iconos disponibles incorrecta")
	}
	if res["tiendaAspectos"] == nil || len(res["tiendaAspectos"].([]interface{})) != 13 {
		t.Log(res)
		t.Fatal("Lista de aspectos disponibles incorrecta")
	}
}

func modificarUsuario(id int, clave, tipo, dato string, t *testing.T) {

	res := realizarPeticionAPI("personalizarUsuario",
		url.Values{
			"idUsuario": {strconv.Itoa(id)},
			"nuevoDato": {dato},
			"clave":     {clave},
			"tipo":      {tipo},
		}, t)

	if res["code"].(float64) != 0 {
		t.Fatal(res)
	}
}

func Test_ModificarUsuario(t *testing.T) {

	id := crearCuenta(nombre1, "", clave1, false, t)
	modificarUsuario(id, clave1, "Nombre", nombre2, t)
	res := realizarPeticionAPI("recargarUsuario",
		url.Values{
			"idUsuario": {strconv.Itoa(id)},
			"clave":     {clave1},
		}, t)
	if res["usuario"].(map[string]interface{})["nombre"].(string) != nombre2 {
		t.Fatal(res)
	}
	modificarUsuario(id, clave1, "Clave", clave2, t)
	res = realizarPeticionAPI("recargarUsuario",
		url.Values{
			"idUsuario": {strconv.Itoa(id)},
			"clave":     {clave2},
		}, t)
	if res["usuario"].(map[string]interface{})["nombre"].(string) != nombre2 {
		t.Fatal(res)
	}
	borrarCuenta(id, clave2, t)

	id = crearCuenta(nombre3, correo1, clave1, false, t)
	modificarUsuario(id, clave1, "Correo", "", t)
	res = realizarPeticionAPI("recargarUsuario",
		url.Values{
			"idUsuario": {strconv.Itoa(id)},
			"clave":     {clave1},
		}, t)
	if res["usuario"].(map[string]interface{})["correo"].(string) != "" {
		t.Fatal(res)
	}
	modificarUsuario(id, clave1, "Correo", correo1, t)
	res = realizarPeticionAPI("recargarUsuario",
		url.Values{
			"idUsuario": {strconv.Itoa(id)},
			"clave":     {clave1},
		}, t)
	if res["usuario"].(map[string]interface{})["correo"].(string) != correo1 {
		t.Fatal(res)
	}
	modificarUsuario(id, clave1, "Correo", correo2, t)
	res = realizarPeticionAPI("recargarUsuario",
		url.Values{
			"idUsuario": {strconv.Itoa(id)},
			"clave":     {clave1},
		}, t)
	if res["usuario"].(map[string]interface{})["correo"].(string) != correo2 {
		t.Fatal(res)
	}
	modificarUsuario(id, clave1, "RecibeCorreos", "true", t)
	res = realizarPeticionAPI("recargarUsuario",
		url.Values{
			"idUsuario": {strconv.Itoa(id)},
			"clave":     {clave1},
		}, t)
	if res["usuario"].(map[string]interface{})["recibeCorreos"].(bool) != true {
		t.Fatal(res)
	}
	borrarCuenta(id, clave1, t)
}

func Test_ComprarEquipar(t *testing.T) {
	id := crearCuenta(nombre1, "", clave1, false, t)

	res := realizarPeticionAPI("comprar",
		url.Values{
			"idUsuario": {strconv.Itoa(id)},
			"cosmetico": {strconv.Itoa(1)},
			"clave":     {clave1},
			"tipo":      {"Icono"},
		}, t)
	if res["code"].(float64) != 0 {
		t.Fatal(res)
	}
	res = realizarPeticionAPI("recargarUsuario",
		url.Values{
			"idUsuario": {strconv.Itoa(id)},
			"clave":     {clave1},
		}, t)
	if len(res["iconos"].([]interface{})) != 2 {
		t.Fatal(res)
	}
	modificarUsuario(id, clave1, "Icono", "1", t)
	res = realizarPeticionAPI("recargarUsuario",
		url.Values{
			"idUsuario": {strconv.Itoa(id)},
			"clave":     {clave1},
		}, t)
	if res["usuario"].(map[string]interface{})["icono"].(float64) != 1 {
		t.Fatal(res)
	}

	res = realizarPeticionAPI("comprar",
		url.Values{
			"idUsuario": {strconv.Itoa(id)},
			"cosmetico": {strconv.Itoa(1)},
			"clave":     {clave1},
			"tipo":      {"Aspecto"},
		}, t)
	if res["code"].(float64) != 0 {
		t.Fatal(res)
	}
	res = realizarPeticionAPI("recargarUsuario",
		url.Values{
			"idUsuario": {strconv.Itoa(id)},
			"clave":     {clave1},
		}, t)
	if len(res["aspectos"].([]interface{})) != 2 {
		t.Fatal(res)
	}
	modificarUsuario(id, clave1, "Aspecto", "1", t)
	res = realizarPeticionAPI("recargarUsuario",
		url.Values{
			"idUsuario": {strconv.Itoa(id)},
			"clave":     {clave1},
		}, t)
	if res["usuario"].(map[string]interface{})["aspecto"].(float64) != 1 {
		t.Fatal(res)
	}

	borrarCuenta(id, clave1, t)
}

func Test_aux(t *testing.T) {
	res := realizarPeticionAPI("borrarCuenta",
		url.Values{
			"idUsuario": {strconv.Itoa(105)},
			"clave":     {clave1},
		}, t)

	if res["code"].(float64) != 0 {
		t.Fatal(res)
	}
}
