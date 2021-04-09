package test

import (
	"PS_Risk_server/baseDatos"
	"PS_Risk_server/mensajes"
	"database/sql"
	"fmt"
	"testing"
)

func TestEnviarSolicitudAmistad(t *testing.T) {
	bd, err := sql.Open("postgres", baseDeDatos)
	if err != nil {
		t.Fatalf("No se ha podido abrir conexión con la base de "+
			"datos.\nError:%q", err)
	}
	daoAmigos := baseDatos.NuevoAmigosDAO(bd)
	daoUsuario := baseDatos.NuevoUsuarioDAO(bd)
	u1, err := daoUsuario.CrearCuenta(nombreAmigo1, correoAmigo1, claveTest,
		recibeCorreosTest)
	if err != nil {
		// Si da error es posible que sea porque el usuario ya existe
		u1, err = daoUsuario.IniciarSesionNombre(nombreAmigo1, claveTest)
		if err != nil {
			t.Fatalf("No se ha podido validar ni crear al usuario amigo1.\n"+
				"Error:%q", err)
		}
	}
	// Crear el usuario amigo2 solo si no existe. Si no puede crearlo porque ya
	// existe, intentar iniciar sesión (sus datos se necesitan para comprobar
	// las notificaciones)
	u2, err := daoUsuario.CrearCuenta(nombreAmigo2, correoAmigo2, claveTest,
		recibeCorreosTest)
	if err != nil {
		u2, err = daoUsuario.IniciarSesionNombre(nombreAmigo2, claveTest)
		if err != nil {
			t.Fatalf("No se ha podido validar ni crear al usuario amigo2.\n"+
				"Error:%q", err)
		}
	}
	obtenido := daoAmigos.EnviarSolicitudAmistad(u1, nombreAmigo2)
	if c, _ := coincideErrorNulo(obtenido); !c {
		t.Fatalf("Error %q, al enviar solicitud de amistad", obtenido)
	}
	obtenido = daoUsuario.ObtenerNotificaciones(u2)
	esperado := mensajes.JsonData{
		"notificaciones": []mensajes.JsonData{
			mensajes.NotificacionJson(u1.Id, "Peticion de amistad", nombreAmigo1),
		},
	}
	if !comprobarJson(obtenido, esperado) {
		t.Fatalf("ObtenerNotificaciones() = %q, se esperaba %q", obtenido, esperado)
	}
}

func TestAceptarSolicitudAmistad(t *testing.T) {
	bd, err := sql.Open("postgres", baseDeDatos)
	if err != nil {
		t.Fatalf("No se ha podido abrir conexión con la base de "+
			"datos.\nError:%q", err)
	}
	daoAmigos := baseDatos.NuevoAmigosDAO(bd)
	daoUsuario := baseDatos.NuevoUsuarioDAO(bd)
	u1, err := daoUsuario.IniciarSesionNombre(nombreAmigo1, claveTest)
	if err != nil {
		t.Fatalf("Error iniciando sesión: %q\n", err)
	}
	u2, err := daoUsuario.IniciarSesionNombre(nombreAmigo2, claveTest)
	if err != nil {
		t.Fatalf("Error iniciando sesión: %q\n", err)
	}
	// Comprobar que se borran las dos solicitudes al aceptar una
	// Es posible que ya existan y den error
	envioSolicitud := daoAmigos.EnviarSolicitudAmistad(u1, nombreAmigo2)
	if envioSolicitud["code"] == baseDatos.ErrorAmistadDuplicada {
		t.Fatal(envioSolicitud["err"])
	}
	envioSolicitud = daoAmigos.EnviarSolicitudAmistad(u2, nombreAmigo1)
	if envioSolicitud["code"] == baseDatos.ErrorAmistadDuplicada {
		t.Fatal(envioSolicitud["err"])
	}
	obtenido := daoAmigos.AceptarSolicitudAmistad(u2, u1.Id)
	if c, _ := coincideErrorNulo(obtenido); !c {
		t.Fatalf("Error %q, al aceptar solicitud de amistad", obtenido)
	}
	esperado := mensajes.JsonData{
		"amigos": []mensajes.JsonData{
			mensajes.AmigoJson(u2.Id, 0, 0, nombreAmigo2),
		},
	}
	obtenido = daoAmigos.ObtenerAmigos(u1)
	if !comprobarJson(obtenido, esperado) {
		t.Fatalf("ObtenerAmigos() = %q, se esperaba %q", obtenido, esperado)
	}
	esperado = mensajes.JsonData{
		"amigos": []mensajes.JsonData{
			mensajes.AmigoJson(u1.Id, 0, 0, nombreAmigo1),
		},
	}
	obtenido = daoAmigos.ObtenerAmigos(u2)
	if !comprobarJson(obtenido, esperado) {
		t.Fatalf("ObtenerAmigos() = %q, se esperaba %q", obtenido, esperado)
	}
	esperado = mensajes.JsonData{
		"notificaciones": nil,
	}
	obtenido = daoUsuario.ObtenerNotificaciones(u1)
	if !comprobarJson(obtenido, esperado) {
		t.Fatalf("ObtenerNotificaciones() = %q, se esperaba %q", obtenido, esperado)
	}
	obtenido = daoUsuario.ObtenerNotificaciones(u2)
	if !comprobarJson(obtenido, esperado) {
		t.Fatalf("ObtenerNotificaciones() = %q, se esperaba %q", obtenido, esperado)
	}
}

func TestEliminarAmigo(t *testing.T) {
	bd, err := sql.Open("postgres", baseDeDatos)
	if err != nil {
		t.Fatalf("No se ha podido abrir conexión con la base de"+
			"datos.\nError:%q", err)
	}
	daoAmigos := baseDatos.NuevoAmigosDAO(bd)
	daoUsuario := baseDatos.NuevoUsuarioDAO(bd)
	u1, err := daoUsuario.IniciarSesionNombre(nombreAmigo1, claveTest)
	if err != nil {
		t.Fatalf("Error iniciando sesión: %q\n", err)
	}
	u2, err := daoUsuario.IniciarSesionNombre(nombreAmigo2, claveTest)
	if err != nil {
		t.Fatalf("Error iniciando sesión: %q\n", err)
	}
	obtenido := daoAmigos.EliminarAmigo(u1, u2.Id)
	coincide, esperado := coincideErrorNulo(obtenido)
	if !coincide {
		t.Fatalf("EliminarAmigo() = %q, se esperaba %q", obtenido, esperado)
	}
}

func TestRechazarSolicitudAmistad(t *testing.T) {
	fmt.Println("d")
	bd, err := sql.Open("postgres", baseDeDatos)
	if err != nil {
		t.Fatalf("No se ha podido abrir conexión con la base de "+
			"datos.\nError:%q", err)
	}
	daoAmigos := baseDatos.NuevoAmigosDAO(bd)
	daoUsuario := baseDatos.NuevoUsuarioDAO(bd)
	u1, err := daoUsuario.IniciarSesionNombre(nombreAmigo1, claveTest)
	if err != nil {
		t.Fatalf("Error iniciando sesión: %q\n", err)
	}
	u2, err := daoUsuario.IniciarSesionNombre(nombreAmigo2, claveTest)
	if err != nil {
		t.Fatalf("Error iniciando sesión: %q\n", err)
	}
	envioSolicitud := daoAmigos.EnviarSolicitudAmistad(u2, nombreAmigo1)
	if envioSolicitud["code"] == baseDatos.ErrorAmistadDuplicada {
		t.Fatal(envioSolicitud["err"])
	}
	obtenido := daoAmigos.RechazarSolicitudAmistad(u1, u2.Id)
	coincide, esperado := coincideErrorNulo(obtenido)
	if !coincide {
		t.Fatalf("RechazarSolicitudAmistad() = %q, se esperaba %q", obtenido, esperado)
	}
	esperado = mensajes.JsonData{
		"notificaciones": nil,
	}
	obtenido = daoUsuario.ObtenerNotificaciones(u1)
	if !comprobarJson(obtenido, esperado) {
		t.Fatalf("ObtenerNotificaciones() = %q, se esperaba %q", obtenido, esperado)
	}
	esperado = mensajes.JsonData{
		"amigos": nil,
	}
	obtenido = daoAmigos.ObtenerAmigos(u2)
	if !comprobarJson(obtenido, esperado) {
		t.Fatalf("ObtenerAmigos() = %q, se esperaba %q", obtenido, esperado)
	}
}
