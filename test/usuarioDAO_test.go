package test

import (
	"PS_Risk_server/baseDatos"
	"database/sql"
	"testing"
)

const (
	baseDeDatos       = "postgres://ehoyoqkpcgqyrc:c0f640db53a47820e660f19d987b013d7b01d9ac217b669f971070926969dfe0@ec2-54-72-155-238.eu-west-1.compute.amazonaws.com:5432/d3cnglhgpo3rgk"
	nombreTest        = "usuario0"
	correoTest        = "correo0@mail.com"
	nombreAmigo1      = "nombre1"
	correoAmigo1      = "correo1@mail.com"
	nombreAmigo2      = "nombre2"
	correoAmigo2      = "correo2@mail.com"
	claveTest         = "6d5074b4bf2b913866157d7674f1eda042c5c614876de876f7512702d2572a06"
	recibeCorreosTest = true
)

func TestCrearCuenta(t *testing.T) {
	db, err := sql.Open("postgres", baseDeDatos)
	if err != nil {
		t.Fatalf("No se ha podido abrir conexión con la base de "+
			"datos.\nError:%q", err)
	}
	ud := baseDatos.NuevoUsuarioDAO(db)
	u1, err := ud.CrearCuenta(nombreTest, correoTest, claveTest, recibeCorreosTest)
	if err != nil {
		t.Fatalf("Error creando cuenta %q", err)
	}
	u2, err := ud.ObtenerUsuario(u1.Id, u1.Clave)
	if err != nil {
		t.Fatalf("Error no se ha podido obtener usuario %q", err)
	}
	if u1 != u2 {
		t.Fatalf("No coincide %v con %v", u1, u2)
	}
}

func TestIniciarSesion(t *testing.T) {
	db, err := sql.Open("postgres", baseDeDatos)
	if err != nil {
		t.Fatalf("No se ha podido abrir conexión con la base de "+
			"datos.\nError:%q", err)
	}
	ud := baseDatos.NuevoUsuarioDAO(db)
	_, err = ud.IniciarSesionCorreo(correoTest, claveTest)
	if err != nil {
		t.Fatalf("Error no se ha podido iniciar sesion %q", err)
	}
	_, err = ud.IniciarSesionNombre(nombreTest, claveTest)
	if err != nil {
		t.Fatalf("Error no se ha podido iniciar sesion %q", err)
	}
}

func TestModificarUsuario(t *testing.T) {
	nuevoNombre := "nuevo Nombre"
	nuevoCorreo := "nuevoCorreo@mail.com"
	nuevaClave := "82ef67bb06675af2d43639806236ad1189253ce86e6210c072a3a265987df429"
	db, err := sql.Open("postgres", baseDeDatos)
	if err != nil {
		t.Fatalf("No se ha podido abrir conexión con la base de datos.\nError:%q", err)
	}
	ud := baseDatos.NuevoUsuarioDAO(db)
	u1, err := ud.IniciarSesionNombre(nombreTest, claveTest)
	if err != nil {
		t.Fatalf("Error iniciando sesion %q", err)
	}
	u1.Nombre = nuevoNombre
	u1.Correo = nuevoCorreo
	u1.Clave = nuevaClave
	obtenido := ud.ActualizarUsuario(u1)
	if c, _ := coincideErrorNulo(obtenido); !c {
		t.Fatalf("Error obtenido %q", obtenido)
	}
	u2, err := ud.ObtenerUsuario(u1.Id, nuevaClave)
	if err != nil {
		t.Fatalf("Error iniciando sesion %q", err)
	}
	if u1 != u2 {
		t.Fatalf("No coincide %v con %v", u1, u2)
	}
}
