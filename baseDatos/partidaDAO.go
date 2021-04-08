package baseDatos

import (
	"PS_Risk_server/mensajes"
	"PS_Risk_server/mensajesInternos"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
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

func (dao *PartidaDAO) CrearPartida(creador Usuario, tiempoTurno int,
	nombreSala string, wsCreador *websocket.Conn) (*Partida, error) {
	var idPartida int
	err := dao.bd.QueryRow(crearPartida, creador.Id, nombreSala, []byte(`{}`)).Scan(&idPartida)
	if err != nil {
		return nil, err
	}
	p := &Partida{
		IdPartida:   idPartida,
		IdCreador:   creador.Id,
		TiempoTurno: tiempoTurno,
		TurnoActual: 0,
		Fase:        0,
		Nombre:      nombreSala,
		Empezada:    false,
		Territorios: []Territorio{},
		Jugadores:   []Jugador{},
		Conexiones:  sync.Map{},
		Mensajes:    make(chan mensajesInternos.MensajePartida, MaxMensajes),
	}
	p.Jugadores = append(p.Jugadores, crearJugador(creador))
	p.Conexiones.Store(creador.Id, wsCreador)
	return p, nil
}

func (dao *PartidaDAO) IniciarPartida(p *Partida, idCreador int) mensajes.JsonData {
	var res mensajes.JsonData
	err := p.IniciarPartida(idCreador)
	if err != nil {
		return mensajes.ErrorJsonPartida(err.Error(), ErrorIniciarPartida)
	}
	ctx := context.Background()
	tx, err := dao.bd.BeginTx(ctx, nil)
	if err != nil {
		p.AnularInicio()
		return mensajes.ErrorJsonPartida(err.Error(), ErrorIniciarPartida)
	}
	_, err = tx.ExecContext(ctx, borrarInvitaciones, p.IdPartida)
	if err != nil {
		tx.Rollback()
		p.AnularInicio()
		return mensajes.ErrorJsonPartida(err.Error(), ErrorIniciarPartida)
	}
	estadoJson, _ := json.Marshal(p)
	_, err = tx.ExecContext(ctx, actualizarEstadoPartida, estadoJson, p.IdPartida)
	if err != nil {
		tx.Rollback()
		p.AnularInicio()
		return mensajes.ErrorJsonPartida(err.Error(), ErrorIniciarPartida)
	}
	for _, j := range p.Jugadores {
		_, err = tx.ExecContext(ctx, guardarJugadores, p.IdPartida, j)
		if err != nil {
			tx.Rollback()
			p.AnularInicio()
			return mensajes.ErrorJsonPartida(err.Error(), ErrorIniciarPartida)
		}
	}
	err = mapstructure.Decode(p, &res)
	if err != nil {
		tx.Rollback()
		p.AnularInicio()
		return mensajes.ErrorJsonPartida(err.Error(), ErrorIniciarPartida)
	}
	err = tx.Commit()
	if err != nil {
		p.AnularInicio()
		return mensajes.ErrorJsonPartida(err.Error(), ErrorIniciarPartida)
	}
	res["_tipoMensaje"] = "p"
	return res
}

func (dao *PartidaDAO) InvitarPartida(p *Partida, idCreador int, idInvitado int) error {
	if p.Empezada {
		return errors.New("no se puede invitar a nadie a una partida que ya ha empezado")
	}
	if p.IdCreador != idCreador {
		return errors.New("solo el creador de la partida puede invitar a otros jugadores")
	}
	if p.EstaEnPartida(idInvitado) {
		return errors.New("no se puede invitar a alguien que ya está en la partida")
	}
	id1 := min(idCreador, idInvitado)
	id2 := max(idCreador, idInvitado)
	err := dao.bd.QueryRow(consultaAmistad, id1, id2).Scan(&id1)
	if err != nil && err == sql.ErrNoRows {
		return errors.New("no se puede invitar a una partida a alguien que no es amigo")
	} else if err != nil {
		return err
	}
	_, err = dao.bd.Exec(crearInvitacion, idInvitado, p.IdPartida)
	return err
}

func (dao *PartidaDAO) EntrarPartida(p *Partida, u Usuario, ws *websocket.Conn) mensajes.JsonData {
	var (
		numInvitaciones int
		res             mensajes.JsonData
	)
	err := dao.bd.QueryRow(consultaInvitacion, u.Id, p.IdPartida).Scan(&numInvitaciones)
	if err != nil {
		return mensajes.ErrorJsonPartida(err.Error(), ErrorUnirsePartida)
	}
	if numInvitaciones == 0 {
		return mensajes.ErrorJsonPartida("No puedes unirte a una partida sin ser invitado",
			ErrorFaltaPermisoUnirse)
	}
	err = p.EntrarPartida(u, ws)
	if err != nil {
		return mensajes.ErrorJsonPartida(err.Error(), ErrorUnirsePartida)
	}
	mapstructure.Decode(p, &res)
	res["_tipoMensaje"] = "d"
	return res
}

func (dao *PartidaDAO) BorrarPartida(p *Partida) error {
	_, err := dao.bd.Exec(borrarPartida, p.IdPartida)
	return err
}

func (dao *PartidaDAO) AbandonarPartida(p *Partida, IdUsuario int) error {
	if p.IdCreador == IdUsuario {
		return dao.BorrarPartida(p)
	}
	err := p.ExpulsarDePartida(IdUsuario)
	if err != nil {
		return err
	}
	return nil
}
