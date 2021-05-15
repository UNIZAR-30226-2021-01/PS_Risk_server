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
		t.Fatal(res)
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

	idAmigoTest1 := crearCuenta(nombre1, "", clave1, false, t)
	idAmigoTest2 := crearCuenta(nombre2, "", clave1, false, t)

	res := realizarPeticionAPI("enviarSolicitudAmistad",
		url.Values{
			"idUsuario":   {strconv.Itoa(idAmigoTest1)},
			"nombreAmigo": {nombre2},
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
		t.Fatal("No se ha recibido la notificación correcta")
	}

	// Enviar la solicitud de vuelta: deben añadirse como amigos
	enviarSolicitud(idAmigoTest2, nombre1, clave1, t)

	res = realizarPeticionAPI("notificaciones",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest2)},
			"clave":     {clave1},
		}, t)

	if res["notificaciones"] != nil {
		t.Log(res)
		t.Fatal("No se ha eliminado correctamente la notificación")
	}

	res = realizarPeticionAPI("amigos",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest1)},
			"clave":     {clave1},
		}, t)
	if _, ok := res["code"]; ok {
		t.Fatal(res)
	}
	if res["amigos"] == nil {
		t.Fatal("Los usuarios no se han añadido como amigos")
	}
	amigo := res["amigos"].([]interface{})[0].(map[string]interface{})
	amigoTest := map[string]interface{}{
		"id":      idAmigoTest2,
		"nombre":  nombre2,
		"icono":   0,
		"aspecto": 0,
	}
	if !comprobarJson(amigo, amigoTest) {
		t.Log(amigo)
		t.Log(amigoTest)
		t.Fatal("Los datos de amigos no coinciden")
	}

	// Casos de error
	enviarSolicitudError(idAmigoTest2+1, nombre1, clave1, t)
	res = realizarPeticionAPI("enviarSolicitudAmistad",
		url.Values{
			"idUsuario":   {"k"},
			"nombreAmigo": {nombre1},
			"clave":       {clave1},
		}, t)
	if res["code"].(float64) == 0 {
		t.Fatal(res)
	}
	enviarSolicitudError(idAmigoTest1, nombre2, "claveIncorrecta", t)
	enviarSolicitudError(idAmigoTest1, nombre3, clave1, t)
	enviarSolicitudError(idAmigoTest1, nombre1, clave1, t)
	enviarSolicitudError(idAmigoTest1, nombre2, clave1, t)

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

func enviarSolicitudError(id int, amigo, clave string, t *testing.T) {
	res := realizarPeticionAPI("enviarSolicitudAmistad",
		url.Values{
			"idUsuario":   {strconv.Itoa(id)},
			"nombreAmigo": {amigo},
			"clave":       {clave},
		}, t)
	if res["code"].(float64) == 0 {
		t.Fatal("Solicitud enviada sin error")
	}
}

func Test_AceptarSolicitudAmistad(t *testing.T) {
	idAmigoTest1 := crearCuenta(nombre1, "", clave1, false, t)
	idAmigoTest2 := crearCuenta(nombre2, "", clave1, false, t)

	enviarSolicitud(idAmigoTest1, nombre2, clave1, t)

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
	if res["amigos"].([]interface{})[0].(map[string]interface{})["nombre"].(string) != nombre2 {
		t.Fatal(res)
	}

	res = realizarPeticionAPI("amigos",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest2)},
			"clave":     {clave1},
		}, t)
	if res["amigos"].([]interface{})[0].(map[string]interface{})["nombre"].(string) != nombre1 {
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

	aceptarSolicitudesIncorrectas(idAmigoTest1, idAmigoTest2, clave1, t)

	borrarCuenta(idAmigoTest1, clave1, t)
	borrarCuenta(idAmigoTest2, clave1, t)
}

func aceptarSolicitudesIncorrectas(idAmigo1, idAmigo2 int, clave string, t *testing.T) {
	aceptarSolicitudError(idAmigo2+1, idAmigo2, clave, t)
	res := realizarPeticionAPI("gestionAmistad",
		url.Values{
			"idUsuario": {"k"},
			"idAmigo":   {strconv.Itoa(idAmigo1)},
			"clave":     {clave1},
			"decision":  {"Aceptar"},
		}, t)
	if res["code"].(float64) == 0 {
		t.Fatal("Se ha aceptado la solicitud con idUsuario=\"k\"")
	}
	res = realizarPeticionAPI("gestionAmistad",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigo2)},
			"idAmigo":   {"k"},
			"clave":     {clave1},
			"decision":  {"Aceptar"},
		}, t)
	if res["code"].(float64) == 0 {
		t.Fatal("Se ha aceptado la solicitud con idAmigo=\"k\"")
	}
	aceptarSolicitudError(idAmigo1, idAmigo2, "clave incorrecta", t)
	aceptarSolicitudError(idAmigo2, idAmigo2, clave1, t)
	res = realizarPeticionAPI("gestionAmistad",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigo2)},
			"idAmigo":   {strconv.Itoa(idAmigo1)},
			"clave":     {clave},
			"decision":  {"incorrecto"},
		}, t)
	if res["code"].(float64) == 0 {
		t.Fatal("No ha dado error un valor incorrecto en decision")
	}

	// Comprobar que no se ha añadido ningún amigo
	res = realizarPeticionAPI("amigos",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigo2)},
			"clave":     {clave},
		}, t)
	if res["amigos"] == nil || len(res["amigos"].([]interface{})) != 1 {
		t.Log(res)
		t.Fatal("La lista de amigos no es la esperada después de aceptar " +
			"solicitudes de amistad incorrectas")
	}
}

func aceptarSolicitudError(id1, id2 int, clave string, t *testing.T) {
	res := realizarPeticionAPI("gestionAmistad",
		url.Values{
			"idUsuario": {strconv.Itoa(id2)},
			"idAmigo":   {strconv.Itoa(id1)},
			"clave":     {clave1},
			"decision":  {"Aceptar"},
		}, t)
	if res["code"].(float64) == 0 {
		t.Fatal("No ha habido error al aceptar la solicitud de amistad")
	}
}

func Test_RechazarSolicitudAmistad(t *testing.T) {
	idAmigoTest1 := crearCuenta(nombre1, "", clave1, false, t)
	idAmigoTest2 := crearCuenta(nombre2, "", clave1, false, t)

	enviarSolicitud(idAmigoTest1, nombre2, clave1, t)

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
		t.Log(res)
		t.Fatal("No ha sido rechazado correctamente")
	}

	res = realizarPeticionAPI("amigos",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest2)},
			"clave":     {clave1},
		}, t)
	if res["amigos"] != nil {
		t.Log(res)
		t.Fatal("No ha sido rechazado correctamente")
	}

	res = realizarPeticionAPI("notificaciones",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest2)},
			"clave":     {clave1},
		}, t)
	if res["notificaciones"] != nil {
		t.Log(res)
		t.Error("No se ha eliminado la solicitud de amistad")
	}

	// Rechazar una solicitud que no existe: error
	res = realizarPeticionAPI("gestionAmistad",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest2)},
			"idAmigo":   {strconv.Itoa(idAmigoTest1)},
			"clave":     {clave1},
			"decision":  {"Rechazar"},
		}, t)
	if res["code"].(float64) == 0 {
		t.Fatal("No ha dado error eliminar una solicitud de amistad inexistente")
	}

	borrarCuenta(idAmigoTest1, clave1, t)
	borrarCuenta(idAmigoTest2, clave1, t)
}

func Test_EliminarAmigo(t *testing.T) {
	idAmigoTest1 := crearCuenta(nombre1, "", clave1, false, t)
	idAmigoTest2 := crearCuenta(nombre2, "", clave1, false, t)

	enviarSolicitud(idAmigoTest1, nombre2, clave1, t)
	gestionAmistad(idAmigoTest2, idAmigoTest1, clave1, "Aceptar", t)

	res := realizarPeticionAPI("gestionAmistad",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest1)},
			"idAmigo":   {strconv.Itoa(idAmigoTest2)},
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
		t.Log(res)
		t.Fatal("No se ha eliminado el amigo de la lista del primer usuario")
	}

	res = realizarPeticionAPI("amigos",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest2)},
			"clave":     {clave1},
		}, t)
	if res["amigos"] != nil {
		t.Log(res)
		t.Fatal("No se ha eliminado el amigo de la lista del segundo usuario")
	}

	// Borrar a alguien que no es amigo: error
	res = realizarPeticionAPI("gestionAmistad",
		url.Values{
			"idUsuario": {strconv.Itoa(idAmigoTest1)},
			"idAmigo":   {strconv.Itoa(idAmigoTest2)},
			"clave":     {clave1},
			"decision":  {"Borrar"},
		}, t)
	if res["code"].(float64) == 0 {
		t.Fatal("No ha dado error eliminar a un amigo que no era amigo")
	}

	borrarCuenta(idAmigoTest1, clave1, t)
	borrarCuenta(idAmigoTest2, clave1, t)
}

func gestionAmistad(id1, id2 int, clave, decision string, t *testing.T) {
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

func modificarUsuarioError(id int, clave, tipo, dato string, t *testing.T) {
	res := realizarPeticionAPI("personalizarUsuario",
		url.Values{
			"idUsuario": {strconv.Itoa(id)},
			"nuevoDato": {dato},
			"clave":     {clave},
			"tipo":      {tipo},
		}, t)

	if res["code"].(float64) == 0 {
		t.Fatal("El usuario se ha modificado sin error")
	}
}

func Test_ModificarUsuarioConDatosIncorrectos(t *testing.T) {
	claveDemasiadoLarga := "a"
	for len(claveDemasiadoLarga) <= 64 {
		claveDemasiadoLarga = claveDemasiadoLarga + claveDemasiadoLarga
	}

	id := crearCuenta(nombre1, correo1, clave1, true, t)

	modificarUsuarioError(id+1, clave1, "Nombre", nombre2, t)

	res := realizarPeticionAPI("personalizarUsuario",
		url.Values{
			"idUsuario": {"k"},
			"nuevoDato": {nombre2},
			"clave":     {clave1},
			"tipo":      {"Nombre"},
		}, t)
	if res["code"].(float64) == 0 {
		t.Fatal("No ha dado error modificar un usuario con id no numérico")
	}

	modificarUsuarioError(id, clave2, "Nombre", nombre2, t)
	modificarUsuarioError(id, clave1, "Tipo incorrecto", nombre2, t)
	modificarUsuarioError(id, clave1, "Nombre", "", t)
	modificarUsuarioError(id, clave1, "Nombre", "1234567890.123456789a", t)
	modificarUsuarioError(id, clave1, "Nombre", "nombre@invalido", t)
	modificarUsuarioError(id, clave1, "Clave", "", t)
	modificarUsuarioError(id, clave1, "Clave", claveDemasiadoLarga, t)
	modificarUsuarioError(id, clave1, "Correo", "", t)
	modificarUsuarioError(id, clave1, "Correo", "correo_invalido", t)

	res = realizarPeticionAPI("personalizarUsuario",
		url.Values{
			"idUsuario": {strconv.Itoa(id)},
			"nuevoDato": {"k"},
			"clave":     {clave1},
			"tipo":      {"RecibeCorreos"},
		}, t)
	if res["code"].(float64) == 0 {
		t.Fatal("No ha dado error modificar un usuario con recibeCorreos no booleano")
	}

	// Modificar recibeCorreos y el correo para probar más casos
	modificarUsuario(id, clave1, "RecibeCorreos", "false", t)
	modificarUsuario(id, clave1, "Correo", "", t)
	modificarUsuarioError(id, clave1, "RecibeCorreos", "true", t)

	// Crear otro usuario para comprobar que se mantiene unicidad en nombre y correo
	id2 := crearCuenta(nombre2, correo2, clave1, false, t)
	modificarUsuarioError(id, clave1, "Nombre", nombre2, t)
	modificarUsuarioError(id, clave1, "Correo", correo2, t)

	borrarCuenta(id, clave1, t)
	borrarCuenta(id2, clave1, t)
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
