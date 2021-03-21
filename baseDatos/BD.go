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
)

// Consultas
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
	consultaAspectosUsuario = "SELECT aspecto.id_aspecto AS id, aspecto.precio AS precio " +
		"FROM aspecto INNER JOIN aspectoscomprados ON aspecto.id_aspecto = aspectoscomprados.id_aspecto " +
		"WHERE aspectoscomprados.id_usuario = "
	consultaIconosUsuario = "SELECT icono.id_icono AS id, icono.precio AS precio " +
		"FROM icono INNER JOIN iconoscomprados ON icono.id_icono = iconoscomprados.id_icono " +
		"WHERE iconoscomprados.id_usuario = "
)

func NuevaBD(bbdd string) (*BD, error) {
	db, err := sql.Open("postgres", bbdd)
	if err != nil {
		return &BD{bbdd: bbdd, bd: db}, err
	}
	execScript(db, "scripts/destruirBBDD.sql")
	execScript(db, "scripts/crearBBDD.sql")
	return &BD{bbdd: bbdd, bd: db}, err
}

func (b *BD) CrearCuenta(nombre, correo, clave string,
	recibeCorreos bool) mensajes.JsonData {

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

	return b.leerDatosUsuario(id, 1, 1, 0, nombre, correo, clave, recibeCorreos)
}

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
