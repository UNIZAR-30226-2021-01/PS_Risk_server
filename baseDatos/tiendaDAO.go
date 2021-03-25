package baseDatos

import (
	"PS_Risk_server/mensajes"
	"database/sql"
	"strconv"
)

const (
	consultaAspectos = "SELECT id_aspecto AS id, precio FROM aspecto"
	consultaIconos   = "SELECT id_icono AS id, precio FROM icono"
)

type Tienda struct {
	Iconos, Aspectos []mensajes.JsonData
}

type TiendaDAO struct {
	bd *sql.DB
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
