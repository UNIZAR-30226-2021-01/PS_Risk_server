package usuarios

import (
	"PS_Risk_server/mensajes"
	"errors"
	"strconv"
)

type Usuario struct {
	Id, Icono, Aspecto, Riskos int
	Nombre, Correo, Clave      string
	RecibeCorreos              bool
}

func (u *Usuario) ToJSON() mensajes.JsonData {
	return mensajes.JsonData{
		"id":            u.Id,
		"nombre":        u.Nombre,
		"correo":        u.Correo,
		"clave":         u.Clave,
		"recibeCorreos": u.RecibeCorreos,
		"icono":         u.Icono,
		"aspecto":       u.Aspecto,
		"riskos":        u.Riskos,
	}
}
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
