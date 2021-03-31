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

type Tienda struct {
	Iconos, Aspectos []mensajes.JsonData
}

func (t *Tienda) ObtenerPrecioIcono(id int) (int, bool) {
	for _, icono := range t.Iconos {
		if icono["id"].(int) == id {
			return icono["precio"].(int), true
		}
	}
	return 0, false
}

func (t *Tienda) ObtenerPrecioAspecto(id int) (int, bool) {
	for _, aspecto := range t.Aspectos {
		if aspecto["id"].(int) == id {
			return aspecto["precio"].(int), true
		}
	}
	return 0, false
}

type TiendaDAO struct {
	bd *sql.DB
}

func NuevaTiendaDAO(bd *sql.DB) (dao TiendaDAO) {
	return TiendaDAO{bd: bd}
}

func (dao *TiendaDAO) ObtenerTienda() (Tienda, error) {
	var t Tienda
	aspectos, err := dao.leerCosmetico(consultaAspectos)
	if err != nil {
		return t, err
	}
	iconos, err := dao.leerCosmetico(consultaIconos)
	if err != nil {
		return t, err
	}
	t = Tienda{Iconos: iconos, Aspectos: aspectos}
	return t, nil
}

func (dao *TiendaDAO) ObtenerAspectos(u Usuario) ([]mensajes.JsonData, error) {
	return dao.leerCosmetico(consultaAspectosUsuario + strconv.Itoa(u.Id))
}

func (dao *TiendaDAO) ObtenerIconos(u Usuario) ([]mensajes.JsonData, error) {
	return dao.leerCosmetico(consultaIconosUsuario + strconv.Itoa(u.Id))
}

func (dao *TiendaDAO) ComprarIcono(u *Usuario, id, precio int) mensajes.JsonData {
	return dao.comprar(u, id, precio, comprarIcono)
}

func (dao *TiendaDAO) ComprarAspecto(u *Usuario, id, precio int) mensajes.JsonData {
	return dao.comprar(u, id, precio, comprarAspecto)
}

func (dao *TiendaDAO) comprar(u *Usuario, id, precio int, sql string) mensajes.JsonData {
	if u.Riskos-precio < 0 {
		return mensajes.ErrorJson("Riskos insuficientes", 1)
	}
	// Iniciar una transaccion, solo se modifican las tablas si se modifican
	// todas
	ctx := context.Background()
	tx, err := dao.bd.BeginTx(ctx, nil)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), 1)
	}
	_, err = tx.ExecContext(ctx, sql, u.Id, id)
	if err != nil {
		tx.Rollback()
		return mensajes.ErrorJson(err.Error(), 1)
	}
	_, err = tx.ExecContext(ctx, reducirRiskos, precio, u.Id)
	if err != nil {
		tx.Rollback()
		return mensajes.ErrorJson(err.Error(), 1)
	}
	err = tx.Commit()
	if err != nil {
		return mensajes.ErrorJson(err.Error(), 1)
	}
	// Fin de la transaccion
	u.Riskos -= precio
	return mensajes.ErrorJson("", 0)
}

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
