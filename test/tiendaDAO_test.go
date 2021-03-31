package test

import (
	"PS_Risk_server/baseDatos"
	"PS_Risk_server/mensajes"
	"database/sql"
	"strconv"
	"testing"
)

func TestComprar(t *testing.T) {
	db, err := sql.Open("postgres", baseDeDatos)
	if err != nil {
		t.Fatalf("No se ha podido abrir conexión con la base de "+
			"datos.\nError:%q", err)
	}
	ud := baseDatos.NuevoUsuarioDAO(db)
	td := baseDatos.NuevaTiendaDAO(db)
	u1, err := ud.ObtenerUsuario(1, claveTest)
	if err != nil {
		t.Fatalf("Error " + err.Error())
	}
	u1.Riskos = 2000
	obtenido := ud.ActualizarUsuario(u1)
	coincide, esperado := coincideErrorNulo(obtenido)
	if !coincide {
		t.Fatalf("Error %q, se esperaba %q", obtenido, esperado)
	}
	ti, err := td.ObtenerTienda()
	if err != nil {
		t.Fatalf("Error " + err.Error())
	}
	pre, enc := ti.ObtenerPrecioAspecto(1)
	if !enc {
		t.Fatalf("Aspecto no encontrado")
	}
	obtenido = td.ComprarAspecto(&u1, 1, pre)
	coincide, esperado = coincideErrorNulo(obtenido)
	if !coincide {
		t.Fatalf("Error %q, se esperaba %q", obtenido, esperado)
	}
	pre, enc = ti.ObtenerPrecioIcono(1)
	if !enc {
		t.Fatalf("Aspecto no encontrado")
	}
	obtenido = td.ComprarIcono(&u1, 1, pre)
	coincide, esperado = coincideErrorNulo(obtenido)
	if !coincide {
		t.Fatalf("Error %q, se esperaba %q", obtenido, esperado)
	}
	u1, err = ud.ObtenerUsuario(1, claveTest)
	if err != nil {
		t.Fatalf("Error " + err.Error())
	}
	if u1.Riskos != 1500 {
		t.Fatalf("Error calculando Riskos " + strconv.Itoa(u1.Riskos))
	}
	icns, err := td.ObtenerIconos(u1)
	if err != nil {
		t.Fatalf("Error " + err.Error())
	}
	aspc, err := td.ObtenerAspectos(u1)
	if err != nil {
		t.Fatalf("Error " + err.Error())
	}
	obtenido = mensajes.JsonData{
		"Iconos":   icns,
		"Aspectos": aspc,
	}
	esperado = mensajes.JsonData{
		"Iconos": []mensajes.JsonData{
			{"id": 0, "precio": 0},
			{"id": 1, "precio": 200},
		},
		"Aspectos": []mensajes.JsonData{
			{"id": 0, "precio": 0},
			{"id": 1, "precio": 300},
		},
	}
	if !comprobarJson(obtenido, esperado) {
		t.Fatalf("Tienda = %q, se esperaba %q", obtenido, esperado)
	}
}
