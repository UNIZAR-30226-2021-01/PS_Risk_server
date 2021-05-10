/*
	El paquete mensajes proporciona un tipo de dato para leer, almacenar y
	manipular los mensajes en formato JSON que se intercambian con los clientes,
	funciones para crear un mensaje a partir de los datos que debe contener, y
	los códigos de los distintos tipos de error que se pueden producir.
*/
package mensajes

// Códigos de error posibles
const (
	NoError       = 0
	ErrorPeticion = iota
	ErrorUsuario  = iota
	CierreSala    = iota
)

type JsonData map[string]interface{}

/*
	UsuarioJson devuelve un usuario en formato json con los datos proporcionados.
*/
func UsuarioJson(id, icono, aspecto, riskos int, nombre, correo, clave string,
	recibeCorreos bool) JsonData {
	return JsonData{
		"id":            id,
		"nombre":        nombre,
		"correo":        correo,
		"clave":         clave,
		"recibeCorreos": recibeCorreos,
		"icono":         icono,
		"aspecto":       aspecto,
		"riskos":        riskos,
	}
}

/*
	CosmeticoJson devuelve un cosmetico (icono o aspecto) en formato json con
	los datos proporcionados.
*/
func CosmeticoJson(id, precio int) JsonData {
	return JsonData{
		"id":     id,
		"precio": precio,
	}
}

/*
	ErrorJson devuelve un error en formato json con la información proporcionada.
*/
func ErrorJson(e string, c int) JsonData {
	return JsonData{
		"err":  e,
		"code": c,
	}
}

/*
	ErrorJsonPartida devuelve un error en formato json con la información
	proporcionada y un campo "_tipoMensaje".
*/
func ErrorJsonPartida(e string, c int) JsonData {
	return JsonData{
		"_tipoMensaje": "e",
		"err":          e,
		"code":         c,
	}
}

/*
	AmigoJson devuelve en formato json los datos de un usuario que se muestran
	a sus amigos.
*/
func AmigoJson(id, icono, aspecto int, nombre string) JsonData {
	return JsonData{
		"id":      id,
		"nombre":  nombre,
		"icono":   icono,
		"aspecto": aspecto,
	}
}

/*
	NotificacionJson devuelve una notificación en formato json con la
	información proporcionada.
*/
func NotificacionJson(idE int, tipo, info string) JsonData {
	return JsonData{
		"infoExtra": info,
		"tipo":      tipo,
		"idEnvio":   idE,
	}
}
