package test

import (
	"PS_Risk_server/baseDatos"
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
	esperado := partidas.NuevaPartida(1, u.Id, nombreSalaTest, nil)
	if distintas(p, esperado) {
		t.Fatalf("No coincide lo recibido %v con lo esperado %v", p, esperado)
	}
}

func distintas(p1, p2 partidas.Partida) bool {
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
