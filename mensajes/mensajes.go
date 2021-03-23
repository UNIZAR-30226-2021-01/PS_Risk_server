package mensajes

type JsonData map[string]interface{}

func UsuarioJson(id, icono, aspecto, riskos int, nombre, correo, clave string, recibeCorreos bool) JsonData {
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

func CosmeticoJson(id, precio int) JsonData {
	return JsonData{
		"id":     id,
		"precio": precio,
	}
}

func ErrorJson(e string, c int) JsonData {
	return JsonData{
		"err":  e,
		"code": c,
	}
}

func AmigoJson(id, icono, aspecto int, nombre string) JsonData {
	return JsonData{
		"id":      id,
		"nombre":  nombre,
		"icono":   icono,
		"aspecto": aspecto,
	}
}

func NotificacionJson(idE int, tipo, info string) JsonData {
	return JsonData{
		"infoExtra": info,
		"tipo":      tipo,
		"idEnvio":   idE,
	}
}
