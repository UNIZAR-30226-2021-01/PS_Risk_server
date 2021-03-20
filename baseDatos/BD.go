package baseDatos

import (
	"PS_Risk_server/mensajes"
	"context"
	"database/sql"

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
)

func NuevaBD(bbdd string) (*BD, error) {
	db, err := sql.Open("postgres", bbdd)
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
	// Un usuario recién creado solo ha "comprado" los iconos y aspectos por
	// defecto. Ahora son los dos con id 1 y precio 0
	// Si se añaden más deberíamos hacer dos funciones separadas que los
	// devuelvan
	iconos := [1]mensajes.JsonData{mensajes.CosmeticoJson(1, 0)}
	aspectos := [1]mensajes.JsonData{mensajes.CosmeticoJson(1, 0)}
	// Añadirlo también en la base de datos
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
	tiendaAspectos, err := b.leerCosmetico(consultaAspectos)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorCrearCuenta)
	}
	tiendaIconos, err := b.leerCosmetico(consultaIconos)
	if err != nil {
		return mensajes.ErrorJson(err.Error(), ErrorCrearCuenta)
	}
	return mensajes.JsonData{
		"usuario":        mensajes.UsuarioJson(id, 1, 1, 0, nombre, correo, clave, recibeCorreos),
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
