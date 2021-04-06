package baseDatos

import (
	"PS_Risk_server/mensajes"
	"PS_Risk_server/partidas"
	"PS_Risk_server/usuarios"
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/gorilla/websocket"
)

const (
	crearPartida = "INSERT INTO partida (id_creador, nombre, json_estado) " +
		"VALUES ($1, $2, $3) RETURNING id_partida"
	borrarInvitaciones      = "DELETE FROM invitacionPartida WHERE id_envia = $1"
	actualizarEstadoPartida = "UPDATE partida SET json_estado = $1 WHERE " +
		"id_partida = $2"
	crearInvitacion = "INSERT INTO invitacionPartida (id_recibe, id_envia) " +
		"VALUES ($1, $2)"
	consultaInvitacion = "SELECT COUNT(*) FROM invitacionPartida WHERE " +
		"id_recibe = $1 AND id_envia = $2"
	borrarPartida    = "DELETE FROM partida WHERE id_partida = $1"
	guardarJugadores = "INSERT INTO juega (id_partida, id_usuario) VALUES ($1, $2)"
	obtenerPartida   = "SELECT id_creador, json_estado FROM partida " +
		"WHERE id_partida = $1"
)

type PartidaDAO struct {
	bd *sql.DB
}

func NuevaPartidaDAO(bd *sql.DB) PartidaDAO {
	return PartidaDAO{bd: bd}
}

func (dao *PartidaDAO) CrearPartida(creador usuarios.Usuario, tiempoTurno int,
	nombreSala string, wsCreador *websocket.Conn) (*partidas.Partida, error) {
	var p *partidas.Partida
	var idPartida int
	err := dao.bd.QueryRow(crearPartida, creador.Id, nombreSala, []byte(`{}`)).Scan(&idPartida)
	if err != nil {
		return p, err
	}
	p = partidas.NuevaPartida(idPartida, creador.Id, tiempoTurno, nombreSala, wsCreador)
	estado, err := dao.estadoJsonPartida(p, false)
	if err != nil {
		return p, err
	}
	estadoJson, err := json.Marshal(estado)
	if err != nil {
		return p, err
	}
	_, err = dao.bd.Exec(actualizarEstadoPartida, estadoJson, idPartida)
	return p, err
}

func (dao *PartidaDAO) IniciarPartida(p *partidas.Partida,
	u usuarios.Usuario) mensajes.JsonData {
	err := p.IniciarPartida(u.Id)
	if err != nil {
		return mensajes.ErrorJsonPartida(err.Error(), ErrorIniciarPartida)
	}
	estado, err := dao.estadoJsonPartida(p, false)
	if err != nil {
		return mensajes.ErrorJsonPartida(err.Error(), ErrorIniciarPartida)
	}
	estadoJson, err := json.Marshal(estado)
	if err != nil {
		return mensajes.ErrorJsonPartida(err.Error(), ErrorIniciarPartida)
	}
	ctx := context.Background()
	tx, err := dao.bd.BeginTx(ctx, nil)
	if err != nil {
		p.Empezada = false
		return mensajes.ErrorJsonPartida(err.Error(), ErrorIniciarPartida)
	}
	_, err = tx.ExecContext(ctx, borrarInvitaciones, p.IdPartida)
	if err != nil {
		tx.Rollback()
		p.Empezada = false
		return mensajes.ErrorJsonPartida(err.Error(), ErrorIniciarPartida)
	}
	_, err = tx.ExecContext(ctx, actualizarEstadoPartida, estadoJson, p.IdPartida)
	if err != nil {
		tx.Rollback()
		p.Empezada = false
		return mensajes.ErrorJsonPartida(err.Error(), ErrorIniciarPartida)
	}
	for _, j := range p.Jugadores {
		_, err = tx.ExecContext(ctx, guardarJugadores, p.IdPartida, j)
		if err != nil {
			tx.Rollback()
			p.Empezada = false
			return mensajes.ErrorJsonPartida(err.Error(), ErrorIniciarPartida)
		}
	}
	err = tx.Commit()
	if err != nil {
		p.Empezada = false
		return mensajes.ErrorJsonPartida(err.Error(), ErrorIniciarPartida)
	}
	return respuestaInicioPartida(estado)
}

func (dao *PartidaDAO) InvitarPartida(p *partidas.Partida, u usuarios.Usuario,
	idInvitado int) error {
	if p.Empezada {
		return errors.New("no se puede invitar a nadie a una partida que ya ha empezado")
	}
	if p.IdCreador != u.Id {
		return errors.New("solo el creador de la partida puede invitar a otros jugadores")
	}
	if p.EstaEnPartida(idInvitado) {
		return errors.New("no se puede invitar a alguien que ya está en la partida")
	}
	id1 := min(u.Id, idInvitado)
	id2 := max(u.Id, idInvitado)
	err := dao.bd.QueryRow(consultaAmistad, id1, id2).Scan(&id1)
	if err != nil && err == sql.ErrNoRows {
		return errors.New("no se puede invitar a una partida a alguien que no es amigo")
	} else if err != nil {
		return err
	}
	_, err = dao.bd.Exec(crearInvitacion, idInvitado, p.IdPartida)
	return err
}

func (dao *PartidaDAO) EntrarPartida(p *partidas.Partida, u usuarios.Usuario,
	ws *websocket.Conn) mensajes.JsonData {
	var numInvitaciones int
	err := dao.bd.QueryRow(consultaInvitacion, u.Id, p.IdPartida).Scan(&numInvitaciones)
	if err != nil {
		return mensajes.ErrorJsonPartida(err.Error(), ErrorUnirsePartida)
	}
	if numInvitaciones == 0 {
		return mensajes.ErrorJsonPartida("No puedes unirte a una partida sin ser invitado",
			ErrorFaltaPermisoUnirse)
	}
	err = p.EntrarPartida(u.Id, ws)
	if err != nil {
		return mensajes.ErrorJsonPartida(err.Error(), ErrorUnirsePartida)
	}
	estado, err := dao.estadoJsonPartida(p, false)
	if err != nil {
		return mensajes.ErrorJsonPartida(err.Error(), ErrorUnirsePartida)
	}
	estadoJson, err := json.Marshal(estado)
	if err != nil {
		return mensajes.ErrorJsonPartida(err.Error(), ErrorUnirsePartida)
	}
	_, err = dao.bd.Exec(actualizarEstadoPartida, estadoJson, p.IdPartida)
	if err != nil {
		p.AbandonarPartida(u.Id)
		return mensajes.ErrorJsonPartida(err.Error(), ErrorUnirsePartida)
	}
	return respuestaDatosSala(estado)
}

func (dao *PartidaDAO) BorrarPartida(p *partidas.Partida) error {
	_, err := dao.bd.Exec(borrarPartida, p.IdPartida)
	return err
}

func (dao *PartidaDAO) AbandonarPartida(p *partidas.Partida, u usuarios.Usuario) error {
	if p.IdCreador == u.Id {
		return dao.BorrarPartida(p)
	}
	wsInterface, ok := p.Conexiones.Load(u.Id)
	if !ok {
		return errors.New("el usuario no existe en la partida")
	}
	err := p.AbandonarPartida(u.Id)
	if err != nil {
		return err
	}
	ws, tieneWs := wsInterface.(*websocket.Conn)
	if !tieneWs {
		ws = nil
	}
	estado, err := dao.estadoJsonPartida(p, false)
	if err != nil {
		p.EntrarPartida(u.Id, ws)
		return err
	}
	estadoJson, err := json.Marshal(estado)
	if err != nil {
		p.EntrarPartida(u.Id, ws)
		return err
	}
	_, err = dao.bd.Exec(actualizarEstadoPartida, estadoJson, p.IdPartida)
	if err != nil {
		p.EntrarPartida(u.Id, ws)
	}
	return err
}

func (dao *PartidaDAO) ObtenerPartida(idPartida int) (*partidas.Partida, error) {
	var p *partidas.Partida
	var idCreador int
	estadoJson := []byte{}
	err := dao.bd.QueryRow(obtenerPartida, idPartida).Scan(&idCreador, &estadoJson)
	if err != nil {
		return p, err
	}
	estado := mensajes.JsonData{}
	err = json.Unmarshal(estadoJson, &estado)
	if err != nil {
		return p, err
	}
	return partidas.PartidaDesdeJson(estado, idCreador)
}

func (dao *PartidaDAO) listaJugadoresJson(p *partidas.Partida,
	incluirEstadoJugadores bool) ([]mensajes.JsonData, error) {
	jugadores := []mensajes.JsonData{}
	daoUsuario := NuevoUsuarioDAO(dao.bd)
	for _, idJugador := range p.Jugadores {
		u, err := daoUsuario.obtenerUsuarioId(idJugador)
		if err != nil { // No debería ocurrir nunca, se comprueban antes de unirse
			return jugadores, err
		}
		j := mensajes.JsonData{
			"id":      idJugador,
			"nombre":  u.Nombre,
			"icono":   u.Icono,
			"aspecto": u.Aspecto,
		}
		if incluirEstadoJugadores {
			j["sigueVivo"] = p.JugadoresVivos[idJugador]
		}
		jugadores = append(jugadores, j)
	}
	return jugadores, nil
}

func (dao *PartidaDAO) estadoJsonPartida(p *partidas.Partida,
	incluirEstadoJugadores bool) (mensajes.JsonData, error) {
	var respuesta mensajes.JsonData
	jugadores, err := dao.listaJugadoresJson(p, incluirEstadoJugadores)
	if err != nil {
		return respuesta, err
	}
	territorios := []mensajes.JsonData{}
	for _, t := range p.Territorios {
		territorios = append(territorios, t.ToJSON())
	}
	respuesta = mensajes.JsonData{
		"idSala":           p.IdPartida,
		"nombreSala":       p.Nombre,
		"empezada":         p.Empezada,
		"tiempoTurno":      p.TiempoTurno,
		"turnoActual":      p.TurnoActual,
		"fase":             p.Fase,
		"jugadores":        jugadores,
		"listaTerritorios": territorios,
	}
	return respuesta, nil
}

func respuestaDatosSala(estadoCompleto mensajes.JsonData) mensajes.JsonData {
	return mensajes.JsonData{
		"_tipoMensaje": "d",
		"idSala":       estadoCompleto["idSala"],
		"nombreSala":   estadoCompleto["nombreSala"],
		"tiempoTurno":  estadoCompleto["tiempoTurno"],
		"jugadores":    estadoCompleto["jugadores"],
	}
}

func respuestaInicioPartida(estadoCompleto mensajes.JsonData) mensajes.JsonData {
	return mensajes.JsonData{
		"_tipoMensaje":     "p",
		"nombreSala":       estadoCompleto["nombreSala"],
		"fase":             estadoCompleto["fase"],
		"tiempoTurno":      estadoCompleto["tiempoTurno"],
		"turnoActual":      estadoCompleto["turnoActual"],
		"jugadores":        estadoCompleto["jugadores"],
		"listaTerritorios": estadoCompleto["listaTerritorios"],
	}
}
