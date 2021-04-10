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

/*
	PartidaDAO permite modificar la tabla de partidas y sus relacionadas en la base de datos.
*/
type PartidaDAO struct {
	bd *sql.DB
}

/*
	NuevaPartidaDAO crea un DAO para partidas.
*/
func NuevaPartidaDAO(bd *sql.DB) PartidaDAO {
	return PartidaDAO{bd: bd}
}

/*
	CrearPartida crea una partida y la guarda en la base de datos.
	Devuelve error en caso de que no se haya podido crear.
*/
func (dao *PartidaDAO) CrearPartida(creador Usuario, tiempoTurno int, nombreSala string,
	wsCreador *websocket.Conn) (*Partida, error) {

	var idPartida int

	// Crea la partida en la base de datos
	err := dao.bd.QueryRow(crearPartida, creador.Id, nombreSala, []byte(`{}`)).Scan(&idPartida)
	if err != nil {
		return nil, err
	}

	// Crea la estructura de datos
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
		Mensajes:    make(chan mensajesInternos.MensajePartida, maxMensajes),
	}
	p.Jugadores = append(p.Jugadores, CrearJugador(creador))
	p.Conexiones.Store(creador.Id, wsCreador)
	return p, nil
}

/*
	IniciarPartida guarda en la base de datos que la partida se ha iniciado y elimina
	todas las invitaciones a la partida. Devuelve el estado de la partida o un error,
	en caso de que no se haya podido iniciar, en formato json.
*/
func (dao *PartidaDAO) IniciarPartida(p *Partida, idCreador int) mensajes.JsonData {

	var res mensajes.JsonData

	// Iniciar la partida
	err := p.IniciarPartida(idCreador)
	if err != nil {
		return mensajes.ErrorJsonPartida(err.Error(), ErrorIniciarPartida)
	}

	// Inicia una transacción en la base de datos
	ctx := context.Background()
	tx, err := dao.bd.BeginTx(ctx, nil)
	if err != nil {
		p.AnularInicio()
		return mensajes.ErrorJsonPartida(err.Error(), ErrorIniciarPartida)
	}
	// Borrar todas las invitaciones a la partida
	_, err = tx.ExecContext(ctx, borrarInvitaciones, p.IdPartida)
	if err != nil {
		tx.Rollback()
		p.AnularInicio()
		return mensajes.ErrorJsonPartida(err.Error(), ErrorIniciarPartida)
	}
	// Actualiza el estado de la partida en la base de datos
	estadoJson, _ := json.Marshal(p)
	_, err = tx.ExecContext(ctx, actualizarEstadoPartida, estadoJson, p.IdPartida)
	if err != nil {
		tx.Rollback()
		p.AnularInicio()
		return mensajes.ErrorJsonPartida(err.Error(), ErrorIniciarPartida)
	}
	// Guardar en la base de datos qué jugadores juegan en la partida
	for _, j := range p.Jugadores {
		_, err = tx.ExecContext(ctx, guardarJugadores, p.IdPartida, j)
		if err != nil {
			tx.Rollback()
			p.AnularInicio()
			return mensajes.ErrorJsonPartida(err.Error(), ErrorIniciarPartida)
		}
	}
	// Codificar los datos de la partida en formato json
	mapstructure.Decode(p, &res)
	// Finalizar la transaccion
	err = tx.Commit()
	if err != nil {
		p.AnularInicio()
		return mensajes.ErrorJsonPartida(err.Error(), ErrorIniciarPartida)
	}
	res["_tipoMensaje"] = "p"
	return res
}

/*
	InvitarPartida guarda en la base de datos que un jugador ha sido invitado a una
	partida. Devuelve error en caso de que no se pueda completar la invitación.
*/
func (dao *PartidaDAO) InvitarPartida(p *Partida, idCreador int, idInvitado int) error {
	// Comprobar si se puede invitar el usuario a la partida
	if p.Empezada {
		return errors.New("no se puede invitar a nadie a una partida que ya ha empezado")
	}
	if p.IdCreador != idCreador {
		return errors.New("solo el creador de la partida puede invitar a otros jugadores")
	}
	if p.EstaEnPartida(idInvitado) {
		return errors.New("no se puede invitar a alguien que ya está en la partida")
	}

	// Comprobar si los usuarios son amigos
	id1 := min(idCreador, idInvitado)
	id2 := max(idCreador, idInvitado)
	err := dao.bd.QueryRow(consultaAmistad, id1, id2).Scan(&id1)
	if err != nil && err == sql.ErrNoRows {
		return errors.New("no se puede invitar a una partida a alguien que no es amigo")
	} else if err != nil {
		return err
	}

	// Guardar en la base de datos la invitación
	_, err = dao.bd.Exec(crearInvitacion, idInvitado, p.IdPartida)
	return err
}

/*
	EntrarPartida comprueba en la base de datos si un usuario está invitado a una partida y en
	caso de que lo esté lo añade a la partida. Devuelve el estado de la partida o un error, en
	caso de que no se haya podido añadir, en formato json.
*/
func (dao *PartidaDAO) EntrarPartida(p *Partida, u Usuario, ws *websocket.Conn) mensajes.JsonData {
	var (
		numInvitaciones int
		res             mensajes.JsonData
	)

	// Comprobar si el usuario está invitado
	err := dao.bd.QueryRow(consultaInvitacion, u.Id, p.IdPartida).Scan(&numInvitaciones)
	if err != nil {
		return mensajes.ErrorJsonPartida(err.Error(), ErrorUnirsePartida)
	}
	if numInvitaciones == 0 {
		return mensajes.ErrorJsonPartida("No puedes unirte a una partida sin ser invitado",
			ErrorFaltaPermisoUnirse)
	}

	// Actualizar la partida con el nuevo usuario
	err = p.EntrarPartida(u, ws)
	if err != nil {
		return mensajes.ErrorJsonPartida(err.Error(), ErrorUnirsePartida)
	}

	// Codificar los datos de la partida en formato json
	mapstructure.Decode(p, &res)
	res["_tipoMensaje"] = "d"
	return res
}

/*
	AbandonarPartida elimina a un jugador de la partida, este no puede ser el creador de la misma.
	Devuelve el estado de la partida o un error, en caso de que no se haya podido borrar,
	en formato json.
*/
func (dao *PartidaDAO) AbandonarPartida(p *Partida, IdUsuario int) mensajes.JsonData {
	var res mensajes.JsonData

	// Esta función no se puede utilizar para borrar al creador
	if p.IdCreador == IdUsuario {
		return mensajes.ErrorJsonPartida("Mal uso de la funcion", ErrorUnirsePartida)
	}

	// Eliminar al jugador de la partida
	p.ExpulsarDePartida(IdUsuario)

	// Codificar los datos de la partida en formato json
	mapstructure.Decode(p, &res)
	res["_tipoMensaje"] = "d"
	return res
}

/*
	BorrarPartida borra una partida de la base de datos.
	Devuelve un error si no se ha podido borrar.
*/
func (dao *PartidaDAO) BorrarPartida(p *Partida) error {
	_, err := dao.bd.Exec(borrarPartida, p.IdPartida)
	return err
}
