package baseDatos

import (
	"PS_Risk_server/mensajes"
	"context"
	"database/sql"
	"errors"
	"strconv"
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
	consultaInvitaciones = "SELECT id_envia AS idEnvio, nombre FROM invitacionPartida " +
		"LEFT JOIN partida ON id_partida = id_envia WHERE id_recibe = $1"
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

type Usuario struct {
	Id, Icono, Aspecto, Riskos int
	Nombre, Correo, Clave      string
	RecibeCorreos              bool
}

func (u *Usuario) ToJSON() mensajes.JsonData {
	return mensajes.JsonData{
		"id":            u.Id,
		"nombre":        u.Nombre,
		"correo":        u.Correo,
		"clave":         u.Clave,
		"recibeCorreos": u.RecibeCorreos,
		"icono":         u.Icono,
		"aspecto":       u.Aspecto,
		"riskos":        u.Riskos,
	}
}

func (u *Usuario) Modificar(c string, v string) error {
	var err error
	switch c {
	case "Aspecto":
		u.Aspecto, err = strconv.Atoi(v)
		if err != nil {
			return err
		}
	case "Icono":
		u.Icono, err = strconv.Atoi(v)
		if err != nil {
			return err
		}
	case "Correo":
		u.Correo = v
	case "Clave":
		u.Clave = v
	case "Nombre":
		u.Nombre = v
	case "RecibeCorreos":
		u.RecibeCorreos, err = strconv.ParseBool(v)
		if err != nil {
			return err
		}
	default:
		return errors.New("el campo a modificar no existe")
	}
	return nil
}

type UsuarioDAO struct {
	bd *sql.DB
}

func NuevoUsuarioDAO(bd *sql.DB) UsuarioDAO {
	return UsuarioDAO{bd: bd}
}

func (dao *UsuarioDAO) CrearCuenta(nombre, correo, clave string,
	recibeCorreos bool) (Usuario, error) {

	var u Usuario
	// Iniciar una transaccion, solo se modifican las tablas si se modifican
	// todas
	ctx := context.Background()
	tx, err := dao.bd.BeginTx(ctx, nil)
	if err != nil {
		return u, err
	}
	id := 0
	err = tx.QueryRowContext(ctx, crearUsuario, nombre, correo, clave,
		recibeCorreos).Scan(&id)
	if err != nil {
		tx.Rollback()
		return u, err
	}
	_, err = tx.ExecContext(ctx, darIconosPorDefecto, id)
	if err != nil {
		tx.Rollback()
		return u, err
	}
	_, err = tx.ExecContext(ctx, darAspectosPorDefecto, id)
	if err != nil {
		tx.Rollback()
		return u, err
	}
	err = tx.Commit()
	if err != nil {
		return u, err
	}
	// Fin de la transaccion
	u = Usuario{
		Id: id, Icono: 0, Aspecto: 0, Riskos: 1000,
		Nombre: nombre, Correo: correo, Clave: clave,
		RecibeCorreos: recibeCorreos,
	}
	return u, nil
}

func (dao *UsuarioDAO) IniciarSesionNombre(nombre, clave string) (Usuario, error) {

	var id, icono, aspecto, riskos int
	var correo string
	var recibeCorreos bool
	var u Usuario

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

func (dao *UsuarioDAO) IniciarSesionCorreo(correo, clave string) (Usuario, error) {

	var id, icono, aspecto, riskos int
	var nombre string
	var recibeCorreos bool
	var u Usuario

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

func (dao *UsuarioDAO) ObtenerUsuario(id int, clave string) (Usuario, error) {

	var icono, aspecto, riskos int
	var nombre, correo string
	var recibeCorreos bool
	var u Usuario

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

func (dao *UsuarioDAO) ActualizarUsuario(u Usuario) mensajes.JsonData {
	var id int
	err := dao.bd.QueryRow(comprobarIconoComprado, u.Id, u.Icono).Scan(&id)
	if err != nil {
		return mensajes.ErrorJson("Icono no comprado", ErrorModificarUsuario)
	}
	err = dao.bd.QueryRow(comprobarAspectoComprado, u.Id, u.Aspecto).Scan(&id)
	if err != nil {
		return mensajes.ErrorJson("Aspecto no comprado", ErrorModificarUsuario)
	}
	res, err := dao.bd.Exec(actualizarUsuario, u.Aspecto, u.Icono, u.Nombre,
		u.Correo, u.Clave, u.RecibeCorreos, u.Id)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorModificarUsuario)
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return mensajes.ErrorJson("Error modificando usuario", ErrorModificarUsuario)
	}
	return mensajes.ErrorJson("", 0)
}

func (dao *UsuarioDAO) IncrementarRiskos(u *Usuario, r int) error {
	_, err := dao.bd.Exec(incrementarRiskos, r, u.Id)
	if err != nil {
		return err
	}
	u.Riskos += r
	return nil
}

func (dao *UsuarioDAO) ObtenerNotificaciones(u Usuario) mensajes.JsonData {
	var notificaciones []mensajes.JsonData
	n, err := dao.leerNotificaciones(u.Id, consultaSolicitudes, "Peticion de amistad")
	notificaciones = append(notificaciones, n...)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorNotificaciones)
	}
	n, err = dao.leerNotificaciones(u.Id, consultaInvitaciones, "Invitación")
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorNotificaciones)
	}
	notificaciones = append(notificaciones, n...)
	n, err = dao.leerNotificaciones(u.Id, consultaTurnos, "Notificación de turno")
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorNotificaciones)
	}
	notificaciones = append(notificaciones, n...)
	return mensajes.JsonData{"notificaciones": notificaciones}
}

func (dao *UsuarioDAO) leerNotificaciones(id int, consulta,
	tipo string) ([]mensajes.JsonData, error) {

	filas, err := dao.bd.Query(consulta, id)
	if err != nil {
		return nil, err
	}
	var notificaciones []mensajes.JsonData
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