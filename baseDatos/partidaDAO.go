package baseDatos

import (
	"PS_Risk_server/mensajes"
	"PS_Risk_server/mensajesInternos"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/lib/pq"
	"github.com/mitchellh/mapstructure"
)

const (
	crearPartida = "INSERT INTO partida (id_creador, nombre, empezada, json_estado) " +
		"VALUES ($1, $2, false, $3) RETURNING id_partida"
	borrarInvitaciones      = "DELETE FROM invitacionPartida WHERE id_envia = $1"
	actualizarEstadoPartida = "UPDATE partida SET json_estado = $1 WHERE " +
		"id_partida = $2"
	crearInvitacion = "INSERT INTO invitacionPartida (id_recibe, id_envia) " +
		"VALUES ($1, $2)"
	borrarInvitacion = "DELETE FROM invitacionPartida WHERE id_recibe = $1 " +
		"AND id_envia = $2"
	consultaInvitacion = "SELECT COUNT(*) FROM invitacionPartida WHERE " +
		"id_recibe = $1 AND id_envia = $2"
	borrarPartida    = "DELETE FROM partida WHERE id_partida = $1"
	guardarJugadores = "INSERT INTO juega (id_partida, id_usuario) VALUES ($1, $2)"
	obtenerPartida   = "SELECT id_creador, json_estado FROM partida " +
		"WHERE id_partida = $1"
	consultaPartidas = "SELECT id_partida FROM juega WHERE id_usuario = $1"
	guardarTurno     = "INSERT INTO notificacionTurno (id_recibe, id_envia) " +
		"VALUES ($1, $2)"
	borrarTurno        = "DELETE FROM notificacionTurno WHERE id_envia = $1"
	borrarTurnoJugador = "DELETE FROM notificacionTurno WHERE id_envia = $1 " +
		"AND id_recibe = $2"
	empezarPartida           = "UPDATE partida SET empezada = true WHERE id_partida = $1"
	eliminarSalas            = "DELETE FROM partida WHERE empezada = false"
	obtenerPartidasEmpezadas = "SELECT json_estado AS p FROM partida"
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

	if nombreSala == "" {
		return nil, errors.New("no se puede crear una sala sin nombre")
	}

	if tiempoTurno < 10 {
		return nil, errors.New("no se puede poner un tiempo de turno menor de 10 minutos")
	}

	// Crea la partida en la base de datos
	err := dao.bd.QueryRow(crearPartida, creador.Id, nombreSala, []byte(`{}`)).Scan(&idPartida)
	if err != nil {
		return nil, err
	}

	// Crea la estructura de datos
	p := &Partida{
		IdPartida:           idPartida,
		IdCreador:           creador.Id,
		TiempoTurno:         tiempoTurno,
		TurnoActual:         0,
		TurnoJugador:        0,
		Fase:                0,
		Nombre:              nombreSala,
		Empezada:            false,
		Territorios:         []Territorio{},
		Jugadores:           []Jugador{},
		Conexiones:          sync.Map{},
		Mensajes:            make(chan mensajesInternos.MensajePartida, maxMensajes),
		UltimoTurno:         "",
		MovimientoRealizado: false,
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
		return mensajes.ErrorJsonPartida(err.Error(), mensajes.ErrorPeticion)
	}

	// Inicia una transacción en la base de datos
	ctx := context.Background()
	tx, err := dao.bd.BeginTx(ctx, nil)
	if err != nil {
		p.AnularInicio()
		return mensajes.ErrorJsonPartida(err.Error(), mensajes.ErrorPeticion)
	}
	// Borrar todas las invitaciones a la partida
	_, err = tx.ExecContext(ctx, borrarInvitaciones, p.IdPartida)
	if err != nil {
		tx.Rollback()
		p.AnularInicio()
		return mensajes.ErrorJsonPartida(err.Error(), mensajes.ErrorPeticion)
	}
	// Guardar a qué jugador le corresponde turno
	_, err = tx.ExecContext(ctx, guardarTurno, p.IdCreador, p.IdPartida)
	if err != nil {
		tx.Rollback()
		p.AnularInicio()
		return mensajes.ErrorJsonPartida(err.Error(), mensajes.ErrorPeticion)
	}
	// Actualiza el estado de la partida en la base de datos
	estadoJson, _ := json.Marshal(p)
	_, err = tx.ExecContext(ctx, actualizarEstadoPartida, estadoJson, p.IdPartida)
	if err != nil {
		tx.Rollback()
		p.AnularInicio()
		return mensajes.ErrorJsonPartida(err.Error(), mensajes.ErrorPeticion)
	}
	// Marcar la partida como iniciada
	_, err = tx.ExecContext(ctx, empezarPartida, p.IdPartida)
	if err != nil {
		tx.Rollback()
		p.AnularInicio()
		return mensajes.ErrorJsonPartida(err.Error(), mensajes.ErrorPeticion)
	}
	// Guardar en la base de datos qué jugadores juegan en la partida
	for _, j := range p.Jugadores {
		_, err = tx.ExecContext(ctx, guardarJugadores, p.IdPartida, j.Id)
		if err != nil {
			tx.Rollback()
			p.AnularInicio()
			return mensajes.ErrorJsonPartida(err.Error(), mensajes.ErrorPeticion)
		}
	}
	// Codificar los datos de la partida en formato json
	mapstructure.Decode(p, &res)
	// Finalizar la transaccion
	err = tx.Commit()
	if err != nil {
		p.AnularInicio()
		return mensajes.ErrorJsonPartida(err.Error(), mensajes.ErrorPeticion)
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
	if err == sql.ErrNoRows {
		return errors.New("no se puede invitar a una partida a alguien que no es amigo")
	}
	if err != nil {
		return err
	}

	// Guardar en la base de datos la invitación
	_, err = dao.bd.Exec(crearInvitacion, idInvitado, p.IdPartida)
	if e, ok := err.(*pq.Error); ok {
		if e.Code.Name() == violacionUnicidad {
			if strings.Contains(e.Error(), "invitacionpartida_pkey") {
				return errors.New("ya has invitado a este usuario")
			}
		}
	}
	return err
}

/*
	EntrarPartida añade un usuario a una partida y borra la invitación.
	Devuelve el estado de la partida o un error, en caso de que no se haya
	podido añadir o no existiera la invitación, en formato json.
*/
func (dao *PartidaDAO) EntrarPartida(p *Partida, u Usuario, ws *websocket.Conn) mensajes.JsonData {
	var res mensajes.JsonData

	// Consumir la invitación
	resultado, err := dao.bd.Exec(borrarInvitacion, u.Id, p.IdPartida)
	if err != nil {
		return mensajes.ErrorJsonPartida(err.Error(), mensajes.ErrorPeticion)
	}
	// Comprobar que había invitación
	numInvitaciones, err := resultado.RowsAffected()
	if err != nil {
		return mensajes.ErrorJsonPartida(err.Error(), mensajes.ErrorPeticion)
	}
	if numInvitaciones == 0 {
		return mensajes.ErrorJsonPartida("No puedes unirte a una partida sin ser invitado",
			mensajes.ErrorPeticion)
	}

	// Actualizar la partida con el nuevo usuario
	err = p.EntrarPartida(u, ws)
	if err != nil {
		return mensajes.ErrorJsonPartida(err.Error(), mensajes.ErrorPeticion)
	}

	// Codificar los datos de la partida en formato json
	mapstructure.Decode(p, &res)
	res["_tipoMensaje"] = "d"
	delete(res, "turnoJugador")
	return res
}

/*
	RechazarPartida borra una invitación a una partida para un usuario sin
	añadirlo a ella.
	Devuelve error si no se ha podido eliminar la invitación.
*/
func (dao *PartidaDAO) RechazarPartida(idPartida int, u Usuario) error {
	resultado, err := dao.bd.Exec(borrarInvitacion, u.Id, idPartida)
	if err != nil {
		return err
	}
	borradas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if borradas == 0 {
		return errors.New("no se puede rechazar una partida a la que no te " +
			"han invitado")
	}
	return nil
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
		return mensajes.ErrorJsonPartida("Mal uso de la función", mensajes.ErrorPeticion)
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

/*
	ObtenerPartidas obtiene los identificadores de las partidas que juega un usuario.
	Devuelve un error si no se han podido obtener.
*/
func (dao *PartidaDAO) ObtenerPartidas(u Usuario) ([]int, error) {
	var resultado []int

	// Consulta para obtener los identificadores de las partidas
	filas, err := dao.bd.Query(consultaPartidas, u.Id)
	if err != nil {
		return nil, err
	}
	for filas.Next() {
		var idPartida int
		if err := filas.Scan(&idPartida); err != nil {
			return nil, err
		}
		resultado = append(resultado, idPartida)
	}
	if filas.Err() != nil {
		return nil, filas.Err()
	}

	return resultado, nil
}

func (dao *PartidaDAO) NotificarTurno(p *Partida) error {
	// Inicia una transacción en la base de datos
	ctx := context.Background()
	tx, err := dao.bd.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Borrar el turno anterior de la base de datos
	_, err = tx.ExecContext(ctx, borrarTurno, p.IdPartida)
	if err != nil {
		tx.Rollback()
		return err
	}
	// Guardar el turno actual en la base de datos
	_, err = tx.ExecContext(ctx, guardarTurno, p.Jugadores[p.TurnoJugador].Id, p.IdPartida)
	if err != nil {
		tx.Rollback()
		return err
	}
	// Finalizar la transaccion
	err = tx.Commit()
	return err
}

/*
	BorrarNotificacionTurno elimina de la base de datos la notificación que indica
	que es el turno del jugador indicado en la partida indicada.
	Devuelve el error ocurrido, o nil si se ha podido eliminar correctamente.
*/
func (dao *PartidaDAO) BorrarNotificacionTurno(idPartida int, idUsuario int) error {
	resultado, err := dao.bd.Exec(borrarTurnoJugador, idPartida, idUsuario)
	if err != nil {
		return err
	}
	borradas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if borradas == 0 {
		return errors.New("no se ha podido eliminar la notificación")
	}
	return nil
}

func (dao *PartidaDAO) ActualizarPartida(p *Partida) {
	estadoJson, _ := json.Marshal(p)
	dao.bd.Exec(actualizarEstadoPartida, estadoJson, p.IdPartida)
}

func (dao *PartidaDAO) EliminarSalas() error {
	_, err := dao.bd.Exec(eliminarSalas)
	return err
}

func (dao *PartidaDAO) ObtenerPartidasEmpezadas() ([][]byte, error) {
	var resultado [][]byte

	filas, err := dao.bd.Query(obtenerPartidasEmpezadas)
	if err != nil {
		return nil, err
	}
	for filas.Next() {
		var datos []byte
		if err := filas.Scan(&datos); err != nil {
			return nil, err
		}
		resultado = append(resultado, datos)
	}
	return resultado, nil
}
