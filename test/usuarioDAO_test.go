package test

import (
	"PS_Risk_server/baseDatos"
	"database/sql"
	"testing"
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
