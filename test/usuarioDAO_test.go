package test

import (
	"PS_Risk_server/baseDatos"
	"database/sql"
	"testing"
)

const (
	baseDeDatos       = "..."
	nombreTest        = "usuario0"
	correoTest        = "correo0@mail.com"
	nombreAmigo1      = "nombre1"
	correoAmigo1      = "correo1@mail.com"
	nombreAmigo2      = "nombre2"
	correoAmigo2      = "correo2@mail.com"
	claveTest         = "6d5074b4bf2b913866157d7674f1eda042c5c614876de876f7512702d2572a06"
	recibeCorreosTest = true
)

func TestUsuarioDAO(t *testing.T) {
	db, err := sql.Open("postgres", baseDeDatos)
	if err != nil {
		t.Errorf("No se ha podido abrir conexión con la base de "+
			"datos.\nError:%q", err)
		return
	}
	ud := baseDatos.NuevoUsuarioDAO(db)
	u1, err := ud.CrearCuenta(nombreTest, correoTest, claveTest, recibeCorreosTest)
	if err != nil {
		t.Errorf("Error creando cuenta %q", err)
		return
	}
	u2, err := ud.ObtenerUsuario(u1.Id, u1.Clave)
	if err != nil {
		t.Errorf("Error no se ha podido obtener usuario %q", err)
		return
	}
	if u1 != u2 {
		t.Errorf("No coincide %v con %v", u1, u2)
		return
	}
}
