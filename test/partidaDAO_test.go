package test

import (
	"PS_Risk_server/baseDatos"
	"PS_Risk_server/mensajes"
	"PS_Risk_server/partidas"
	"database/sql"
	"reflect"
	"testing"
)

func TestCrearPartida(t *testing.T) {
	db, err := sql.Open("postgres", baseDeDatos)
	if err != nil {
		t.Fatalf("No se ha podido abrir conexión con la base de "+
			"datos.\nError:%q", err)
	}
	ud := baseDatos.NuevoUsuarioDAO(db)
	u, err := ud.IniciarSesionNombre(nombreAmigo1, claveTest)
	if err != nil {
		t.Fatalf("Error iniciando sesión %q", err)
	}
	pd := baseDatos.NuevaPartidaDAO(db)
	p, err := pd.CrearPartida(u, tiempoTurnoTest, nombreSalaTest, nil)
	if err != nil {
		t.Fatalf("Error creando partida %q", err)
	}
	esperado := partidas.NuevaPartida(1, u.Id, tiempoTurnoTest, nombreSalaTest, nil)
	if distintas(p, esperado) {
		t.Fatalf("No coincide lo recibido %v con lo esperado %v", p, esperado)
	}
}

func TestObtenerPartida(t *testing.T) {
	db, err := sql.Open("postgres", baseDeDatos)
	if err != nil {
		t.Fatalf("No se ha podido abrir conexión con la base de "+
			"datos.\nError:%q", err)
	}
	ud := baseDatos.NuevoUsuarioDAO(db)
	pd := baseDatos.NuevaPartidaDAO(db)
	u, err := ud.IniciarSesionNombre(nombreAmigo1, claveTest)
	if err != nil {
		t.Fatalf("Error iniciando sesión %q", err)
	}
	p1, err := pd.CrearPartida(u, tiempoTurnoTest, nombreSalaTest, nil)
	if err != nil {
		t.Fatalf("Error creando partida %q", err)
	}
	p2, err := pd.ObtenerPartida(p1.IdPartida)
	if err != nil {
		t.Fatalf("Error leyendo partida de la base de datos %q", err)
	}
	if distintas(p1, p2) {
		t.Fatalf("La partida devuelta por la función crear y la devuelta por"+
			"leer son distintas:\n%v\n%v", p1, p2)
	}
}

func TestInvitarPartida(t *testing.T) {
	db, err := sql.Open("postgres", baseDeDatos)
	if err != nil {
		t.Fatalf("No se ha podido abrir conexión con la base de "+
			"datos.\nError:%q", err)
	}
	ud := baseDatos.NuevoUsuarioDAO(db)
	pd := baseDatos.NuevaPartidaDAO(db)
	p, err := pd.ObtenerPartida(1)
	if err != nil {
		t.Fatalf("Error obteniendo datos de partida %q", err)
	}
	u1, err := ud.ObtenerUsuarioId(p.IdCreador)
	if err != nil {
		t.Fatalf("Error obteniendo datos de usuario %q", err)
	}
	u2, err := ud.IniciarSesionNombre(nombreAmigo2, claveTest)
	if err != nil {
		t.Fatalf("Error iniciando sesión %q", err)
	}
	err = pd.InvitarPartida(p, u1, u2.Id)
	if err != nil {
		t.Fatalf("Error invitando usuario a partida %q", err)
	}
}

func TestEntrarPartida(t *testing.T) {
	db, err := sql.Open("postgres", baseDeDatos)
	if err != nil {
		t.Fatalf("No se ha podido abrir conexión con la base de "+
			"datos.\nError:%q", err)
	}
	ud := baseDatos.NuevoUsuarioDAO(db)
	pd := baseDatos.NuevaPartidaDAO(db)
	p, err := pd.ObtenerPartida(1)
	if err != nil {
		t.Fatalf("Error obteniendo datos de partida %q", err)
	}
	u, err := ud.IniciarSesionNombre(nombreAmigo2, claveTest)
	if err != nil {
		t.Fatalf("Error iniciando sesión %q", err)
	}
	obtenido := pd.EntrarPartida(p, u, nil)
	u1, err := ud.ObtenerUsuarioId(p.IdCreador)
	if err != nil {
		t.Fatalf("Error obteniendo datos de usuario %q", err)
	}
	esperado := mensajes.JsonData{
		"_tipoMensaje": "d",
		"tiempoTurno":  tiempoTurnoTest,
		"nombreSala":   nombreSalaTest,
		"idSala":       1,
		"jugadores": []mensajes.JsonData{
			{
				"id":      u1.Id,
				"nombre":  u1.Nombre,
				"icono":   u1.Icono,
				"aspecto": u1.Aspecto,
			},
			{
				"id":      u.Id,
				"nombre":  u.Nombre,
				"icono":   u.Icono,
				"aspecto": u.Aspecto,
			},
		},
	}
	if !comprobarJson(obtenido, esperado) {
		t.Fatalf("El mensaje esperado: %v\n No coincide con el obtenido: %v", esperado, obtenido)
	}
}

func TestIniciarPartida(t *testing.T) {
	db, err := sql.Open("postgres", baseDeDatos)
	if err != nil {
		t.Fatalf("No se ha podido abrir conexión con la base de "+
			"datos.\nError:%q", err)
	}
	ud := baseDatos.NuevoUsuarioDAO(db)
	pd := baseDatos.NuevaPartidaDAO(db)
	ad := baseDatos.NuevoAmigosDAO(db)
	p, err := pd.ObtenerPartida(1)
	if err != nil {
		t.Fatalf("Error obteniendo datos de partida %q", err)
	}
	u1, err := ud.ObtenerUsuarioId(p.IdCreador)
	if err != nil {
		t.Fatalf("Error obteniendo datos de usuario %q", err)
	}
	u2, err := ud.IniciarSesionNombre(nombreAmigo2, claveTest)
	if err != nil {
		t.Fatalf("Error iniciando sesión %q", err)
	}
	u3, err := ud.ObtenerUsuarioId(1)
	if err != nil {
		t.Fatalf("Error obteniendo datos de usuario %q", err)
	}
	ad.EnviarSolicitudAmistad(u1, u3.Nombre)
	ad.AceptarSolicitudAmistad(u3, u1.Id)
	pd.InvitarPartida(p, u1, u3.Id)
	pd.EntrarPartida(p, u3, nil)
	obtenido := pd.IniciarPartida(p, u1)
	territoriosEsperados := []mensajes.JsonData{}
	for i := 0; i < 42; i++ {
		territoriosEsperados = append(territoriosEsperados, mensajes.JsonData{
			"numJugador": 0,
			"tropas":     0,
		})
	}
	esperado := mensajes.JsonData{
		"_tipoMensaje": "p",
		"tiempoTurno":  tiempoTurnoTest,
		"nombreSala":   nombreSalaTest,
		//"idSala":           1,
		"turnoActual":      0,
		"fase":             0,
		"listaTerritorios": territoriosEsperados,
		"jugadores": []mensajes.JsonData{
			{
				"id":        u1.Id,
				"nombre":    u1.Nombre,
				"icono":     u1.Icono,
				"aspecto":   u1.Aspecto,
				"sigueVivo": true,
			},
			{
				"id":        u2.Id,
				"nombre":    u2.Nombre,
				"icono":     u2.Icono,
				"aspecto":   u2.Aspecto,
				"sigueVivo": true,
			},
			{
				"id":        u3.Id,
				"nombre":    u3.Nombre,
				"icono":     u3.Icono,
				"aspecto":   u3.Aspecto,
				"sigueVivo": true,
			},
		},
	}
	if !comprobarDatosPartida(obtenido, esperado, false) {
		t.Fatalf("El mensaje esperado: %v\n No coincide con el obtenido: %v", esperado, obtenido)
	}
}

func TestBorrarPartida(t *testing.T) {
	db, err := sql.Open("postgres", baseDeDatos)
	if err != nil {
		t.Fatalf("No se ha podido abrir conexión con la base de "+
			"datos.\nError:%q", err)
	}
	pd := baseDatos.NuevaPartidaDAO(db)
	p, err := pd.ObtenerPartida(1)
	if err != nil {
		t.Fatalf("Error obteniendo datos de partida %q", err)
	}
	err = pd.BorrarPartida(p)
	if err != nil {
		t.Fatalf("Error al borrar la partida %q", err)
	}
	p, err = pd.ObtenerPartida(1)
	if err == nil {
		t.Fatalf("La partida sigue existiendo en la base de datos: %v", p)
	}
}

func TestAbandonarPartida(t *testing.T) {
	db, err := sql.Open("postgres", baseDeDatos)
	if err != nil {
		t.Fatalf("No se ha podido abrir conexión con la base de "+
			"datos.\nError:%q", err)
	}
	ud := baseDatos.NuevoUsuarioDAO(db)
	pd := baseDatos.NuevaPartidaDAO(db)
	u1, err := ud.IniciarSesionNombre(nombreAmigo1, claveTest)
	if err != nil {
		t.Fatalf("Error iniciando sesión %q", err)
	}
	u2, err := ud.IniciarSesionNombre(nombreAmigo2, claveTest)
	if err != nil {
		t.Fatalf("Error iniciando sesión %q", err)
	}
	p, err := pd.CrearPartida(u1, tiempoTurnoTest, nombreSalaTest, nil)
	if err != nil {
		t.Fatalf("Error creando partida %q", err)
	}
	pd.InvitarPartida(p, u1, u2.Id)
	pd.EntrarPartida(p, u2, nil)
	err = pd.AbandonarPartida(p, u2)
	if err != nil {
		t.Fatalf("Error eliminando de la partida al segundo jugador %q", err)
	}
	esperado := partidas.NuevaPartida(p.IdPartida, p.IdCreador, tiempoTurnoTest,
		nombreSalaTest, nil)
	if distintas(p, esperado) {
		t.Fatalf("No coinciden los datos de la partida %v con lo esperado %v", p, esperado)
	}
}

func distintas(p1, p2 *partidas.Partida) bool {
	if p1.IdPartida != p2.IdPartida || p1.IdCreador != p2.IdCreador ||
		p1.Empezada != p2.Empezada || p1.Nombre != p2.Nombre ||
		!reflect.DeepEqual(p1.Jugadores, p2.Jugadores) {
		return true
	}
	if p1.Empezada {
		if p1.TiempoTurno != p2.TiempoTurno || p1.Fase != p2.Fase ||
			p1.TurnoActual != p2.TurnoActual ||
			!reflect.DeepEqual(p1.JugadoresVivos, p2.JugadoresVivos) ||
			!reflect.DeepEqual(p1.Territorios, p2.Territorios) {
			return true
		}
	}
	return false
}
