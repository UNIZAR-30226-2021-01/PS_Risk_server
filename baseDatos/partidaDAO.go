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
	obtenerPartidasEmpezadas = "SELECT json_estado AS p FROM partida WHERE empezada = true"
	obtenerSalas             = "SELECT id_partida FROM partida WHERE empezada = false"
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

	if tiempoTurno < 1 {
		return nil, errors.New("no se puede poner un tiempo de turno menor de 1 minuto")
	}

	// Crea la partida en la base de datos
	err := dao.bd.QueryRow(crearPartida, creador.Id, nombreSala,
		[]byte(`{}`)).Scan(&idPartida)
	if err != nil {
		e := err.(*pq.Error)
		if e.Code.Name() == cadenaDemasiadoLarga {
			// Hay un atributo de texto demasiado largo, pero el error no da informaci??n
			// suficiente para saber cu??l es. El ??nico que deber??a poder ocurrir
			// es el de nombre de usuario con m??s de 20 caracteres
			err = errors.New("el nombre de la partida es demasiado largo, solo se " +
				"admiten nombres de 20 caracteres o menos")
		}
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

	// Inicia una transacci??n en la base de datos
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
	// Guardar a qu?? jugador le corresponde turno
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
	// Guardar en la base de datos qu?? jugadores juegan en la partida
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
	partida. Devuelve error en caso de que no se pueda completar la invitaci??n.
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
		return errors.New("no se puede invitar a alguien que ya est?? en la partida")
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

	// Guardar en la base de datos la invitaci??n
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
	EntrarPartida a??ade un usuario a una partida y borra la invitaci??n.
	Devuelve el estado de la partida o un error, en caso de que no se haya
	podido a??adir o no existiera la invitaci??n, en formato json.
*/
func (dao *PartidaDAO) EntrarPartida(p *Partida, u Usuario, ws *websocket.Conn) mensajes.JsonData {
	var res mensajes.JsonData

	// Consumir la invitaci??n
	resultado, err := dao.bd.Exec(borrarInvitacion, u.Id, p.IdPartida)
	if err != nil {
		return mensajes.ErrorJsonPartida(err.Error(), mensajes.ErrorPeticion)
	}
	// Comprobar que hab??a invitaci??n
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
	RechazarPartida borra una invitaci??n a una partida para un usuario sin
	a??adirlo a ella.
	Devuelve error si no se ha podido eliminar la invitaci??n.
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
	AbandonarPartida elimina a un jugador de la partida, este no puede ser el
	creador de la misma.
	Devuelve el estado de la partida o un error, en caso de que no se haya
	podido borrar, en formato json.
*/
func (dao *PartidaDAO) AbandonarPartida(p *Partida, IdUsuario int) mensajes.JsonData {
	var res mensajes.JsonData

	// Esta funci??n no se puede utilizar para borrar al creador
	if p.IdCreador == IdUsuario {
		return mensajes.ErrorJsonPartida("Mal uso de la funci??n", mensajes.ErrorPeticion)
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

/*
	NotificarTurno guarda en la base de datos de qui??n es el turno actual en la
	partida, y borra la notificaci??n del turno anterior si exist??a.
*/
func (dao *PartidaDAO) NotificarTurno(p *Partida) {
	// Borrar el turno anterior de la base de datos
	dao.bd.Exec(borrarTurno, p.IdPartida)

	// Guardar el turno actual en la base de datos
	dao.bd.Exec(guardarTurno, p.Jugadores[p.TurnoJugador].Id, p.IdPartida)
}

/*
	BorrarNotificacionTurno elimina de la base de datos la notificaci??n que indica
	que es el turno del jugador indicado en la partida indicada.
	Devuelve el error ocurrido, o nil si se ha podido eliminar correctamente.
*/
func (dao *PartidaDAO) BorrarNotificacionTurno(idPartida, idUsuario int) error {
	resultado, err := dao.bd.Exec(borrarTurnoJugador, idPartida, idUsuario)
	if err != nil {
		return errors.New("no se ha podido eliminar la notificaci??n")
	}
	borradas, _ := resultado.RowsAffected()
	if borradas == 0 {
		return errors.New("no se ha podido eliminar la notificaci??n")
	}
	return nil
}

/*
	ActualizarPartida guarda en la base de datos el estado en que se encuentra
	la partida actualmente.
*/
func (dao *PartidaDAO) ActualizarPartida(p *Partida) {
	estadoJson, _ := json.Marshal(p)
	dao.bd.Exec(actualizarEstadoPartida, estadoJson, p.IdPartida)
}

/*
	EliminarSalas borra de la base de datos todas las partidas que no est??n
	empezadas.
*/
func (dao *PartidaDAO) EliminarSalas() error {
	_, err := dao.bd.Exec(eliminarSalas)
	return err
}

/*
	ObtenerPartidasEmpezadas devuelve un slice con el estado en formato json
	(slice de bytes) de cada partida empezada que hay en la base de datos.
	Devuelve error si hay alg??n problema.
*/
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

/*
	ObtenerSalas devuelve una lista con los identificadores de todas las partidas
	sin empezar que hay guardadas en la base de datos.
	Devuelve error si hay alg??n problema.
*/
func (dao *PartidaDAO) ObtenerSalas() ([]int, error) {
	var resultado []int

	filas, err := dao.bd.Query(obtenerSalas)
	if err != nil {
		return nil, err
	}
	for filas.Next() {
		var id int
		if err := filas.Scan(&id); err != nil {
			return nil, err
		}
		resultado = append(resultado, id)
	}
	return resultado, nil
}
