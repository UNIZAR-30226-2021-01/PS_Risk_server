package baseDatos

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"strings"

	//"encoding/json"

	_ "github.com/lib/pq"
)

type cosmetico struct {
	id     int
	precio int
}

type BD struct {
	bbdd string
	bd   *sql.DB
}

// Códigos de error posibles
const (
	NoError                 = 0
	ErrorCrearCuenta        = iota
	ErrorLeerAspectosTienda = iota
	ErrorLeerIconosTienda   = iota
)

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

func NuevaBD(bbdd string) *BD {
	//You should establish DB connection without SSL encryption, like that:
	db, err := sql.Open("postgres", bbdd)
	if err != nil {
		fmt.Println("No se ha podido abrir conexión con la base de "+
			"datos.\nError:", err)
	}
	b := &BD{
		bbdd: bbdd,
		bd:   db,
	}
	return b
}

func NuevaBDConexionLocal(bbdd string, restablecer bool) *BD {
	//You should establish DB connection without SSL encryption, like that:
	db, err := sql.Open("postgres", "user=postgres password=aseDaTos549 dbname="+
		bbdd+" sslmode=disable")
	//db, err := sql.Open("postgres", bbdd)
	if err != nil {
		fmt.Println("No se ha podido abrir conexión con la base de "+
			"datos.\nError:", err)
	}
	b := &BD{
		bbdd: bbdd,
		bd:   db,
	}

	if restablecer {
		execScript(b.bd, "../scripts/destruirBBDD.sql")
		execScript(b.bd, "../scripts/crearBBDD.sql")
	}
	return b
}

func (b *BD) CrearCuenta(nombre, correo, clave string,
	recibeCorreos bool) map[string]interface{} {
	// El aspecto e icono por defecto habría que dejarlos como constante o
	// función en algún sitio en vez de poner un 1, y lo mismo con los riskos
	// iniciales
	sentencia := "INSERT INTO usuario (aspecto, icono, nombre, correo, clave," +
		" riskos, recibeCorreos) VALUES (1, 1, $1, $2, $3, 0, $4) " +
		"RETURNING id_usuario"
	id := 0
	err := b.bd.QueryRow(sentencia, nombre, correo, clave,
		recibeCorreos).Scan(&id)
	if err != nil {
		fmt.Println("Error:", err)
	}

	resultado := make(map[string]interface{})
	errorDado := make(map[string]interface{})
	usuario := make(map[string]interface{})
	// Un usuario recién creado solo ha "comprado" los iconos y aspectos por
	// defecto. Ahora son los dos con id 1 y precio 0
	// Si se añaden más deberíamos hacer dos funciones separadas que los
	// devuelvan
	cosmeticoDefecto := cosmetico{
		id:     1,
		precio: 0,
	}
	iconos := [1]cosmetico{cosmeticoDefecto}
	aspectos := [1]cosmetico{cosmeticoDefecto}
	// Añadirlo también en la base de datos
	darIconosPorDefecto := "INSERT INTO iconosComprados (id_usuario, id_icono)" +
		" VALUES ($1, 1)"
	darAspectosPorDefecto := "INSERT INTO aspectosComprados (id_usuario, " +
		"id_aspecto) VALUES ($1, 1)"
	_, err = b.bd.Exec(darIconosPorDefecto, id)
	_, err = b.bd.Exec(darAspectosPorDefecto, id)

	codigoError := "code"
	mensajeError := "err"
	tiendaAspectos := "tiendaAspectos"
	tiendaIconos := "tiendaIconos"
	if err == nil {
		var codError int
		var textoError string
		resultado[tiendaAspectos], resultado[tiendaIconos], codError,
			textoError = b.leerTienda()
		if codError != NoError {
			resultado[codigoError] = codError
			resultado[mensajeError] = textoError
		} else {
			errorDado[codigoError] = 0
			errorDado[mensajeError] = ""
		}
	} else {
		errorDado[codigoError] = ErrorCrearCuenta
		errorDado[mensajeError] = err.Error()
		resultado[tiendaAspectos] = make(map[string]interface{})
		resultado[tiendaIconos] = make(map[string]interface{})
	}
	usuario["id"] = id
	usuario["nombre"] = nombre
	usuario["icono"] = 1
	usuario["aspecto"] = 1
	usuario["correo"] = correo
	usuario["clave"] = clave
	usuario["riskos"] = 0
	usuario["recibeCorreos"] = recibeCorreos
	resultado["error"] = errorDado
	resultado["usuario"] = usuario
	resultado["iconos"] = iconos
	resultado["aspectos"] = aspectos
	return resultado
}

func (b *BD) leerTienda() ([]cosmetico, []cosmetico, int, string) {
	consultaAspectos := "SELECT id_aspecto AS id, precio FROM aspecto"
	consultaIconos := "SELECT id_icono AS id, precio FROM icono"
	var aux cosmetico
	var aspectos []cosmetico
	var iconos []cosmetico

	filas, err := b.bd.Query(consultaAspectos)
	if err != nil {
		return make([]cosmetico, 0), make([]cosmetico, 0),
			ErrorLeerAspectosTienda, err.Error()
	}

	for filas.Next() {
		if err := filas.Scan(&aux.id, &aux.precio); err != nil {
			return make([]cosmetico, 0), make([]cosmetico, 0),
				ErrorLeerAspectosTienda, err.Error()
		}
		aspectos = append(aspectos, aux)
	}

	filas, err = b.bd.Query(consultaIconos)
	if err != nil {
		return make([]cosmetico, 0), make([]cosmetico, 0),
			ErrorLeerIconosTienda, err.Error()
	}

	for filas.Next() {
		if err := filas.Scan(&aux.id, &aux.precio); err != nil {
			return make([]cosmetico, 0), make([]cosmetico, 0),
				ErrorLeerIconosTienda, err.Error()
		}
		iconos = append(iconos, aux)
	}
	return aspectos, iconos, NoError, ""
}

func (b *BD) LeerMaxIdUsuario() int {
	var idLeido interface{}
	var id int
	err := b.bd.QueryRow("SELECT MAX(id_usuario) FROM usuario").Scan(&idLeido)
	if err != nil {
		if err == sql.ErrNoRows {
			id = 0
		} else {
			fmt.Println("Error leyendo el máximo id:", err)
		}
	} else {
		if idLeido == nil {
			id = 0
		} else {
			switch idLeidoTipo := idLeido.(type) {
			case int64: // No puede ser de otro tipo
				id = int(idLeidoTipo)
			}
		}
	}
	return id
}
