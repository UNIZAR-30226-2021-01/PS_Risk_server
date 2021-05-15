package baseDatos

import (
	"PS_Risk_server/mensajes"
	"errors"
	"strconv"
)

/*
	Usuario contiene los datos de una cuenta de un usuario.
*/
type Usuario struct {
	Id            int    `json:"id"`
	Icono         int    `json:"icono"`
	Aspecto       int    `json:"aspecto"`
	Riskos        int    `json:"riskos"`
	Nombre        string `json:"nombre"`
	Correo        string `json:"correo"`
	Clave         string `json:"clave"`
	RecibeCorreos bool   `json:"recibeCorreos"`
}

/*
	ToJSON devuelve el usuario en formato JSON.

	CAMBIAR PARA USAR EL MAP ENCODER
*/
func (u *Usuario) ToJSON() mensajes.JsonData {
	return mensajes.JsonData{
		"id":            u.Id,
		"nombre":        u.Nombre,
		"correo":        u.Correo,
		"recibeCorreos": u.RecibeCorreos,
		"icono":         u.Icono,
		"aspecto":       u.Aspecto,
		"riskos":        u.Riskos,
	}
}

/*
	Modificar modifica el valor de un campo de la estructura de datos utilizando
	el nombre del campo y el valor codificado como un string.
*/
func (u *Usuario) Modificar(c string, v string) error {
	var err error
	switch c {
	case "Aspecto":
		u.Aspecto, err = strconv.Atoi(v)
		if err != nil {
			return err
		}
	case "Icono":
		u.Icono, err = strconv.Atoi(v)
		if err != nil {
			return err
		}
	case "Correo":
		u.Correo = v
	case "Clave":
		u.Clave = v
	case "Nombre":
		u.Nombre = v
	case "RecibeCorreos":
		u.RecibeCorreos, err = strconv.ParseBool(v)
		if err != nil {
			return err
		}
	default:
		return errors.New("el campo a modificar no existe")
	}
	return nil
}
