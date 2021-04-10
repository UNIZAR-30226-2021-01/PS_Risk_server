package baseDatos

import (
	"PS_Risk_server/mensajes"
	"context"
	"database/sql"
	"strconv"
)

const (
	consultaAspectos = "SELECT id_aspecto AS id, precio FROM aspecto"
	consultaIconos   = "SELECT id_icono AS id, precio FROM icono"
	reducirRiskos    = "UPDATE usuario SET riskos = riskos - $1 WHERE id_usuario = $2"
	comprarAspecto   = "INSERT INTO aspectosComprados (id_usuario, id_aspecto) " +
		"VALUES ($1, $2)"
	comprarIcono = "INSERT INTO iconosComprados (id_usuario, id_icono) " +
		"VALUES ($1, $2)"
)

/*
	TiendaDAO permite modificar y leer las tablas de iconosComprados, aspectosComprados,
	usuario, iconos y aspectos.
*/
type TiendaDAO struct {
	bd *sql.DB
}

/*
	NuevaTiendaDAO crea un TiendaDAO.
*/
func NuevaTiendaDAO(bd *sql.DB) (dao TiendaDAO) {
	return TiendaDAO{bd: bd}
}

/*
	ObtenerTienda devuelve los datos de la tienda de la base de datos.
*/
func (dao *TiendaDAO) ObtenerTienda() (Tienda, error) {
	var t Tienda

	// Leer los aspectos de la base de datos
	aspectos, err := dao.leerCosmetico(consultaAspectos)
	if err != nil {
		return t, err
	}

	// Leer los iconos de la base de datos
	iconos, err := dao.leerCosmetico(consultaIconos)
	if err != nil {
		return t, err
	}
	t = Tienda{Iconos: iconos, Aspectos: aspectos}
	return t, nil
}

/*
	ObtenerAspectos devuelve los aspectos comprados en formato json de un usuario.
	Devuelve error en caso de no poder obtenerlos.
*/
func (dao *TiendaDAO) ObtenerAspectos(u Usuario) ([]mensajes.JsonData, error) {
	return dao.leerCosmetico(consultaAspectosUsuario + strconv.Itoa(u.Id))
}

/*
	ObtenerIconos devuelve los iconos comprados en formato json de un usuario.
	Devuelve error en caso de no poder obtenerlos.
*/
func (dao *TiendaDAO) ObtenerIconos(u Usuario) ([]mensajes.JsonData, error) {
	return dao.leerCosmetico(consultaIconosUsuario + strconv.Itoa(u.Id))
}

/*
	ComprarIcono compra un icono para un usuario.
	Si no se ha podido completar la compra, devuelve el error ocurrido en formato
	json, en caso contrario devuelve error con código NoError en el mismo formato.
*/
func (dao *TiendaDAO) ComprarIcono(u *Usuario, id, precio int) mensajes.JsonData {
	return dao.comprar(u, id, precio, comprarIcono)
}

/*
	ComprarIcono compra un aspecto para un usuario.
	Si no se ha podido completar la compra, devuelve el error ocurrido en formato
	json, en caso contrario devuelve error con código NoError en el mismo formato.
*/
func (dao *TiendaDAO) ComprarAspecto(u *Usuario, id, precio int) mensajes.JsonData {
	return dao.comprar(u, id, precio, comprarAspecto)
}

/*
	comprar compra un cosmetico para un usuario.
	Si no se ha podido completar la compra, devuelve el error ocurrido en formato
	json, en caso contrario devuelve error con código NoError en el mismo formato.
*/
func (dao *TiendaDAO) comprar(u *Usuario, id, precio int, sql string) mensajes.JsonData {
	// Comprobar si el usuario tiene dinero suficiente
	if u.Riskos-precio < 0 {
		return mensajes.ErrorJson("Riskos insuficientes", 1)
	}

	// Iniciar una transacción, solo se modifican las tablas si se modifican
	// todas
	ctx := context.Background()
	tx, err := dao.bd.BeginTx(ctx, nil)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), 1)
	}

	// Registra el cosmetico en la base de datos
	_, err = tx.ExecContext(ctx, sql, u.Id, id)
	if err != nil {
		tx.Rollback()
		return mensajes.ErrorJson(err.Error(), 1)
	}

	// Reduce el dinero del usuario en la base de datos
	_, err = tx.ExecContext(ctx, reducirRiskos, precio, u.Id)
	if err != nil {
		tx.Rollback()
		return mensajes.ErrorJson(err.Error(), 1)
	}

	// Fin de la transacción
	err = tx.Commit()
	if err != nil {
		return mensajes.ErrorJson(err.Error(), 1)
	}

	// Reducir el dinero del usuario en la variable
	u.Riskos -= precio
	return mensajes.ErrorJson("", 0)
}

/*
	leerCosmetico devuelve en formato json los elementos de una consulta que
	devuelva iconos o aspectos. Si no se pueden leer devuelve error.
*/
func (dao *TiendaDAO) leerCosmetico(consulta string) ([]mensajes.JsonData, error) {
	filas, err := dao.bd.Query(consulta)
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
