package baseDatos

import (
	"PS_Risk_server/mensajes"
	"context"
	"database/sql"
	"strings"

	"github.com/lib/pq"
)

/*
	AmigosDAO permite modificar las tablas de amigos y solicitudes de amistad en
	la base de datos.
*/
type AmigosDAO struct {
	bd *sql.DB
}

/*
	NuevoAmigosDAO crea un DAO para gestionar relaciones de amistad entre usuarios.
*/
func NuevoAmigosDAO(bd *sql.DB) AmigosDAO {
	return AmigosDAO{bd: bd}
}

const (
	consultaAmigos = "SELECT id_usuario AS id, nombre, icono, aspecto FROM usuario INNER JOIN " +
		"(SELECT id_usuario2 AS idAmigo FROM esamigo WHERE id_usuario1 = $1 UNION " +
		"SELECT id_usuario1 AS idAmigo FROM esamigo WHERE id_usuario2 = $1) AS amigos ON id_usuario = idAmigo"
	eliminarAmistad      = "DELETE FROM esAmigo WHERE id_usuario1 = $1 AND id_usuario2 = $2"
	eliminarInvitaciones = "DELETE FROM invitacionPartida i USING partida p " +
		"WHERE i.id_envia = p.id_partida AND p.id_creador = $1 AND i.id_recibe = $2"
	eliminarSolicitudAmistad  = "DELETE FROM solicitudAmistad WHERE id_envia = $1 AND id_recibe = $2"
	consultarSolicitudAmistad = "SELECT COUNT(*) FROM solicitudAmistad WHERE id_envia = $1 AND id_recibe = $2"
	crearAmistad              = "INSERT INTO esAmigo (id_usuario1, id_usuario2) VALUES ($1, $2)"
	solicitarAmistad          = "INSERT INTO solicitudAmistad (id_envia, id_recibe) VALUES ($1, $2)"
	consultaAmistad           = "SELECT id_usuario1 FROM esAmigo WHERE id_usuario1 = $1 AND " +
		"id_usuario2 = $2"
	obtenerIdUsuario = "SELECT id_usuario FROM usuario WHERE nombre = $1"
)

/*
	ObtenerAmigos devuelve la lista de amigos en formato json del usuario indicado.
	Si ocurre algún error devuelve el error en formato json.
*/
func (dao *AmigosDAO) ObtenerAmigos(u Usuario) mensajes.JsonData {
	filas, err := dao.bd.Query(consultaAmigos, u.Id)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
	}
	var amigos []mensajes.JsonData
	for filas.Next() {
		var id, icono, aspecto int
		var nombre string
		if err := filas.Scan(&id, &nombre, &icono, &aspecto); err != nil {
			return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
		}
		amigos = append(amigos, mensajes.AmigoJson(id, icono, aspecto, nombre))
	}
	return mensajes.JsonData{"amigos": amigos}
}

/*
	EliminarAmigo borra de la base de datos la relación de amistad entre los
	usuarios indicados.
	Devuelve en formato json el error ocurrido o la ausencia de errores.
*/
func (dao *AmigosDAO) EliminarAmigo(u Usuario, id int) mensajes.JsonData {
	// El primer usuario en la tabla es siempre el de menor id
	id1 := min(u.Id, id)
	id2 := max(u.Id, id)

	// Iniciar transacción
	ctx := context.Background()
	tx, err := dao.bd.BeginTx(ctx, nil)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
	}

	// Eliminar de la base de datos que son amigos
	resultadoConsulta, err := tx.ExecContext(ctx, eliminarAmistad, id1, id2)
	if err != nil {
		tx.Rollback()
		return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
	}
	// Si no eran amigos se devuelve un error pero no pasa nada
	filasEliminadas, err := resultadoConsulta.RowsAffected()
	if err != nil {
		tx.Rollback()
		return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
	} else if filasEliminadas == 0 {
		tx.Rollback()
		return mensajes.ErrorJson("Los usuarios no eran amigos", mensajes.ErrorPeticion)
	}
	// Eliminar todas las invitaciones que tuvieran mutuamente
	_, err = tx.ExecContext(ctx, eliminarInvitaciones, id1, id2)
	if err != nil {
		tx.Rollback()
		return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
	}
	_, err = tx.ExecContext(ctx, eliminarInvitaciones, id2, id1)
	if err != nil {
		tx.Rollback()
		return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
	}

	// Finalizar la transacción
	err = tx.Commit()
	if err != nil {
		return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
	}
	return mensajes.ErrorJson("", mensajes.NoError)
}

/*
	AceptarSolicitudAmistad añade en la base de datos una relación de amistad entre
	los usuarios indicados y elimina la solicitud de amistad.
	Si el usuario que acepta también le había enviado una solicitud de amistad
	al otro usuario, esta segunda solicitud también se elimina.
	Devuelve en formato json el error ocurrido o la ausencia de errores.
*/
func (dao *AmigosDAO) AceptarSolicitudAmistad(u Usuario, id int) mensajes.JsonData {
	id1 := min(u.Id, id)
	id2 := max(u.Id, id)

	// Iniciar una transacción, solo se modifican las tablas si se modifican todas
	ctx := context.Background()
	tx, err := dao.bd.BeginTx(ctx, nil)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
	}

	// Eliminar la solicitud ya que se ha aceptado
	resultadoConsulta, err := tx.ExecContext(ctx, eliminarSolicitudAmistad, id, u.Id)
	if err != nil {
		tx.Rollback()
		return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
	}
	// Si no existía la solicitud no se pueden hacer amigos
	filasEliminadas, err := resultadoConsulta.RowsAffected()
	if err != nil {
		tx.Rollback()
		return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
	} else if filasEliminadas == 0 {
		tx.Rollback()
		return mensajes.ErrorJson("No existe la solicitud de amistad", mensajes.ErrorPeticion)
	}
	// Si el usuario que acepta ha enviado solicitud de amistad al otro, eliminarla.
	// No tiene por qué existir, si no está no se debe abortar el proceso.
	_, err = tx.ExecContext(ctx, eliminarSolicitudAmistad, u.Id, id)
	if err != nil {
		tx.Rollback()
		return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
	}
	// Guardar en la base de datos que son amigos
	_, err = tx.ExecContext(ctx, crearAmistad, id1, id2)
	if err != nil {
		tx.Rollback()
		return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
	}

	// Fin de la transacción
	err = tx.Commit()
	if err != nil {
		return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
	}
	return mensajes.ErrorJson("", mensajes.NoError)
}

/*
	RechazarSolicitudAmistad elimina la notificación de solicitud de amistad entre
	los usuarios indicados sin añadirlos como amigos en la base de datos.
	Devuelve en formato json el error ocurrido o la ausencia de errores.
*/
func (dao *AmigosDAO) RechazarSolicitudAmistad(u Usuario, id int) mensajes.JsonData {
	// Eliminar de la base de datos la solicitud
	resultadoConsulta, err := dao.bd.Exec(eliminarSolicitudAmistad, id, u.Id)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
	}
	// Si no había solicitud se devuelve un error pero no pasa nada
	filasEliminadas, err := resultadoConsulta.RowsAffected()
	if err != nil {
		return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
	} else if filasEliminadas == 0 {
		return mensajes.ErrorJson("No existe la solicitud de amistad", mensajes.ErrorPeticion)
	}
	return mensajes.ErrorJson("", mensajes.NoError)
}

/*
	EnviarSolicitudAmistad guarda en la base de datos una solicitud de amistad
	enviada por u al usuario de nombre amigo.
	Devuelve en formato json el error ocurrido o la ausencia de errores.
*/
func (dao *AmigosDAO) EnviarSolicitudAmistad(u Usuario, amigo string) mensajes.JsonData {
	var idAmigo int
	var numSolicitudes int

	// Comprobar si el usuario al que se le quiere enviar la solicitud existe
	err := dao.bd.QueryRow(obtenerIdUsuario, amigo).Scan(&idAmigo)
	if err == sql.ErrNoRows {
		return mensajes.ErrorJson("No existe ningún usuario con el nombre indicado",
			mensajes.ErrorPeticion)
	}
	if err != nil {
		return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
	}

	id1 := min(u.Id, idAmigo)
	id2 := max(u.Id, idAmigo)

	// Comprobar si los usuarios no son amigos ya
	err = dao.bd.QueryRow(consultaAmistad, id1, id2).Scan(&id1)
	if err == nil {
		return mensajes.ErrorJson("Los usuarios ya son amigos", mensajes.ErrorPeticion)
	}
	if err != sql.ErrNoRows {
		return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
	}

	// Comprobar si existe una solicitud de amistad del otro usuario
	err = dao.bd.QueryRow(consultarSolicitudAmistad, idAmigo, u.Id).Scan(&numSolicitudes)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
	}
	if numSolicitudes == 0 {
		// No había solicitud previa, enviar la nueva
		// Guardar en la base de datos que se ha enviado la solicitud
		_, err = dao.bd.Exec(solicitarAmistad, u.Id, idAmigo)
		if err != nil {
			if e, ok := err.(*pq.Error); ok {
				if e.Code.Name() == violacionUnicidad {
					if strings.Contains(e.Error(), "solicitudamistad_pkey") {
						return mensajes.ErrorJson("Ya has enviado una solicitud de "+
							"amistad a este usuario", mensajes.ErrorPeticion)
					}
				} else if e.Code.Name() == violacionRestricciones {
					if strings.Contains(e.Error(), "solicitudamistad_check") {
						return mensajes.ErrorJson("No puedes enviarte una solicitud"+
							" de amistad a ti mismo", mensajes.ErrorPeticion)
					}
				}
			}
			return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
		}
	} else {
		// Había solicitud previa, aceptarla
		return dao.AceptarSolicitudAmistad(u, idAmigo)
	}

	return mensajes.ErrorJson("", mensajes.NoError)
}
