package baseDatos

import (
	"PS_Risk_server/mensajes"
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

type BD struct {
	bbdd string
	bd   *sql.DB
}

// Códigos de error posibles
const (
	NoError                 = 0
	ErrorIniciarTransaccion = iota
	ErrorCrearCuenta        = iota
	ErrorLeerAspectosTienda = iota
	ErrorLeerIconosTienda   = iota
	ErrorTipoIncorrecto     = iota
	ErrorBusquedaUsuario    = iota
	ErrorModificarUsuario   = iota
	ErrorIniciarSesion      = iota
	ErrorCampoIncorrecto    = iota
)

// Consultas SQL
const (
	crearUsuario = "INSERT INTO usuario (aspecto, icono, nombre, correo, clave," +
		" riskos, recibeCorreos) VALUES (1, 1, $1, $2, $3, 0, $4) " +
		"RETURNING id_usuario"
	darIconosPorDefecto = "INSERT INTO iconosComprados (id_usuario, id_icono)" +
		" VALUES ($1, 1)"
	darAspectosPorDefecto = "INSERT INTO aspectosComprados (id_usuario, " +
		"id_aspecto) VALUES ($1, 1)"
	consultaAspectos = "SELECT id_aspecto AS id, precio FROM aspecto"
	consultaIconos   = "SELECT id_icono AS id, precio FROM icono"
	consultaUsuario  = "SELECT aspecto, correo, icono, nombre, recibeCorreos, " +
		"riskos FROM usuario WHERE id_usuario = $1 AND clave = $2"
	consultaUsuarioNombre = "SELECT aspecto, correo, icono, id_usuario, recibeCorreos, " +
		"riskos FROM usuario WHERE nombre = $1 AND clave = $2"
	consultaUsuarioCorreo = "SELECT aspecto, nombre, icono, id_usuario, recibeCorreos, " +
		"riskos FROM usuario WHERE correo = $1 AND clave = $2"
	consultaAspectosUsuario = "SELECT aspecto.id_aspecto AS id, aspecto.precio AS precio " +
		"FROM aspecto INNER JOIN aspectoscomprados ON aspecto.id_aspecto = aspectoscomprados.id_aspecto " +
		"WHERE aspectoscomprados.id_usuario = "
	consultaIconosUsuario = "SELECT icono.id_icono AS id, icono.precio AS precio " +
		"FROM icono INNER JOIN iconoscomprados ON icono.id_icono = iconoscomprados.id_icono " +
		"WHERE iconoscomprados.id_usuario = "
	actualizarUsuarioInicio = "UPDATE usuario SET "
	actualizarUsuarioFin    = " = $1 WHERE id_usuario = $2 AND clave = $3"
	consultaAmigos          = "SELECT id_usuario as id, nombre, icono, aspecto FROM usuario, " +
		"(SELECT id_usuario1 as idAmigo1 FROM esamigo WHERE id_usuario2 = $1) as amigos1, " +
		"(SELECT id_usuario2 as idAmigo2 FROM esamigo WHERE id_usuario1 = $1) as amigos2 " +
		"WHERE id_usuario = idAmigo1 OR id_usuario = idAmigo2"
	consultaSolicitudes = "SELECT id_envia AS idEnvio, nombre FROM solicitudAmistad LEFT JOIN usuario ON " +
		"id_usuario = id_envia WHERE id_recibe = $1"
	consultaInvitaciones = "SELECT id_envia AS idEnvio, nombre FROM invitacionPartida LEFT JOIN partida ON " +
		"id_partida = id_envia WHERE id_recibe = $1"
	consultaTurnos = "SELECT id_envia AS idEnvio, nombre FROM notificacionTurno LEFT JOIN partida ON " +
		"id_partida = id_envia WHERE id_recibe = $1"
	solicitarAmistad = "INSERT INTO solicitudAmistad (id_envia, id_recibe) VALUES ($1, $2)"
)

// NuevaBD crea una nueva conexion a la base de datos bbdd y la formatea
func NuevaBD(bbdd string) (*BD, error) {
	db, err := sql.Open("postgres", bbdd)
	if err != nil {
		return &BD{bbdd: bbdd, bd: db}, err
	}
	execScript(db, "scripts/destruirBBDD.sql")
	execScript(db, "scripts/crearBBDD.sql")
	return &BD{bbdd: bbdd, bd: db}, err
}

// Cerrar cierra la conexion con la base de datos
func (b *BD) Cerrar() {
	b.bd.Close()
}

// CrearCuenta crea una nueva cuenta con nombre, correo, clave, recibeCorreos y
// los valores por defecto de riskos, iconos y aspectos. Devuelve todos los datos del
// nuevo usuario creado
func (b *BD) CrearCuenta(nombre, correo, clave string,
	recibeCorreos bool) mensajes.JsonData {

	// Iniciar una transaccion, solo se modifican las tablas si se modifican
	// todas
	ctx := context.Background()
	tx, err := b.bd.BeginTx(ctx, nil)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorIniciarTransaccion)
	}
	id := 0
	err = tx.QueryRowContext(ctx, crearUsuario, nombre, correo, clave,
		recibeCorreos).Scan(&id)
	if err != nil {
		tx.Rollback()
		return mensajes.ErrorJson(err.Error(), ErrorCrearCuenta)
	}
	_, err = tx.ExecContext(ctx, darIconosPorDefecto, id)
	if err != nil {
		tx.Rollback()
		return mensajes.ErrorJson(err.Error(), ErrorCrearCuenta)
	}
	_, err = tx.ExecContext(ctx, darAspectosPorDefecto, id)
	if err != nil {
		tx.Rollback()
		return mensajes.ErrorJson(err.Error(), ErrorCrearCuenta)
	}
	err = tx.Commit()
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorCrearCuenta)
	}
	// Fin de la transaccion
	return b.leerDatosUsuario(id, 1, 1, 0, nombre, correo, clave, recibeCorreos)
}

// ModificarUsuario cambia el valor del campo del usuario(id, clave) por el nuevo valor
// Si ocurre algun error devuelve el error en formato json
func (b *BD) ModificarUsuario(id int, clave, campo string, valor interface{}) mensajes.JsonData {
	res, err := b.bd.Exec(actualizarUsuarioInicio+campo+actualizarUsuarioFin, valor, id, clave)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorModificarUsuario)
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return mensajes.ErrorJson("Clave incorrecta", ErrorModificarUsuario)
	}
	return mensajes.ErrorJson("", 0)
}

// ObtenerUsuario devuelve todos los datos del usuario(id, clave) en formato json
// Si ocurre algun error devuelve el error en formato json
func (b *BD) ObtenerUsuario(id int, clave string) mensajes.JsonData {
	var aspecto, icono, riskos int
	var recibeCorreos bool
	var correo, nombre string
	err := b.bd.QueryRow(consultaUsuario, id, clave).Scan(&aspecto, &correo,
		&icono, &nombre, &recibeCorreos, &riskos)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorBusquedaUsuario)
	}
	return b.leerDatosUsuario(id, icono, aspecto, riskos, nombre, correo, clave, recibeCorreos)
}

// leerDatosUsuario devuelve el mensaje en formato json con todos los datos de un usuario.
// Si ocurre algun error devuelve el error en formato json
func (b *BD) leerDatosUsuario(id, icono, aspecto, riskos int, nombre, correo,
	clave string, recibeCorreos bool) mensajes.JsonData {

	tiendaAspectos, err := b.leerCosmetico(consultaAspectos)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorCrearCuenta)
	}
	tiendaIconos, err := b.leerCosmetico(consultaIconos)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorCrearCuenta)
	}
	aspectos, err := b.leerCosmetico(consultaAspectosUsuario + strconv.Itoa(id))
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorCrearCuenta)
	}
	iconos, err := b.leerCosmetico(consultaIconosUsuario + strconv.Itoa(id))
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorCrearCuenta)
	}
	return mensajes.JsonData{
		"usuario":        mensajes.UsuarioJson(id, icono, aspecto, riskos, nombre, correo, clave, recibeCorreos),
		"iconos":         iconos,
		"aspectos":       aspectos,
		"tiendaIconos":   tiendaIconos,
		"tiendaAspectos": tiendaAspectos,
	}
}

// leerCosmetico devuelve en formato json los cosmeticos obtenidos de consulta
// El json tiene el siguiente formato: {[ "id":int, "precio"int ]}
func (b *BD) leerCosmetico(consulta string) ([]mensajes.JsonData, error) {
	filas, err := b.bd.Query(consulta)
	if err != nil {
		return nil, err
	}
	var cosmeticos []mensajes.JsonData
	for filas.Next() {
		var id, precio int
		if err := filas.Scan(&id, &precio); err != nil {
			return nil, err
		}
		cosmeticos = append(cosmeticos, mensajes.CosmeticoJson(id, precio))
	}
	return cosmeticos, nil
}

// IniciarSesionNombre devuelve el mensaje en formato json con todos los datos
// de un usuario si la clave indicada coincide con la clave del usuario con el
// nombre indicado.
// Si ocurre algun error devuelve el error en formato json
func (b *BD) IniciarSesionNombre(nombre, clave string) mensajes.JsonData {
	var aspecto, icono, riskos, id int
	var recibeCorreos bool
	var correo string
	err := b.bd.QueryRow(consultaUsuarioNombre, nombre, clave).Scan(&aspecto, &correo,
		&icono, &id, &recibeCorreos, &riskos)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorIniciarSesion)
	}
	return b.leerDatosUsuario(id, icono, aspecto, riskos, nombre, correo, clave, recibeCorreos)
}

// IniciarSesionCorreo devuelve el mensaje en formato json con todos los datos
// de un usuario menos su clave si la clave indicada coincide con la clave del
// usuario con el correo indicado.
// Si ocurre algun error devuelve el error en formato json
func (b *BD) IniciarSesionCorreo(correo, clave string) mensajes.JsonData {
	var aspecto, icono, riskos, id int
	var recibeCorreos bool
	var nombre string
	err := b.bd.QueryRow(consultaUsuarioCorreo, correo, clave).Scan(&aspecto, &nombre,
		&icono, &id, &recibeCorreos, &riskos)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorIniciarSesion)
	}
	return b.leerDatosUsuario(id, icono, aspecto, riskos, nombre, correo, clave, recibeCorreos)
}

// execScript ejecuta el script en la base de datos db
func execScript(db *sql.DB, script string) {
	file, err := ioutil.ReadFile(script)
	if err != nil {
		fmt.Println(err)
	}
	requests := strings.Split(string(file), ";")
	for _, request := range requests {
		_, err = db.Exec(request)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (b *BD) ObtenerAmigos(id int, clave string) mensajes.JsonData {
	if err := b.comprobarClave(id, clave); err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorIniciarSesion)
	}
	filas, err := b.bd.Query(consultaAmigos, id)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorIniciarSesion)
	}
	var amigos []mensajes.JsonData
	for filas.Next() {
		var id, icono, aspecto int
		var nombre string
		if err := filas.Scan(&id, &nombre, &icono, &aspecto); err != nil {
			return mensajes.ErrorJson(err.Error(), ErrorIniciarSesion)
		}
		amigos = append(amigos, mensajes.AmigoJson(id, icono, aspecto, nombre))
	}
	return mensajes.JsonData{"amigos": amigos}
}

func (b *BD) ObtenerNotificaciones(id int, clave string) mensajes.JsonData {
	if err := b.comprobarClave(id, clave); err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorIniciarSesion)
	}
	var notificaciones []mensajes.JsonData
	n, err := b.leerNotificaciones(id, consultaSolicitudes, "Peticion de amistad")
	notificaciones = append(notificaciones, n...)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorIniciarSesion)
	}
	n, err = b.leerNotificaciones(id, consultaInvitaciones, "Invitacion")
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorIniciarSesion)
	}
	notificaciones = append(notificaciones, n...)
	n, err = b.leerNotificaciones(id, consultaTurnos, "Notificacion de turno")
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorIniciarSesion)
	}
	notificaciones = append(notificaciones, n...)
	return mensajes.JsonData{"notificaciones": notificaciones}
}

func (b *BD) EnviarSolicitudAmistad(id, amigo int, clave string) mensajes.JsonData {
	if err := b.comprobarClave(id, clave); err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorIniciarSesion)
	}
	_, err := b.bd.Exec(solicitarAmistad, id, amigo)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorIniciarSesion)
	}
	return mensajes.ErrorJson("", 0)
}

func (b *BD) leerNotificaciones(id int, consulta, tipo string) ([]mensajes.JsonData, error) {
	filas, err := b.bd.Query(consulta, id)
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
		notificaciones = append(notificaciones, mensajes.NotificacionJson(idEnvia, tipo, nombre))
	}
	return notificaciones, nil
}

func (b *BD) comprobarClave(id int, clave string) error {
	var aspecto int
	return b.bd.QueryRow(consultaUsuario).Scan(&aspecto)
}
