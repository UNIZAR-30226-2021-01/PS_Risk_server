package baseDatos

import (
	"PS_Risk_server/mensajes"
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/lib/pq"
)

const (
	crearUsuario = "INSERT INTO usuario (aspecto, icono, nombre, correo, clave, " +
		"riskos, recibeCorreos) VALUES (0, 0, $1, $2, $3, 1000, $4) " +
		"RETURNING id_usuario"
	darIconosPorDefecto = "INSERT INTO iconosComprados (id_usuario, id_icono) " +
		"VALUES ($1, 0)"
	darAspectosPorDefecto = "INSERT INTO aspectosComprados (id_usuario, " +
		"id_aspecto) VALUES ($1, 0)"

	consultaUsuarioNombre = "SELECT id_usuario AS id, icono, aspecto, riskos, " +
		"correo, recibeCorreos FROM usuario WHERE nombre = $1 AND clave = $2"
	consultaUsuarioCorreo = "SELECT id_usuario AS id, icono, aspecto, riskos, " +
		"nombre, recibeCorreos FROM usuario WHERE correo = $1 AND clave = $2"
	consultaUsuarioId = "SELECT icono, aspecto, riskos, correo, nombre, clave," +
		" recibeCorreos FROM usuario WHERE id_usuario = $1"
	consultaUsuario = "SELECT icono, aspecto, riskos, correo, nombre," +
		" recibeCorreos FROM usuario WHERE id_usuario = $1 AND clave = $2"

	actualizarUsuario = "UPDATE usuario SET aspecto = $1, icono = $2, nombre = $3, " +
		"correo = $4, clave = $5, recibeCorreos = $6 WHERE id_usuario = $7"
	incrementarRiskos = "UPDATE usuario SET riskos = riskos + $1" +
		"WHERE id_usuario = $2"

	consultaSolicitudes = "SELECT id_envia AS idEnvio, nombre FROM solicitudAmistad " +
		"LEFT JOIN usuario ON id_usuario = id_envia WHERE id_recibe = $1"
	consultaInvitaciones = "SELECT id_envia AS idEnvio, usuario.nombre FROM " +
		"invitacionPartida JOIN partida ON id_partida = id_envia JOIN usuario " +
		"ON id_creador = id_usuario WHERE id_recibe = $1"
	consultaTurnos = "SELECT id_envia AS idEnvio, nombre FROM notificacionTurno " +
		"LEFT JOIN partida ON id_partida = id_envia WHERE id_recibe = $1"

	consultaAspectosUsuario = "SELECT aspecto.id_aspecto AS id, aspecto.precio AS precio " +
		"FROM aspecto INNER JOIN aspectoscomprados ON aspecto.id_aspecto = aspectoscomprados.id_aspecto " +
		"WHERE aspectoscomprados.id_usuario = "
	consultaIconosUsuario = "SELECT icono.id_icono AS id, icono.precio AS precio " +
		"FROM icono INNER JOIN iconoscomprados ON icono.id_icono = iconoscomprados.id_icono " +
		"WHERE iconoscomprados.id_usuario = "

	comprobarIconoComprado = "SELECT id_usuario AS id FROM iconosComprados " +
		"WHERE id_usuario = $1 AND id_icono = $2"
	comprobarAspectoComprado = "SELECT id_usuario AS id FROM aspectosComprados " +
		"WHERE id_usuario = $1 AND id_aspecto = $2"
)

/*
	UsuarioDAO permite modificar y leer las tablas de usuario y relacionadas.
*/
type UsuarioDAO struct {
	bd *sql.DB
}

/*
	NuevoUsuarioDAO crea un UsuarioDAO.
*/
func NuevoUsuarioDAO(bd *sql.DB) UsuarioDAO {
	return UsuarioDAO{bd: bd}
}

/*
	CrearCuenta crea una cuenta de usuario en la base de datos y devuelve el
	usuario creado.
	Devuelve error en caso de no poder crearla.
*/
func (dao *UsuarioDAO) CrearCuenta(nombre, correo, clave string,
	recibeCorreos bool) (Usuario, error) {

	var (
		u  Usuario
		id int
	)

	// Iniciar una transacción, solo se modifican las tablas si se modifican
	// todas
	ctx := context.Background()
	tx, err := dao.bd.BeginTx(ctx, nil)
	if err != nil {
		return u, err
	}

	// Crear el usuario en la base de datos
	err = tx.QueryRowContext(ctx, crearUsuario, nombre, correo, clave,
		recibeCorreos).Scan(&id)
	if err != nil {
		tx.Rollback()
		e := err.(*pq.Error)
		if e.Code.Name() == violacionUnicidad {
			if strings.Contains(e.Error(), "usuario_correo_key") {
				return u, errors.New("ya existe un usuario con la dirección de correo indicada")
			} else if strings.Contains(e.Error(), "usuario_nombre_key") {
				return u, errors.New("ya existe un usuario con el nombre indicado")
			}
		}
		return u, err
	}
	// Guardar en la base de datos los iconos por defecto como comprados
	_, err = tx.ExecContext(ctx, darIconosPorDefecto, id)
	if err != nil {
		tx.Rollback()
		return u, err
	}
	// Guardar en la base de datos los aspectos por defecto como comprados
	_, err = tx.ExecContext(ctx, darAspectosPorDefecto, id)
	if err != nil {
		tx.Rollback()
		return u, err
	}

	// Fin de la transacción
	err = tx.Commit()
	if err != nil {
		return u, err
	}

	u = Usuario{
		Id: id, Icono: 0, Aspecto: 0, Riskos: 1000,
		Nombre: nombre, Correo: correo, Clave: clave,
		RecibeCorreos: recibeCorreos,
	}
	return u, nil
}

/*
	IniciarSesionNombre devuelve los datos de un usuario que use el nombre y clave
	pasados como parámetros. Si no existe devuelve error.
*/
func (dao *UsuarioDAO) IniciarSesionNombre(nombre, clave string) (Usuario, error) {

	var id, icono, aspecto, riskos int
	var correo string
	var recibeCorreos bool
	var u Usuario

	// Obtener los datos de usuario de la base de datos
	err := dao.bd.QueryRow(consultaUsuarioNombre, nombre, clave).Scan(&id,
		&icono, &aspecto, &riskos, &correo, &recibeCorreos)
	if err != nil {
		return u, err
	}
	u = Usuario{
		Id: id, Icono: icono, Aspecto: aspecto, Riskos: riskos,
		Nombre: nombre, Correo: correo, Clave: clave,
		RecibeCorreos: recibeCorreos,
	}
	return u, nil
}

/*
	IniciarSesionNombre devuelve los datos de un usuario que use el correo y clave
	pasados como parámetros. Si no existe devuelve error.
*/
func (dao *UsuarioDAO) IniciarSesionCorreo(correo, clave string) (Usuario, error) {

	var id, icono, aspecto, riskos int
	var nombre string
	var recibeCorreos bool
	var u Usuario

	// Obtener los datos de usuario de la base de datos
	err := dao.bd.QueryRow(consultaUsuarioCorreo, correo, clave).Scan(&id,
		&icono, &aspecto, &riskos, &nombre, &recibeCorreos)
	if err != nil {
		return u, err
	}
	u = Usuario{
		Id: id, Icono: icono, Aspecto: aspecto, Riskos: riskos,
		Nombre: nombre, Correo: correo, Clave: clave,
		RecibeCorreos: recibeCorreos,
	}
	return u, nil
}

/*
	ObtenerUsuario devuelve los datos de un usuario que tenga el id y la clave
	pasados como parámetros en la base de datos. Si no existe devuelve error.
*/
func (dao *UsuarioDAO) ObtenerUsuario(id int, clave string) (Usuario, error) {

	var icono, aspecto, riskos int
	var nombre, correo string
	var recibeCorreos bool
	var u Usuario

	// Obtener los datos de usuario de la base de datos
	err := dao.bd.QueryRow(consultaUsuario, id, clave).Scan(&icono,
		&aspecto, &riskos, &correo, &nombre, &recibeCorreos)
	if err != nil {
		return u, err
	}
	u = Usuario{
		Id: id, Icono: icono, Aspecto: aspecto, Riskos: riskos,
		Nombre: nombre, Correo: correo, Clave: clave,
		RecibeCorreos: recibeCorreos,
	}
	return u, nil
}

/*
	ObtenerUsuarioId devuelve los datos de un usuario de la base de datos.
	Si no existe devuelve error.
*/
func (dao *UsuarioDAO) ObtenerUsuarioId(id int) (Usuario, error) {
	var icono, aspecto, riskos int
	var nombre, correo, clave string
	var recibeCorreos bool
	var u Usuario

	err := dao.bd.QueryRow(consultaUsuarioId, id).Scan(&icono,
		&aspecto, &riskos, &correo, &nombre, &clave, &recibeCorreos)
	if err != nil {
		return u, err
	}
	u = Usuario{
		Id: id, Icono: icono, Aspecto: aspecto, Riskos: riskos,
		Nombre: nombre, Correo: correo, Clave: clave,
		RecibeCorreos: recibeCorreos,
	}
	return u, nil
}

/*
	ActualizarUsuario modifica en la base de datos un usuario.
	Si se modifica correctamente, devuelve error con código NoError en formato
	json. En caso contrario devuelve el error ocurrido en el mismo formato.
*/
func (dao *UsuarioDAO) ActualizarUsuario(u Usuario) mensajes.JsonData {
	var id int

	// Comprobar que el icono lo tenga comprado
	err := dao.bd.QueryRow(comprobarIconoComprado, u.Id, u.Icono).Scan(&id)
	if err != nil {
		return mensajes.ErrorJson("Icono no comprado", mensajes.ErrorPeticion)
	}

	// Comprobar que el aspecto lo tenga comprado
	err = dao.bd.QueryRow(comprobarAspectoComprado, u.Id, u.Aspecto).Scan(&id)
	if err != nil {
		return mensajes.ErrorJson("Aspecto no comprado", mensajes.ErrorPeticion)
	}

	// Actualizar el usuario en la base de datos
	res, err := dao.bd.Exec(actualizarUsuario, u.Aspecto, u.Icono, u.Nombre,
		u.Correo, u.Clave, u.RecibeCorreos, u.Id)
	if err != nil {
		e := err.(*pq.Error)
		if e.Code.Name() == violacionUnicidad {
			if strings.Contains(e.Error(), "usuario_correo_key") {
				err = errors.New("ya existe un usuario con la dirección de correo indicada")
			} else if strings.Contains(e.Error(), "usuario_nombre_key") {
				err = errors.New("ya existe un usuario con el nombre indicado")
			}
		}
		return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return mensajes.ErrorJson("Error modificando usuario", mensajes.ErrorPeticion)
	}
	return mensajes.ErrorJson("", mensajes.NoError)
}

/*
	IncrementarRiskos de un usuario en r.
	Devuelve error en caso de no poder hacerlo.
*/
func (dao *UsuarioDAO) IncrementarRiskos(u *Usuario, r int) error {
	_, err := dao.bd.Exec(incrementarRiskos, r, u.Id)
	if err != nil {
		return err
	}
	u.Riskos += r
	return nil
}

/*
	ObtenerNotificaciones devuelve las notificaciones en formato json de un usuario.
	Si no puede obtener las notificaciones devuelve error en formato json.
*/
func (dao *UsuarioDAO) ObtenerNotificaciones(u Usuario) mensajes.JsonData {
	var notificaciones []mensajes.JsonData

	// Obtener las solicitudes de amistad
	n, err := dao.leerNotificaciones(u.Id, consultaSolicitudes, "Peticion de amistad")
	notificaciones = append(notificaciones, n...)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
	}

	// Obtener las invitaciones a partidas
	n, err = dao.leerNotificaciones(u.Id, consultaInvitaciones, "Invitacion")
	if err != nil {
		return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
	}
	notificaciones = append(notificaciones, n...)

	// Obtener las notificaciones de turnos
	n, err = dao.leerNotificaciones(u.Id, consultaTurnos, "Notificacion de turno")
	if err != nil {
		return mensajes.ErrorJson(err.Error(), mensajes.ErrorPeticion)
	}
	notificaciones = append(notificaciones, n...)

	// Devolver resultado
	return mensajes.JsonData{"notificaciones": notificaciones}
}

/*
	leerNotificaciones devuelve un array de un tipo de notificaciones de la base de datos
	en formato json. Devuelve un error en caso de no poder obtenerlas.
*/
func (dao *UsuarioDAO) leerNotificaciones(id int, consulta,
	tipo string) ([]mensajes.JsonData, error) {

	var notificaciones []mensajes.JsonData

	// Obtener las notificaciones
	filas, err := dao.bd.Query(consulta, id)
	if err != nil {
		return nil, err
	}
	for filas.Next() {
		var idEnvia int
		var nombre string
		if err := filas.Scan(&idEnvia, &nombre); err != nil {
			return nil, err
		}
		notificaciones = append(notificaciones,
			mensajes.NotificacionJson(idEnvia, tipo, nombre))
	}
	return notificaciones, nil
}
