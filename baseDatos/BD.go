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
	NoError                     = 0
	ErrorIniciarTransaccion     = iota
	ErrorCrearCuenta            = iota
	ErrorLeerAspectosTienda     = iota
	ErrorLeerIconosTienda       = iota
	ErrorTipoIncorrecto         = iota
	ErrorBusquedaUsuario        = iota
	ErrorModificarUsuario       = iota
	ErrorIniciarSesion          = iota
	ErrorCampoIncorrecto        = iota
	ErrorUsuarioIncorrecto      = iota
	ErrorEliminarAmigo          = iota
	ErrorAceptarAmigo           = iota
	ErrorRechazarAmigo          = iota
	ErrorNotificaciones         = iota
	ErrorLeerAmigos             = iota
	ErrorNombreUsuario          = iota
	ErrorEnviarSolicitudAmistad = iota
	ErrorAmistadDuplicada       = iota
)

// Consultas SQL
const (
	consultaUsuario = "SELECT aspecto, correo, icono, nombre, recibeCorreos, " +
		"riskos FROM usuario WHERE id_usuario = $1 AND clave = $2"
	actualizarUsuarioInicio = "UPDATE usuario SET "
	actualizarUsuarioFin    = " = $1 WHERE id_usuario = $2 AND clave = $3"
	comprobarClaveUsuario   = "SELECT id_usuario FROM usuario WHERE id_usuario = $1 AND clave = $2"
	obtenerIdUsuario        = "SELECT id_usuario FROM usuario WHERE nombre = $1"
)

func CrearBD(bbdd string) (*sql.DB, error) {
	db, err := sql.Open("postgres", bbdd)
	if err != nil {
		return db, err
	}
	execScript(db, "scripts/destruirBBDD.sql")
	execScript(db, "scripts/crearBBDD.sql")
	return db, err
}

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
	return b.leerDatosUsuario(id, 0, 0, 0, nombre, correo, clave, recibeCorreos)
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

// EliminarAmigo borra de la base de datos la relación de amistad entre los
// usuarios indicados.
// Devuelve en formato json el error ocurrido o la ausencia de errores
func (b *BD) EliminarAmigo(idUsuario, idAmigo int, clave string) mensajes.JsonData {
	var id1 int
	err := b.bd.QueryRow(comprobarClaveUsuario, idUsuario, clave).Scan(&id1)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorUsuarioIncorrecto)
	}
	// El primer usuario en la tabla es siempre el de menor id
	id1 = min(idUsuario, idAmigo)
	id2 := max(idUsuario, idAmigo)
	resultadoConsulta, err := b.bd.Exec(eliminarAmistad, id1, id2)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorEliminarAmigo)
	}
	filasEliminadas, err := resultadoConsulta.RowsAffected()
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorEliminarAmigo)
	} else if filasEliminadas == 0 {
		return mensajes.ErrorJson("Los usuarios no eran amigos", ErrorEliminarAmigo)
	}
	return mensajes.ErrorJson("", NoError)
}

// AceptarSolicitudAmistad añade en la base de datos una relación de amistad entre
// los usuarios indicados y elimina la solicitud de amistad.
// Si el usuario que acepta también le había enviado una solicitud de amistad
// al otro usuario, esta segunda solicitud también se elimina.
// Devuelve en formato json el error ocurrido o la ausencia de errores
func (b *BD) AceptarSolicitudAmistad(idUsuario, idAmigo int, clave string) mensajes.JsonData {
	id1 := min(idUsuario, idAmigo)
	id2 := max(idUsuario, idAmigo)
	var id int
	err := b.bd.QueryRow(comprobarClaveUsuario, idUsuario, clave).Scan(&id)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorUsuarioIncorrecto)
	}
	err = b.bd.QueryRow(consultaAmistad, id1, id2).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return mensajes.ErrorJson(err.Error(), ErrorAceptarAmigo)
	} else if err == nil {
		b.bd.Exec(eliminarSolicitudAmistad, idAmigo, idUsuario)
		return mensajes.ErrorJson("Los usuarios ya son amigos", ErrorAceptarAmigo)
	}

	// Iniciar una transaccion, solo se modifican las tablas si se modifican
	// todas
	ctx := context.Background()
	tx, err := b.bd.BeginTx(ctx, nil)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorIniciarTransaccion)
	}
	resultadoConsulta, err := tx.ExecContext(ctx, eliminarSolicitudAmistad, idAmigo, idUsuario)
	if err != nil {
		tx.Rollback()
		return mensajes.ErrorJson(err.Error(), ErrorAceptarAmigo)
	}
	filasEliminadas, err := resultadoConsulta.RowsAffected()
	if err != nil {
		tx.Rollback()
		return mensajes.ErrorJson(err.Error(), ErrorAceptarAmigo)
	} else if filasEliminadas == 0 {
		tx.Rollback()
		return mensajes.ErrorJson("No existe la solicitud de amistad", ErrorAceptarAmigo)
	}
	_, err = tx.ExecContext(ctx, eliminarSolicitudAmistad, idUsuario, idAmigo)
	if err != nil {
		tx.Rollback()
		return mensajes.ErrorJson(err.Error(), ErrorAceptarAmigo)
	}
	_, err = tx.ExecContext(ctx, crearAmistad, id1, id2)
	if err != nil {
		tx.Rollback()
		return mensajes.ErrorJson(err.Error(), ErrorAceptarAmigo)
	}
	err = tx.Commit()
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorAceptarAmigo)
	}
	// Fin de la transaccion
	return mensajes.ErrorJson("", NoError)
}

// RechazarSolicitudAmistad elimina la notificación de solicitud de amistad entre
// los usuarios indicados sin añadirlos como amigos en la base de datos.
// Devuelve en formato json el error ocurrido o la ausencia de errores
func (b *BD) RechazarSolicitudAmistad(idUsuario, idAmigo int, clave string) mensajes.JsonData {
	var id int
	err := b.bd.QueryRow(comprobarClaveUsuario, idUsuario, clave).Scan(&id)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorUsuarioIncorrecto)
	}
	resultadoConsulta, err := b.bd.Exec(eliminarSolicitudAmistad, idAmigo, idUsuario)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorRechazarAmigo)
	}
	filasEliminadas, err := resultadoConsulta.RowsAffected()
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorRechazarAmigo)
	} else if filasEliminadas == 0 {
		return mensajes.ErrorJson("No existe la solicitud de amistad", ErrorRechazarAmigo)
	}
	return mensajes.ErrorJson("", NoError)
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

func (b *BD) EnviarSolicitudAmistad(id int, amigo, clave string) mensajes.JsonData {
	if err := b.comprobarClave(id, clave); err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorIniciarSesion)
	}
	var idAmigo int
	err := b.bd.QueryRow(obtenerIdUsuario, amigo).Scan(&idAmigo)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorIniciarSesion)
	}
	_, err = b.bd.Exec(solicitarAmistad, id, idAmigo)
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
	var idU int
	return b.bd.QueryRow(comprobarClaveUsuario, id, clave).Scan(&idU)
}

func min(n1, n2 int) int {
	if n1 < n2 {
		return n1
	} else {
		return n2
	}
}

func max(n1, n2 int) int {
	if n1 < n2 {
		return n2
	} else {
		return n1
	}
}
